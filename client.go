package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"sort"
	"strings"
	"time"

	"tailscale.com/client/local"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func runLocalClient(ctx context.Context) error {
	// Connect to local Tailscale daemon
	lc := &local.Client{}

	log.Println("Connecting to local Tailscale daemon...")

	// Verify connection by getting status
	status, err := lc.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to local Tailscale daemon: %w\n"+
			"Make sure Tailscale is running on this machine", err)
	}

	if status.BackendState != "Running" {
		return fmt.Errorf("Tailscale is not in Running state (current: %s)", status.BackendState)
	}

	selfName := "unknown"
	if status.Self != nil {
		selfName = status.Self.HostName
		if status.Self.DNSName != "" {
			selfName = status.Self.DNSName
		}
	}

	log.Printf("Connected to Tailscale as %s", selfName)
	log.Printf("Testing every %v", *interval)
	log.Printf("Test duration: %v per peer", *testDuration)
	fmt.Println()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Run initial test immediately
	runTests(ctx, lc)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runTests(ctx, lc)
		}
	}
}

func runTests(ctx context.Context, lc *local.Client) {
	status, err := lc.Status(ctx)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
		return
	}

	// Get list of online peers (excluding self)
	var peers []netip.Addr
	peerNames := make(map[netip.Addr]string)

	for _, ps := range status.Peer {
		if !ps.Online {
			continue
		}
		if len(ps.TailscaleIPs) == 0 {
			continue
		}

		name := ps.HostName
		if ps.DNSName != "" {
			name = ps.DNSName
		}

		// If target is specified, only test exact match
		if *targetPeer != "" {
			if !strings.EqualFold(ps.HostName, *targetPeer) &&
			   !strings.EqualFold(ps.DNSName, *targetPeer) &&
			   !strings.EqualFold(strings.TrimSuffix(ps.DNSName, "."), *targetPeer) {
				continue
			}
		} else if *peerFilter != "" && !contains(name, *peerFilter) {
			// Otherwise apply peer filter
			continue
		}

		peerIP := ps.TailscaleIPs[0]
		peers = append(peers, peerIP)
		peerNames[peerIP] = name
	}

	if len(peers) == 0 {
		log.Println("No online peers found")
		return
	}

	log.Printf("Found %d online peer(s), starting tests...", len(peers))
	fmt.Println()

	results := make([]TestResult, 0, len(peers))

	for _, peerIP := range peers {
		result := testPeer(ctx, lc, peerIP, peerNames[peerIP])
		results = append(results, result)
	}

	printResults(results)
}

func testPeer(ctx context.Context, lc *local.Client, peerIP netip.Addr, peerName string) TestResult {
	result := TestResult{
		PeerName:  peerName,
		PeerIP:    peerIP.String(),
		Timestamp: time.Now(),
	}

	// Get connection type
	connType, connInfo, err := getDetailedConnectionInfo(ctx, lc, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("connection info: %w", err)
		return result
	}
	result.ConnectionType = connType
	result.ConnectionInfo = connInfo

	// Measure latency
	latency, err := measureLatency(ctx, lc, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("latency: %w", err)
		return result
	}
	result.LatencyMs = latency

	// Measure throughput
	throughput, err := measureThroughputDirect(ctx, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("throughput: %w", err)
		return result
	}
	result.ThroughputMbps = throughput

	return result
}

func printResults(results []TestResult) {
	// Sort by peer name for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].PeerName < results[j].PeerName
	})

	fmt.Printf("=== Performance Test Results (%s) ===\n\n", time.Now().Format("2006-01-02 15:04:05"))

	for _, r := range results {
		fmt.Printf("Peer: %s (%s)\n", r.PeerName, r.PeerIP)

		if r.Error != nil {
			fmt.Printf("  Status: ERROR - %v\n\n", r.Error)
			continue
		}

		fmt.Printf("  Connection: %s - %s\n", r.ConnectionType, r.ConnectionInfo)
		fmt.Printf("  Latency: %.2f ms\n", r.LatencyMs)
		fmt.Printf("  Throughput: %.2f Mbps\n", r.ThroughputMbps)
		fmt.Println()
	}

	fmt.Println("---")
	fmt.Println()
}
