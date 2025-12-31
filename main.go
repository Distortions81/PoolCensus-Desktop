package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	verbose    bool
	scansPerRun int
)

const (
	defaultSortBy = "ping"
	defaultOutput = "report.html"
	defaultScanPasses = 3
)

func main() {
	flag.IntVar(&scansPerRun, "scans", defaultScanPasses, "Number of times to scan every configured server")
	flag.BoolVar(&verbose, "verbose", false, "Show detailed scanning logs")
	flag.Parse()

	if scansPerRun <= 0 {
		scansPerRun = defaultScanPasses
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to resolve executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	customPools := filepath.Join(exeDir, "pools.json")
	poolsPath := ""
	if info, err := os.Stat(customPools); err == nil && !info.IsDir() {
		poolsPath = customPools
		fmt.Printf("Using custom pools.json from %s\n", customPools)
	}
	poolsData, err := loadPools(poolsPath)
	if err != nil {
		log.Fatalf("failed to load pools: %v", err)
	}

	agent := loadRandomAgent()
	wallet, err := generateRandomWallet()
	if err != nil {
		log.Fatalf("failed to generate wallet: %v", err)
	}
	worker := generateWorkerName()
	username := wallet + "." + worker

	targets := collectTargets(poolsData, "")
	if len(targets) == 0 {
		log.Fatalf("no pool targets found")
	}

	aggregates := scanTargets(targets, agent, username, wallet, worker, scansPerRun)
	if len(aggregates) == 0 {
		log.Fatalf("no data collected from pools")
	}

	view := buildDashboardView(aggregates, defaultSortBy, defaultBaseReward)

	if err := os.MkdirAll(filepath.Dir(defaultOutput), 0o755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}
	f, err := os.Create(defaultOutput)
	if err != nil {
		log.Fatalf("failed to create report file: %v", err)
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, "dashboard.tmpl", view); err != nil {
		log.Fatalf("failed to render dashboard template: %v", err)
	}

	errorCount := totalErrorCount(aggregates)
	if errorCount > 0 && !verbose {
		fmt.Printf("\nCompleted with %d connection issues; rerun with -verbose for details\n", errorCount)
	}
	fmt.Printf("Report written to %s\n", defaultOutput)
}

func totalErrorCount(aggs []*scanAggregate) int {
	total := 0
	for _, agg := range aggs {
		total += agg.errorCount
	}
	return total
}
