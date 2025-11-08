package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"time"

	"tailscale.com/tsnet"
)

// measureThroughput is used by server mode (with tsnet)
func measureThroughput(ctx context.Context, srv *tsnet.Server, peerIP netip.Addr) (float64, error) {
	// Create context with timeout for the test
	testCtx, cancel := context.WithTimeout(ctx, *testDuration+5*time.Second)
	defer cancel()

	// Dial the peer
	conn, err := srv.Dial(testCtx, "tcp", fmt.Sprintf("%s:%d", peerIP, perfTestPort))
	if err != nil {
		return 0, fmt.Errorf("failed to dial peer: %w", err)
	}
	defer conn.Close()

	// Send test type ('T' for throughput)
	if _, err := conn.Write([]byte{'T'}); err != nil {
		return 0, fmt.Errorf("failed to send test type: %w", err)
	}

	// Read data for the specified duration
	buf := make([]byte, 64*1024) // 64 KB buffer
	totalBytes := int64(0)
	startTime := time.Now()
	deadline := startTime.Add(*testDuration)

	conn.SetDeadline(deadline)

	for time.Now().Before(deadline) {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Check if it's a timeout (expected when duration expires)
			if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
				break
			}
			return 0, fmt.Errorf("read error: %w", err)
		}
		totalBytes += int64(n)
	}

	duration := time.Since(startTime)

	// Calculate throughput in Mbps
	if duration.Seconds() == 0 {
		return 0, fmt.Errorf("test duration too short")
	}

	throughputMbps := float64(totalBytes) * 8 / duration.Seconds() / 1e6

	return throughputMbps, nil
}

// measureThroughputDirect is used by client mode (using system networking)
func measureThroughputDirect(ctx context.Context, peerIP netip.Addr) (float64, error) {
	// Create context with timeout for the test
	testCtx, cancel := context.WithTimeout(ctx, *testDuration+5*time.Second)
	defer cancel()

	// Use a dialer with the context
	var d net.Dialer
	conn, err := d.DialContext(testCtx, "tcp", fmt.Sprintf("%s:%d", peerIP, perfTestPort))
	if err != nil {
		return 0, fmt.Errorf("failed to dial peer: %w", err)
	}
	defer conn.Close()

	// Send test type ('T' for throughput)
	if _, err := conn.Write([]byte{'T'}); err != nil {
		return 0, fmt.Errorf("failed to send test type: %w", err)
	}

	// Read data for the specified duration
	buf := make([]byte, 64*1024) // 64 KB buffer
	totalBytes := int64(0)
	startTime := time.Now()
	deadline := startTime.Add(*testDuration)

	conn.SetDeadline(deadline)

	for time.Now().Before(deadline) {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Check if it's a timeout (expected when duration expires)
			if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
				break
			}
			return 0, fmt.Errorf("read error: %w", err)
		}
		totalBytes += int64(n)
	}

	duration := time.Since(startTime)

	// Calculate throughput in Mbps
	if duration.Seconds() == 0 {
		return 0, fmt.Errorf("test duration too short")
	}

	throughputMbps := float64(totalBytes) * 8 / duration.Seconds() / 1e6

	return throughputMbps, nil
}
