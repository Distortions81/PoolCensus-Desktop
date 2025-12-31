package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PoolsData struct {
	Pools []PoolDefinition `json:"pools"`
}

type PoolDefinition struct {
	PoolID    string         `json:"pool_id"`
	Name      string         `json:"name"`
	Website   string         `json:"website"`
	Type      string         `json:"type"`
	Endpoints []PoolEndpoint `json:"endpoints"`
}

type PoolEndpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	TLS  bool   `json:"tls"`
}

//go:embed pools.json
var embeddedPools []byte

func loadPools(path string) (*PoolsData, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return parsePools(data)
	}
	return parsePools(embeddedPools)
}

func parsePools(data []byte) (*PoolsData, error) {
	var pools PoolsData
	if err := json.Unmarshal(data, &pools); err != nil {
		return nil, fmt.Errorf("failed to parse pools.json: %w", err)
	}
	return &pools, nil
}

func filterPools(pools *PoolsData, filter string) []PoolDefinition {
	if pools == nil {
		return nil
	}
	if strings.TrimSpace(filter) == "" {
		return pools.Pools
	}
	lower := strings.ToLower(filter)
	var selected []PoolDefinition
	for _, pool := range pools.Pools {
		if strings.Contains(strings.ToLower(pool.PoolID), lower) || strings.Contains(strings.ToLower(pool.Name), lower) {
			selected = append(selected, pool)
		}
	}
	return selected
}
