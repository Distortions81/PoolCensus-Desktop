package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	if view, err := loadCachedDashboardView(); err != nil {
		log.Printf("load cached report: %v", err)
	} else if view != nil {
		a.mu.Lock()
		a.lastView = view
		a.mu.Unlock()
		log.Printf("loaded cached report")
	}
	log.Printf("Wails startup complete")
}

func (a *PoolCensusApp) StartScan(passes int) (*dashboardView, error) {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return nil, fmt.Errorf("scan already running")
	}
	a.scanning = true
	lastView := a.lastView
	a.mu.Unlock()

	if passes <= 0 {
		passes = defaultScanPasses
	}

	go func(passes int) {
		defer func() {
			a.mu.Lock()
			a.scanning = false
			a.mu.Unlock()
		}()

		view, err := a.runScan(passes)
		if err != nil {
			log.Printf("scan failed: %v", err)
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "scanError", map[string]string{"message": err.Error()})
			}
			return
		}

		a.mu.Lock()
		a.lastView = view
		a.mu.Unlock()

		if err := saveCachedDashboardView(view); err != nil {
			log.Printf("save cached report: %v", err)
		}

		a.emitComplete(totalIssueCount(view))
	}(passes)

	if lastView == nil {
		return &dashboardView{}, nil
	}
	return lastView, nil
}

func (a *PoolCensusApp) LastReport() *dashboardView {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lastView == nil {
		return &dashboardView{}
	}
	return a.lastView
}

func (a *PoolCensusApp) runScan(passes int) (*dashboardView, error) {
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

	return buildDashboardView(aggregates, defaultSortBy, defaultBaseReward), nil
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

func totalIssueCount(view *dashboardView) int {
	if view == nil {
		return 0
	}
	return len(view.IssueEntries)
}

type dashboardCache struct {
	Version int            `json:"version"`
	SavedAt string         `json:"savedAt"`
	View    *dashboardView `json:"view"`
}

func cachedDashboardViewPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		return "", fmt.Errorf("unable to determine user cache dir")
	}
	dir := filepath.Join(cacheDir, "PoolCensus")
	return filepath.Join(dir, "last_report.json"), nil
}

func loadCachedDashboardView() (*dashboardView, error) {
	path, err := cachedDashboardViewPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache dashboardCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	if cache.Version != 1 {
		return nil, nil
	}
	if cache.View == nil {
		return nil, nil
	}
	return cache.View, nil
}

func saveCachedDashboardView(view *dashboardView) error {
	if view == nil {
		return nil
	}
	path, err := cachedDashboardViewPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	cache := dashboardCache{
		Version: 1,
		SavedAt: time.Now().UTC().Format(time.RFC3339),
		View:    view,
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
