package main

import (
	"math"
	"math/rand"
	"sort"
)

func buildDashboardView(aggregates []*scanAggregate, sortBy string, baseReward float64) *dashboardView {
	clean := make([]*hostEntry, 0, len(aggregates))
	issues := make([]*hostEntry, 0, len(aggregates))

	for _, agg := range aggregates {
		entry := agg.latest
		if entry == nil {
			continue
		}
		plainPort := 0
		tlsPort := 0
		if entry.Port > 0 {
			if entry.TLS {
				tlsPort = entry.Port
			} else {
				plainPort = entry.Port
			}
		}

		view := buildEntryView(entry, baseReward, pingStats{}, agg.pingStats, agg.jobStats, nil, 0, plainPort, tlsPort)
		view.Host = entry.Host
		view.PoolName = entry.PoolTag
		if view.PoolName == "" {
			view.PoolName = entry.Host
		}
		view.LogFile = ""
		view.ScanURL = "#"
		view.HistoryURL = "#"

		entryView := &hostEntry{
			PoolName: view.PoolName,
			Host: &hostView{
				Host:   entry.Host,
				Latest: view,
			},
			LogFile: "",
		}

		if len(view.Issues) > 0 || !entry.Connected {
			issues = append(issues, entryView)
		} else {
			clean = append(clean, entryView)
		}
	}

	if sortBy == "shuffle" {
		shuffleEntries(clean)
		shuffleEntries(issues)
	} else {
		sortEntries(clean, sortBy)
		sortEntries(issues, sortBy)
	}

	return &dashboardView{
		CleanEntries: clean,
		IssueEntries: issues,
		SortBy:       sortBy,
		HostFilter:   "",
	}
}

func sortEntries(entries []*hostEntry, sortBy string) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entryLess(entries[i].Host.Latest, entries[j].Host.Latest, sortBy)
	})
}

func shuffleEntries(entries []*hostEntry) {
	if len(entries) < 2 {
		return
	}
	rand.Shuffle(len(entries), func(i, j int) {
		entries[i], entries[j] = entries[j], entries[i]
	})
}

func entryLess(a, b *entryView, sortBy string) bool {
	if a == nil || b == nil {
		return a != nil
	}
	switch sortBy {
	case "worker":
		if a.WorkerPercent == b.WorkerPercent {
			return a.TimestampRaw.After(b.TimestampRaw)
		}
		return a.WorkerPercent > b.WorkerPercent
	case "job-wait":
		aSort := a.JobWaitSort
		if aSort == 0 {
			aSort = math.Inf(1)
		}
		bSort := b.JobWaitSort
		if bSort == 0 {
			bSort = math.Inf(1)
		}
		if aSort == bSort {
			return a.TimestampRaw.After(b.TimestampRaw)
		}
		return aSort < bSort
	default:
		aSort := a.PingSort
		if aSort == 0 {
			aSort = math.Inf(1)
		}
		bSort := b.PingSort
		if bSort == 0 {
			bSort = math.Inf(1)
		}
		if aSort == bSort {
			return a.TimestampRaw.After(b.TimestampRaw)
		}
		return aSort < bSort
	}
}
