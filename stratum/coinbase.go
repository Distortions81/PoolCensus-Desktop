package stratum

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func DecodeCoinbaseParts(coinbase1, coinbase2, extraNonce1 string, extraNonce2Size int) (*CoinbaseInfo, error) {
	extraNonce2Bytes := make([]byte, extraNonce2Size)
	if extraNonce2Size > 0 {
		if _, err := rand.Read(extraNonce2Bytes); err != nil {
			return nil, err
		}
	}
	extraNonce2 := hex.EncodeToString(extraNonce2Bytes)
	fullHex, err := BuildFullCoinbase(coinbase1, extraNonce1, extraNonce2, coinbase2)
	if err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(fullHex)
	if err != nil {
		return nil, err
	}

	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
		return nil, err
	}

	scriptSig := []byte(nil)
	if len(tx.TxIn) > 0 {
		scriptSig = tx.TxIn[0].SignatureScript
	}

	info := &CoinbaseInfo{
		ExtraNonce2: extraNonce2,
		ScriptSig:   scriptSig,
	}

	for _, out := range tx.TxOut {
		scriptType, address := classifyOutput(out.PkScript)
		info.Outputs = append(info.Outputs, CoinbaseOutput{
			Address:    address,
			ValueBTC:   float64(out.Value) / 1e8,
			ScriptType: scriptType,
		})
	}

	return info, nil
}

func BuildFullCoinbase(coinbase1, extraNonce1, extraNonce2, coinbase2 string) (string, error) {
	if _, err := hex.DecodeString(coinbase1); err != nil {
		return "", fmt.Errorf("coinbase1: %w", err)
	}
	if _, err := hex.DecodeString(coinbase2); err != nil {
		return "", fmt.Errorf("coinbase2: %w", err)
	}
	if _, err := hex.DecodeString(extraNonce1); err != nil {
		return "", fmt.Errorf("extranonce1: %w", err)
	}
	if _, err := hex.DecodeString(extraNonce2); err != nil {
		return "", fmt.Errorf("extranonce2: %w", err)
	}
	return coinbase1 + extraNonce1 + extraNonce2 + coinbase2, nil
}

func classifyOutput(pkScript []byte) (scriptType, address string) {
	if len(pkScript) >= 38 &&
		pkScript[0] == txscript.OP_RETURN &&
		pkScript[1] == 0x24 &&
		pkScript[2] == 0xaa &&
		pkScript[3] == 0x21 &&
		pkScript[4] == 0xa9 &&
		pkScript[5] == 0xed {
		return "witness_commitment", ""
	}

	class := txscript.GetScriptClass(pkScript)
	if class == txscript.NullDataTy {
		return "OP_RETURN", ""
	}

	_, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, &chaincfg.MainNetParams)
	if err == nil && len(addrs) > 0 {
		return class.String(), addrs[0].EncodeAddress()
	}
	return class.String(), ""
}
