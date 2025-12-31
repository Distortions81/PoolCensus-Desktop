package main

import "time"

type logEntry struct {
	Timestamp       string        `json:"timestamp"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Connected       bool          `json:"connected"`
	Error           string        `json:"error"`
	UserAgent       string        `json:"user_agent"`
	Username        string        `json:"username"`
	TotalPayout     float64       `json:"total_payout"`
	PingMs          float64       `json:"ping_ms"`
	JobLatencyMs    float64       `json:"job_latency_ms,omitempty"`
	WalletAddress   string        `json:"wallet_address"`
	WorkerName      string        `json:"worker_name"`
	Password        string        `json:"password"`
	ExtraNonce1     string        `json:"extranonce1"`
	ExtraNonce2Size int           `json:"extranonce2_size"`
	Difficulty      float64       `json:"difficulty"`
	BlockHeight     uint32        `json:"block_height"`
	PoolTag         string        `json:"pool_tag"`
	TLS             bool          `json:"tls"`
	CoinbaseRaw     *coinbaseData `json:"coinbase_raw"`
	Payouts         []payout      `json:"payouts"`
}

type coinbaseData struct {
	CoinBase1 string `json:"coinbase1"`
	CoinBase2 string `json:"coinbase2"`
	FullHex   string `json:"full_hex"`
}

type payout struct {
	OutputIndex int     `json:"output_index"`
	Address     string  `json:"address"`
	Amount      float64 `json:"amount_btc"`
	Type        string  `json:"type"`
}

type issueDetail struct {
	Message     string
	Explanation string
	Score       int
}

type payoutView struct {
	payout
	IsWorker bool
	Percent  float64
}

type pingSummary struct {
	Exists   bool
	Min      string
	Avg      string
	Max      string
	Samples  int
	AvgValue float64
	Jitter   string
	RealPing string
}

type jobSummary struct {
	Exists   bool
	Min      string
	Avg      string
	Max      string
	Jitter   string
	Real     string
	AvgValue float64
}

type entryView struct {
	Timestamp          string
	TimestampRaw       time.Time
	LogFile            string
	PoolName           string
	Host               string
	Port               int
	PortDisplay        string
	Ping               string
	PingSummaryPrimary pingSummary
	PingSummaryTLS     pingSummary
	PingSort           float64
	TotalPayout        float64
	WorkerShare        float64
	WorkerPercent      float64
	PoolWallet         string
	PoolWalletDisp     string
	PoolWalletURL      string
	HasPoolWallet      bool
	TLS                bool
	ShowTLSPanel       bool
	Issues             []issueDetail
	IssueSeverity      int
	Changes            []changeDetail
	HiddenChanges      int
	DisplayPayouts     []payoutView
	SplitCount         int
	HasData            bool
	Connected          bool
	Error              string
	PanelClass         string
	RewardNote         string
	RewardClass        string
	PingClass          string
	ScanURL            string
	HistoryURL         string
	LatestChanges      []changeDetail
	JobLatency         string
	JobLatencyClass    string
	JobWaitSummary     jobSummary
	JobWaitSort        float64
}

type hostView struct {
	Host   string
	Latest *entryView
}

type hostEntry struct {
	PoolName string
	Host     *hostView
	LogFile  string
}

type pingStats struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int
}

func (p *pingStats) Add(ms float64) {
	p.AddBounded(ms, timeoutPingMs)
}

func (p *pingStats) AddBounded(ms, max float64) {
	if ms <= 0 || (max > 0 && ms >= max) {
		return
	}
	if p.Count == 0 {
		p.Min = ms
		p.Max = ms
	} else {
		if ms < p.Min {
			p.Min = ms
		}
		if ms > p.Max {
			p.Max = ms
		}
	}
	p.Sum += ms
	p.Count++
}

func (p pingStats) Avg() float64 {
	if p.Count == 0 {
		return 0
	}
	return p.Sum / float64(p.Count)
}

type changeDetail struct {
	At      string
	AtShort string
	Field   string
	From    string
	To      string
	ScanURL string
}

type dashboardView struct {
	CleanEntries []*hostEntry
	IssueEntries []*hostEntry
	SortBy       string
	HostFilter   string
}

const defaultBaseReward = 3.125
