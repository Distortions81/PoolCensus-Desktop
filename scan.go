package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"poolcensus/desktop/stratum"
)

type scanTarget struct {
	PoolName string
	Host     string
	Port     int
	TLS      bool
}

type scanAggregate struct {
	target     scanTarget
	latest     *logEntry
	pingStats  pingStats
	jobStats   pingStats
	attempts   int
	errorCount int
	errors     []error
}

type scanOptions struct {
	Passes     int
	Verbose    bool
	OnProgress func(current, total int, target scanTarget)
}

func collectTargets(pools *PoolsData, filter string) []scanTarget {
	var targets []scanTarget
	for _, pool := range filterPools(pools, filter) {
		for _, ep := range pool.Endpoints {
			if ep.Host == "" || ep.Port == 0 {
				continue
			}
			targets = append(targets, scanTarget{
				PoolName: pool.Name,
				Host:     ep.Host,
				Port:     ep.Port,
				TLS:      ep.TLS,
			})
		}
	}
	return targets
}

func scanTargets(targets []scanTarget, agent, username, wallet, worker string, opts scanOptions) []*scanAggregate {
	if len(targets) == 0 || opts.Passes <= 0 {
		if len(targets) == 0 {
			return nil
		}
		opts.Passes = defaultScanPasses
	}

	results := make(map[string]*scanAggregate)
	total := len(targets) * opts.Passes
	progress := 0

	for pass := 0; pass < opts.Passes; pass++ {
		for _, target := range targets {
			entry, err := collectFromPool(target, agent, username, wallet, worker, opts.Verbose)
			progress++
			if opts.OnProgress != nil {
				opts.OnProgress(progress, total, target)
			}

			key := fmt.Sprintf("%s:%d", target.Host, target.Port)
			agg, ok := results[key]
			if !ok {
				agg = &scanAggregate{target: target}
				results[key] = agg
			}
			agg.attempts++

			if entry != nil {
				agg.latest = entry
				agg.pingStats.Add(entry.PingMs)
				agg.jobStats.AddBounded(entry.JobLatencyMs, timeoutJobWaitMs)
				if !entry.Connected {
					agg.errorCount++
					agg.errors = append(agg.errors, fmt.Errorf("%s:%d: %s", target.Host, target.Port, entry.Error))
				}
			}
			logVerbose(opts.Verbose, "pool %s:%d: %v", target.Host, target.Port, err)
		}
	}

	aggregates := make([]*scanAggregate, 0, len(results))
	for _, agg := range results {
		if agg.latest != nil {
			aggregates = append(aggregates, agg)
		}
	}
	sort.Slice(aggregates, func(i, j int) bool {
		if aggregates[i].target.Host == aggregates[j].target.Host {
			return aggregates[i].target.Port < aggregates[j].target.Port
		}
		return aggregates[i].target.Host < aggregates[j].target.Host
	})
	return aggregates
}

func collectFromPool(target scanTarget, agent, username, wallet, worker string, verbose bool) (*logEntry, error) {
	client := stratum.NewClient(target.Host, target.Port, username, "x", target.TLS)
	defer client.Close()

	var (
		done         = make(chan struct{}, 1)
		disconnect   = make(chan error, 1)
		jobLatency   float64
		pingMs       float64
		currentDiff  float64
		jobWaitStart time.Time
		captured     *logEntry
	)

	client.OnDisconnect = func(err error) {
		select {
		case disconnect <- err:
		default:
		}
	}

	client.OnNotify = func(params *stratum.NotifyParams) {
		if captured != nil {
			return
		}
		if !jobWaitStart.IsZero() {
			jobLatency = time.Since(jobWaitStart).Seconds() * 1000.0
		}
		info, err := stratum.DecodeCoinbaseParts(
			params.CoinBase1,
			params.CoinBase2,
			client.ExtraNonce1(),
			client.ExtraNonce2Size(),
		)
		if err != nil {
			logVerbose(verbose, "failed to decode coinbase for %s:%d: %v", target.Host, target.Port, err)
		}

		captured = buildJobEntry(target, params, client, agent, username, wallet, worker, currentDiff, pingMs, jobLatency, info)
		select {
		case done <- struct{}{}:
		default:
		}
	}

	client.OnDifficulty = func(diff float64) {
		currentDiff = diff
	}

	if err := client.Connect(); err != nil {
		return buildErrorEntry(target, agent, username, wallet, worker, err), err
	}

	start := time.Now()
	if err := client.Subscribe(agent); err != nil {
		return buildErrorEntry(target, agent, username, wallet, worker, err), err
	}
	pingMs = time.Since(start).Seconds() * 1000.0

	jobWaitStart = time.Now()
	if err := client.Authorize(); err != nil {
		return buildErrorEntry(target, agent, username, wallet, worker, err), err
	}

	select {
	case <-done:
		return captured, nil
	case err := <-disconnect:
		if err == nil {
			err = errors.New("connection closed")
		}
		return buildErrorEntry(target, agent, username, wallet, worker, err), err
	case <-time.After(30 * time.Second):
		err := fmt.Errorf("timeout waiting for job")
		return buildErrorEntryWithConnected(target, agent, username, wallet, worker, err, true), err
	}
}

func buildErrorEntry(target scanTarget, agent, username, wallet, worker string, err error) *logEntry {
	return buildErrorEntryWithConnected(target, agent, username, wallet, worker, err, false)
}

func buildErrorEntryWithConnected(target scanTarget, agent, username, wallet, worker string, err error, connected bool) *logEntry {
	return &logEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Host:          target.Host,
		Port:          target.Port,
		Connected:     connected,
		Error:         err.Error(),
		UserAgent:     agent,
		Username:      username,
		WalletAddress: wallet,
		WorkerName:    worker,
		Password:      "x",
		TLS:           target.TLS,
	}
}

func buildJobEntry(target scanTarget, params *stratum.NotifyParams, client *stratum.Client, agent, username, wallet, worker string, difficulty, pingMs, jobLatency float64, info *stratum.CoinbaseInfo) *logEntry {
	var totalPayout float64
	var payoutList []payout
	if info != nil {
		for i, output := range info.Outputs {
			if output.ScriptType == "witness_commitment" || output.ScriptType == "OP_RETURN" {
				continue
			}
			payoutList = append(payoutList, payout{
				OutputIndex: i,
				Address:     output.Address,
				Amount:      output.ValueBTC,
				Type:        output.ScriptType,
			})
			totalPayout += output.ValueBTC
		}
	}

	blockHeight := uint32(0)
	if info != nil && len(info.ScriptSig) > 0 {
		if blockHeightVal, ok := stratum.ParseScriptSig(info.ScriptSig)["block_height"].(uint32); ok {
			blockHeight = blockHeightVal
		}
	}

	fullCoinbase := ""
	if info != nil {
		fullCoinbase, _ = stratum.BuildFullCoinbase(
			params.CoinBase1,
			client.ExtraNonce1(),
			info.ExtraNonce2,
			params.CoinBase2,
		)
	}

	poolTag := target.PoolName
	if poolTag == "" {
		poolTag = extractPoolTag(params.CoinBase2)
	}

	return &logEntry{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Host:            target.Host,
		Port:            target.Port,
		Connected:       true,
		UserAgent:       agent,
		Username:        username,
		WalletAddress:   wallet,
		WorkerName:      worker,
		Password:        "x",
		ExtraNonce1:     client.ExtraNonce1(),
		ExtraNonce2Size: client.ExtraNonce2Size(),
		Difficulty:      difficulty,
		PingMs:          pingMs,
		JobLatencyMs:    jobLatency,
		BlockHeight:     blockHeight,
		PoolTag:         poolTag,
		TLS:             target.TLS,
		CoinbaseRaw: &coinbaseData{
			CoinBase1: params.CoinBase1,
			CoinBase2: params.CoinBase2,
			FullHex:   fullCoinbase,
		},
		Payouts:     payoutList,
		TotalPayout: totalPayout,
	}
}

func logVerbose(enabled bool, format string, args ...interface{}) {
	if enabled {
		log.Printf(format, args...)
	}
}
