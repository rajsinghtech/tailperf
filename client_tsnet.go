package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"strings"
	"time"

	"tailscale.com/client/local"
	"tailscale.com/tsnet"
)

// runTsnetClient runs a tsnet-based client (creates its own Tailscale node)
func runTsnetClient(ctx context.Context) error {
	srv := &tsnet.Server{
		Hostname: *hostname,
		Dir:      *stateDir,
		AuthKey:  *authKey,
	}
	defer srv.Close()

	log.Printf("Starting tailperf tsnet-client as %s", *hostname)

	if _, err := srv.Up(ctx); err != nil {
		return fmt.Errorf("failed to bring up tsnet: %w", err)
	}

	lc, err := srv.LocalClient()
	if err != nil {
		return fmt.Errorf("failed to get local client: %w", err)
	}

	log.Printf("Client started, testing every %v", *interval)
	log.Printf("Test duration: %v per peer", *testDuration)
	fmt.Println()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Run initial test immediately
	runTestsTsnet(ctx, srv, lc)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runTestsTsnet(ctx, srv, lc)
		}
	}
}

func runTestsTsnet(ctx context.Context, srv *tsnet.Server, lc *local.Client) {
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
		result := testPeerTsnet(ctx, srv, lc, peerIP, peerNames[peerIP])
		results = append(results, result)
	}

	printResults(results)
}

func testPeerTsnet(ctx context.Context, srv *tsnet.Server, lc *local.Client, peerIP netip.Addr, peerName string) TestResult {
	result := TestResult{
		PeerName:  peerName,
		PeerIP:    peerIP.String(),
		Timestamp: time.Now(),
	}

	log.Printf("Testing peer %s (%s) - getting connection info...", peerName, peerIP)

	// Get connection type
	connType, connInfo, err := getDetailedConnectionInfo(ctx, lc, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("connection info: %w", err)
		log.Printf("  Error getting connection info: %v", err)
		return result
	}
	result.ConnectionType = connType
	result.ConnectionInfo = connInfo
	log.Printf("  Connection type: %s", connType)

	// Measure latency
	log.Printf("  Measuring latency...")
	latency, err := measureLatency(ctx, lc, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("latency: %w", err)
		log.Printf("  Error measuring latency: %v", err)
		return result
	}
	result.LatencyMs = latency
	log.Printf("  Latency: %.2fms", latency)

	// Measure throughput using tsnet
	log.Printf("  Measuring throughput...")
	throughput, err := measureThroughput(ctx, srv, peerIP)
	if err != nil {
		result.Error = fmt.Errorf("throughput: %w", err)
		log.Printf("  Error measuring throughput: %v", err)
		return result
	}
	result.ThroughputMbps = throughput
	log.Printf("  Throughput: %.2f Mbps", throughput)

	return result
}
