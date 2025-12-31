package main

import (
	"fmt"
	"math/rand"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

var agentStrings = []string{
	"cgminer/4.10.0",
	"bfgminer/5.5.0",
	"phoenixminer/5.6d",
	"teamredminer/0.10.8",
	"lolminer/1.58",
	"trex/0.24.4",
	"nbminer/40.0",
	"vellamo/1.0",
}

func loadRandomAgent() string {
	return agentStrings[rand.Intn(len(agentStrings))]
}

func generateWorkerName() string {
	patterns := []string{
		"worker-%d",
		"rig-%d",
		"miner-%d",
		"node-%d",
		"farm-%d",
		"s19",
		"t19",
		"avalonminer",
		"whatsminer",
		"worker",
		"mining",
	}
	pattern := patterns[rand.Intn(len(patterns))]
	for i := 0; i < len(pattern)-1; i++ {
		if pattern[i] == '%' && pattern[i+1] == 'd' {
			return fmt.Sprintf(pattern, rand.Intn(100))
		}
	}
	return pattern
}

func generateRandomWallet() (string, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}
	pubKey := privateKey.PubKey()
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	address, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("failed to create address: %w", err)
	}
	return address.EncodeAddress(), nil
}
