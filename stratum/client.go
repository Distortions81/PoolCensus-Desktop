package stratum

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool

	mu             sync.RWMutex
	conn           net.Conn
	reader         *bufio.Reader
	nextID         int
	pending        map[int]chan rpcResponse
	closeOnce      sync.Once
	closed         chan struct{}
	extraNonce1    string
	extraNonce2Len int

	OnNotify      func(params *NotifyParams)
	OnDifficulty  func(diff float64)
	OnDisconnect  func(err error)
	readLoopReady chan struct{}
}

func NewClient(host string, port int, username, password string, useTLS bool) *Client {
	return &Client{
		host:          host,
		port:          port,
		username:      username,
		password:      password,
		useTLS:        useTLS,
		pending:       make(map[int]chan rpcResponse),
		closed:        make(chan struct{}),
		readLoopReady: make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}

	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	dialer := &net.Dialer{Timeout: 12 * time.Second, KeepAlive: 30 * time.Second}
	var conn net.Conn
	var err error
	if c.useTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: c.host})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}

	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, 256*1024)

	go c.readLoop()
	<-c.readLoopReady
	return nil
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		for id, ch := range c.pending {
			close(ch)
			delete(c.pending, id)
		}
	})
}

func (c *Client) ExtraNonce1() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.extraNonce1
}

func (c *Client) ExtraNonce2Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.extraNonce2Len
}

func (c *Client) Subscribe(agent string) error {
	var result []any
	if err := c.call("mining.subscribe", []any{agent}, &result); err != nil {
		return err
	}
	if len(result) < 3 {
		return fmt.Errorf("mining.subscribe: unexpected result")
	}
	en1, ok := result[1].(string)
	if !ok || en1 == "" {
		return fmt.Errorf("mining.subscribe: missing extranonce1")
	}
	en2sizeFloat, ok := result[2].(float64)
	if !ok {
		return fmt.Errorf("mining.subscribe: missing extranonce2_size")
	}

	c.mu.Lock()
	c.extraNonce1 = en1
	c.extraNonce2Len = int(en2sizeFloat)
	c.mu.Unlock()
	return nil
}

func (c *Client) Authorize() error {
	var ok bool
	if err := c.call("mining.authorize", []any{c.username, c.password}, &ok); err != nil {
		return err
	}
	if !ok {
		return errors.New("authorization rejected")
	}
	return nil
}

type rpcRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []any         `json:"params"`
}

type rpcEnvelope struct {
	ID     *int            `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  any             `json:"error"`
}

type rpcResponse struct {
	Result json.RawMessage
	Error  any
}

func (c *Client) call(method string, params []any, out any) error {
	resp, err := c.send(method, params)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("%s: %v", method, resp.Error)
	}
	if out == nil {
		return nil
	}
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return fmt.Errorf("%s: empty result", method)
	}
	if err := json.Unmarshal(resp.Result, out); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}

func (c *Client) send(method string, params []any) (rpcResponse, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return rpcResponse{}, errors.New("not connected")
	}
	c.nextID++
	id := c.nextID
	respCh := make(chan rpcResponse, 1)
	c.pending[id] = respCh
	conn := c.conn
	c.mu.Unlock()

	req := rpcRequest{ID: id, Method: method, Params: params}
	payload, err := json.Marshal(req)
	if err != nil {
		return rpcResponse{}, err
	}
	payload = append(payload, '\n')
	if _, err := conn.Write(payload); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return rpcResponse{}, err
	}

	select {
	case resp, ok := <-respCh:
		if !ok {
			return rpcResponse{}, errors.New("connection closed")
		}
		return resp, nil
	case <-c.closed:
		return rpcResponse{}, errors.New("connection closed")
	case <-time.After(20 * time.Second):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return rpcResponse{}, fmt.Errorf("%s: timeout", method)
	}
}

func (c *Client) readLoop() {
	defer func() {
		c.Close()
	}()
	close(c.readLoopReady)

	for {
		select {
		case <-c.closed:
			return
		default:
		}

		line, err := readLine(c.reader)
		if err != nil {
			if c.OnDisconnect != nil {
				c.OnDisconnect(err)
			}
			return
		}

		var env rpcEnvelope
		if err := json.Unmarshal(line, &env); err != nil {
			continue
		}

		if env.ID != nil {
			c.mu.Lock()
			respCh := c.pending[*env.ID]
			delete(c.pending, *env.ID)
			c.mu.Unlock()
			if respCh != nil {
				respCh <- rpcResponse{Result: env.Result, Error: env.Error}
				close(respCh)
			}
			continue
		}

		switch env.Method {
		case "mining.notify":
			params, err := decodeNotifyParams(env.Params)
			if err != nil {
				continue
			}
			if c.OnNotify != nil {
				c.OnNotify(params)
			}
		case "mining.set_extranonce":
			en1, en2, err := decodeExtranonce(env.Params)
			if err != nil {
				continue
			}
			c.mu.Lock()
			if en1 != "" {
				c.extraNonce1 = en1
			}
			if en2 > 0 {
				c.extraNonce2Len = en2
			}
			c.mu.Unlock()
		case "mining.set_difficulty":
			diff := decodeDifficulty(env.Params)
			if c.OnDifficulty != nil && diff > 0 {
				c.OnDifficulty(diff)
			}
		}
	}
}

func readLine(r *bufio.Reader) ([]byte, error) {
	var out []byte
	for {
		chunk, err := r.ReadBytes('\n')
		out = append(out, chunk...)
		if err == nil {
			return out, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return nil, err
	}
}

func decodeNotifyParams(raw json.RawMessage) (*NotifyParams, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	if len(arr) < 4 {
		return nil, errors.New("notify: too few params")
	}
	var coinbase1, coinbase2 string
	if err := json.Unmarshal(arr[2], &coinbase1); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr[3], &coinbase2); err != nil {
		return nil, err
	}
	return &NotifyParams{CoinBase1: coinbase1, CoinBase2: coinbase2}, nil
}

func decodeDifficulty(raw json.RawMessage) float64 {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) < 1 {
		return 0
	}
	var diff float64
	_ = json.Unmarshal(arr[0], &diff)
	return diff
}

func decodeExtranonce(raw json.RawMessage) (string, int, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return "", 0, err
	}
	if len(arr) < 2 {
		return "", 0, errors.New("set_extranonce: too few params")
	}
	var en1 string
	if err := json.Unmarshal(arr[0], &en1); err != nil {
		return "", 0, err
	}
	var en2sizeFloat float64
	if err := json.Unmarshal(arr[1], &en2sizeFloat); err != nil {
		return "", 0, err
	}
	return en1, int(en2sizeFloat), nil
}
