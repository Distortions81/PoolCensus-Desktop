package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type PoolCensusApp struct {
	ctx      context.Context
	mu       sync.Mutex
	scanning bool
	lastView *dashboardView
}

type ScanProgress struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	TLS     bool   `json:"tls"`
}

type ScanComplete struct {
	ErrorCount int `json:"errorCount"`
}

func NewPoolCensusApp() *PoolCensusApp {
	return &PoolCensusApp{}
}

func (a *PoolCensusApp) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("Wails startup complete")
}

func (a *PoolCensusApp) StartScan(passes int) (*dashboardView, error) {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return nil, fmt.Errorf("scan already running")
	}
	a.scanning = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.scanning = false
		a.mu.Unlock()
	}()

	if passes <= 0 {
		passes = defaultScanPasses
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}
	customPools := filepath.Join(filepath.Dir(exePath), "pools.json")
	poolsPath := ""
	if info, err := os.Stat(customPools); err == nil && !info.IsDir() {
		poolsPath = customPools
	}

	poolsData, err := loadPools(poolsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load pools: %w", err)
	}

	agent := loadRandomAgent()
	wallet, err := generateRandomWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to generate wallet: %w", err)
	}
	worker := generateWorkerName()
	username := wallet + "." + worker

	targets := collectTargets(poolsData, "")
	if len(targets) == 0 {
		return nil, fmt.Errorf("no pool targets found")
	}

	aggregates := scanTargets(targets, agent, username, wallet, worker, scanOptions{
		Passes: passes,
		OnProgress: func(current, total int, target scanTarget) {
			a.emitProgress(current, total, target)
		},
	})

	if len(aggregates) == 0 {
		return nil, fmt.Errorf("no data collected from pools")
	}

	view := buildDashboardView(aggregates, defaultSortBy, defaultBaseReward)

	a.mu.Lock()
	a.lastView = view
	a.mu.Unlock()

	a.emitComplete(totalErrorCount(aggregates))

	return view, nil
}

func (a *PoolCensusApp) LastReport() *dashboardView {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastView == nil {
		return &dashboardView{}
	}
	return a.lastView
}

func (a *PoolCensusApp) emitProgress(current, total int, target scanTarget) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "scanProgress", ScanProgress{
		Current: current,
		Total:   total,
		Host:    target.Host,
		Port:    target.Port,
		TLS:     target.TLS,
	})
}

func (a *PoolCensusApp) emitComplete(errors int) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "scanComplete", ScanComplete{ErrorCount: errors})
}

func totalErrorCount(aggs []*scanAggregate) int {
	total := 0
	for _, agg := range aggs {
		total += agg.errorCount
	}
	return total
}
