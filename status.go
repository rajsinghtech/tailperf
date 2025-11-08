package main

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"tailscale.com/client/local"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
)

func getConnectionType(ps *ipnstate.PeerStatus) (ConnectionType, string) {
	// Direct connection: has CurAddr
	if ps.CurAddr != "" {
		return ConnTypeDirect, fmt.Sprintf("direct to %s", ps.CurAddr)
	}

	// Peer relay: tunneled through another peer
	if ps.PeerRelay != "" {
		return ConnTypePeerRelay, fmt.Sprintf("peer-relay via %s", ps.PeerRelay)
	}

	// DERP relay: using Tailscale's relay servers
	if ps.Relay != "" {
		return ConnTypeDERP, fmt.Sprintf("derp-%s", ps.Relay)
	}

	return ConnTypeUnknown, "unknown"
}

func getDetailedConnectionInfo(ctx context.Context, lc *local.Client, peerIP netip.Addr) (ConnectionType, string, error) {
	// First try to get basic status
	status, err := lc.Status(ctx)
	if err != nil {
		return ConnTypeUnknown, "", fmt.Errorf("failed to get status: %w", err)
	}

	// Find the peer in status
	var peerStatus *ipnstate.PeerStatus
	for _, ps := range status.Peer {
		for _, ip := range ps.TailscaleIPs {
			if ip == peerIP {
				peerStatus = ps
				break
			}
		}
		if peerStatus != nil {
			break
		}
	}

	if peerStatus == nil {
		return ConnTypeUnknown, "", fmt.Errorf("peer not found in status")
	}

	connType, connInfo := getConnectionType(peerStatus)

	// Optionally do a ping for more detailed info with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pr, err := lc.Ping(pingCtx, peerIP, tailcfg.PingDisco)
	if err == nil && pr.Err == "" {
		// Ping succeeded, refine connection info
		if pr.Endpoint != "" {
			return ConnTypeDirect, fmt.Sprintf("direct to %s (%.2fms)", pr.Endpoint, pr.LatencySeconds*1000), nil
		}
		if pr.DERPRegionCode != "" {
			return ConnTypeDERP, fmt.Sprintf("derp-%s (%.2fms)", pr.DERPRegionCode, pr.LatencySeconds*1000), nil
		}
	}

	return connType, connInfo, nil
}

func measureLatency(ctx context.Context, lc *local.Client, peerIP netip.Addr) (float64, error) {
	// Add timeout for ping
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pr, err := lc.Ping(pingCtx, peerIP, tailcfg.PingDisco)
	if err != nil {
		return 0, fmt.Errorf("ping failed: %w", err)
	}

	if pr.Err != "" {
		return 0, fmt.Errorf("ping error: %s", pr.Err)
	}

	return pr.LatencySeconds * 1000, nil // Convert to milliseconds
}
