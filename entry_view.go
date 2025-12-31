package main

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"
)

const (
	severitySplitMissingWallet  = 100
	severitySingleMissingWallet = 80
	severityNoPayout            = 120
	severityLowShare            = 50
)

func buildEntryView(entry *logEntry, baseReward float64, tlsPing, plainPing, jobLatency pingStats, changes []changeDetail, hiddenChanges int, plainPort, tlsPort int) *entryView {
	poolWallet, ok := dominantPoolWallet(entry)
	tlsSummary := summarizePing(tlsPing)
	plainSummary := summarizePing(plainPing)
	primarySummary := plainSummary
	showTLSPanel := tlsSummary.Exists
	if !plainSummary.Exists {
		primarySummary = tlsSummary
	}
	jobSummary := summarizeJobLatency(jobLatency)
	jobLatencyVal := jobSummary.AvgValue
	jobLatencyClass := ""
	if jobLatencyVal <= 0 {
		jobLatencyClass = "bad"
	}
	view := &entryView{
		Host:               entry.Host,
		Port:               entry.Port,
		PortDisplay:        computePortDisplay(entry.Port, plainPort, tlsPort),
		Ping:               formatPing(entry.PingMs),
		JobLatency:         formatJobLatency(jobLatencyVal),
		JobLatencyClass:    jobLatencyClass,
		JobWaitSummary:     jobSummary,
		Timestamp:          entry.Timestamp,
		TotalPayout:        entry.TotalPayout,
		WorkerShare:        0,
		Connected:          entry.Connected,
		SplitCount:         len(entry.Payouts),
		HasData:            len(entry.Payouts) > 0 && entry.TotalPayout > 0,
		Changes:            changes,
		HiddenChanges:      hiddenChanges,
		PoolWallet:         poolWallet,
		PoolWalletDisp:     walletDisplay(poolWallet),
		PoolWalletURL:      blockchainAddressURL(poolWallet),
		HasPoolWallet:      ok,
		TLS:                entry.TLS,
		PingSummaryPrimary: primarySummary,
		PingSummaryTLS:     tlsSummary,
		ShowTLSPanel:       showTLSPanel,
	}
	if primarySummary.Exists {
		view.PingSort = primarySummary.AvgValue
		view.Ping = primarySummary.RealPing
	} else {
		view.PingSort = entry.PingMs
	}
	view.JobWaitSort = jobSummary.AvgValue

	if entry.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
			view.TimestampRaw = ts
		}
	}

	workerShare := 0.0
	for _, payout := range entry.Payouts {
		if payout.Address != "" && entry.WalletAddress != "" && payout.Address == entry.WalletAddress {
			workerShare += payout.Amount
		}
	}
	view.WorkerShare = workerShare
	if entry.TotalPayout > 0 {
		view.WorkerPercent = (workerShare / entry.TotalPayout) * 100
	}

	for _, payout := range entry.Payouts {
		percent := 0.0
		if entry.TotalPayout > 0 {
			percent = (payout.Amount / entry.TotalPayout) * 100
		}
		view.DisplayPayouts = append(view.DisplayPayouts, payoutView{
			payout:   payout,
			IsWorker: payout.Address != "" && entry.WalletAddress != "" && payout.Address == entry.WalletAddress,
			Percent:  percent,
		})
	}

	view.Issues, view.IssueSeverity = collectIssues(entry, workerShare, view.WorkerPercent)
	view.RewardNote, view.RewardClass = rewardNoteAndClass(entry.TotalPayout, baseReward)
	view.PanelClass = panelClass(view)
	view.PingClass = pingClass(entry.PingMs)

	return view
}

func summarizePing(stats pingStats) pingSummary {
	if stats.Count == 0 {
		return pingSummary{}
	}
	jitterValue := stats.Max - stats.Min
	if jitterValue < 0 {
		jitterValue = 0
	}
	avg := stats.Avg()
	totalAvg := avg + jitterValue/2
	return pingSummary{
		Exists:   true,
		Min:      formatPing(stats.Min),
		Avg:      formatPing(avg),
		Max:      formatPing(stats.Max),
		Samples:  stats.Count,
		AvgValue: totalAvg,
		Jitter:   formatPing(jitterValue),
		RealPing: formatPing(totalAvg),
	}
}

func summarizeJobLatency(stats pingStats) jobSummary {
	if stats.Count == 0 {
		return jobSummary{}
	}
	jitterValue := stats.Max - stats.Min
	if jitterValue < 0 {
		jitterValue = 0
	}
	avg := stats.Avg()
	real := avg + jitterValue/2
	return jobSummary{
		Exists:   true,
		Min:      formatPing(stats.Min),
		Avg:      formatPing(avg),
		Max:      formatPing(stats.Max),
		Jitter:   formatPing(jitterValue),
		Real:     formatPing(real),
		AvgValue: real,
	}
}

func collectIssues(entry *logEntry, workerShare, workerPercent float64) ([]issueDetail, int) {
	var issues []issueDetail
	severity := 0

	workerWalletPresent := entry.WalletAddress != "" && workerShare > 0

	if len(entry.Payouts) > 1 && !workerWalletPresent {
		msg := fmt.Sprintf("split coinbase (%d outputs) missing worker wallet", len(entry.Payouts))
		issues = append(issues, issueDetail{
			Message:     msg,
			Explanation: "Split coinbases without the worker wallet often indicate the pool is keeping the reward before paying the worker later.",
			Score:       severitySplitMissingWallet,
		})
		if severitySplitMissingWallet > severity {
			severity = severitySplitMissingWallet
		}
	}

	if len(entry.Payouts) == 1 && !workerWalletPresent && entry.WalletAddress != "" {
		msg := "SINGLE PAYOUT MISSING WORKER WALLET"
		issues = append(issues, issueDetail{
			Message:     msg,
			Explanation: "Single payout to a pool wallet means the worker is not paid directly—only the pool promises to distribute earnings later.",
			Score:       severitySingleMissingWallet,
		})
		if severitySingleMissingWallet > severity {
			severity = severitySingleMissingWallet
		}
	}

	if workerWalletPresent && workerPercent+0.001 < 98.0 {
		msg := fmt.Sprintf("worker share %s%% below 98%%", formatTrimmedFloat(workerPercent, 2))
		issues = append(issues, issueDetail{
			Message:     msg,
			Explanation: "Worker receives part of the reward but below 98%, which is unusually low.",
			Score:       severityLowShare,
		})
		if severityLowShare > severity {
			severity = severityLowShare
		}
	}

	if len(entry.Payouts) == 0 || entry.TotalPayout <= 0 {
		if severityNoPayout > severity {
			severity = severityNoPayout
		}
	}

	return issues, severity
}

func rewardNoteAndClass(total, baseReward float64) (string, string) {
	if total <= 0.00000001 {
		return "Payout not recorded yet", "reward-red"
	}
	if total > baseReward {
		return "Total payout amount correct", "reward-blue"
	}
	return fmt.Sprintf("Total payout less than block reward (%s BTC)", formatTrimmedFloat(total, 8)), "reward-red"
}

func panelClass(entry *entryView) string {
	if entry == nil {
		return ""
	}
	if len(entry.Issues) == 0 {
		return "panel-good"
	}
	return "panel-bad"
}

func formatPing(ping float64) string {
	switch {
	case ping <= 0:
		return "n/a"
	case ping >= timeoutPingMs:
		return "timeout"
	case ping < 1:
		return "<1 ms"
	default:
		return fmt.Sprintf("%.0f ms", math.Round(ping))
	}
}

func formatJobLatency(latency float64) string {
	if latency <= 0 {
		return "<timeout>"
	}
	return formatPing(latency)
}

func computePortDisplay(entryPort, plainPort, tlsPort int) string {
	if plainPort > 0 && tlsPort > 0 && plainPort != tlsPort {
		return fmt.Sprintf(":%d/:%d", plainPort, tlsPort)
	}
	if entryPort > 0 {
		return fmt.Sprintf(":%d", entryPort)
	}
	return ""
}

func pingClass(ping float64) string {
	switch {
	case ping >= timeoutPingMs:
		return "ping-red"
	case ping > 150:
		return "ping-red"
	case ping > 100:
		return "ping-orange"
	case ping < 50 && ping > 0:
		return "ping-blue"
	case ping < 60 && ping > 0:
		return "ping-green"
	default:
		return ""
	}
}

func scanURL(_ string, _, _ string) string {
	return "#"
}

func historyURL(_ string, _ string) string {
	return "#"
}

func walletDisplay(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	const keepPrefix = 10
	const keepSuffix = 8
	if len(addr) <= keepPrefix+keepSuffix+3 {
		return addr
	}
	return addr[:keepPrefix] + "..." + addr[len(addr)-keepSuffix:]
}

func blockchainAddressURL(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	return "https://www.blockchain.com/explorer/addresses/btc/" + url.PathEscape(addr)
}
