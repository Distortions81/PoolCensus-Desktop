package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const timeoutPingMs = 9999.0
const timeoutJobWaitMs = 30000.0

func dominantPoolWallet(entry *logEntry) (string, bool) {
	if entry == nil || len(entry.Payouts) == 0 {
		return "", false
	}
	var bestAddr string
	bestAmount := -1.0
	for _, payout := range entry.Payouts {
		addr := strings.TrimSpace(payout.Address)
		if addr == "" {
			continue
		}
		if entry.WalletAddress != "" && addr == entry.WalletAddress {
			continue
		}
		if payout.Amount > bestAmount {
			bestAddr = addr
			bestAmount = payout.Amount
		}
	}
	if bestAddr == "" {
		return "", false
	}
	return bestAddr, true
}

func payoutSummary(entry *logEntry) string {
	if entry == nil || len(entry.Payouts) == 0 {
		return "no outputs"
	}
	const maxOutputs = 3
	lines := make([]string, 0, maxOutputs+2)
	lines = append(lines, fmt.Sprintf("%d output(s):", len(entry.Payouts)))
	for i, p := range entry.Payouts {
		if i >= maxOutputs {
			break
		}
		addr := normalizePayoutAddress(entry, p.Address)
		display := addr
		if addr != "<worker>" && addr != "<none>" {
			display = summarizeString(addr, 18)
		}
		typ := p.Type
		if typ == "" {
			typ = "unknown"
		}
		percent := "n/a"
		if entry.TotalPayout > 0 {
			percent = formatTrimmedFloat((p.Amount/entry.TotalPayout)*100, 2) + "%"
		}
		lines = append(lines, fmt.Sprintf("#%d %s (%s) %s (%s)", p.OutputIndex, formatTrimmedFloat(p.Amount, 8), percent, display, typ))
	}
	if len(entry.Payouts) > maxOutputs {
		lines = append(lines, fmt.Sprintf("+%d more", len(entry.Payouts)-maxOutputs))
	}
	return strings.Join(lines, "\n")
}

func normalizePayoutAddress(entry *logEntry, addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "<none>"
	}
	if entry != nil && entry.WalletAddress != "" && addr == entry.WalletAddress {
		return "<worker>"
	}
	return addr
}

func isEmptyish(s string) bool {
	switch strings.TrimSpace(s) {
	case "", "<none>", "n/a":
		return true
	case "0", "0.0", "0.00", "0.00000000":
		return true
	default:
		return false
	}
}

func summarizeString(s string, max int) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "<none>"
	}
	if max <= 0 || len(trimmed) <= max {
		return trimmed
	}
	if max <= 3 {
		return trimmed[:max]
	}
	return trimmed[:max-3] + "..."
}

func formatTrimmedFloat(f float64, decimals int) string {
	if f == 0 {
		return "0"
	}
	if decimals < 0 {
		decimals = 0
	}
	if decimals > 18 {
		decimals = 18
	}
	s := fmt.Sprintf("%.*f", decimals, f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "-0" {
		return "0"
	}
	return s
}

func extractPoolTag(coinbase2Hex string) string {
	if coinbase2Hex == "" {
		return ""
	}
	raw, err := hex.DecodeString(coinbase2Hex)
	if err != nil || len(raw) == 0 {
		return ""
	}

	for i := 0; i < len(raw); i++ {
		if raw[i] != '/' {
			continue
		}
		for j := i + 1; j < len(raw); j++ {
			b := raw[j]
			if b == '/' {
				candidate := raw[i : j+1]
				if len(candidate) < 3 || len(candidate) > 96 {
					break
				}
				ok := true
				for _, c := range candidate {
					if c < 32 || c > 126 {
						ok = false
						break
					}
				}
				if ok {
					return string(candidate)
				}
				break
			}
			if b < 32 || b > 126 {
				break
			}
		}
	}

	return ""
}

func shortTimestamp(ts string) string {
	parsed, err := time.Parse(time.RFC3339, ts)
	if err == nil {
		return parsed.Local().Format("01-02 15:04:05")
	}
	return summarizeString(ts, 19)
}
