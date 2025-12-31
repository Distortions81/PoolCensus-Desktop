package stratum

type NotifyParams struct {
	CoinBase1 string
	CoinBase2 string
}

type CoinbaseOutput struct {
	Address    string
	ValueBTC   float64
	ScriptType string
}

type CoinbaseInfo struct {
	ExtraNonce2 string
	ScriptSig   []byte
	Outputs     []CoinbaseOutput
}

