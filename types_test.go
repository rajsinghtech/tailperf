package main

import (
	"testing"
	"time"
)

func TestConnectionTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		connType ConnectionType
		expected string
	}{
		{"Direct connection", ConnTypeDirect, "direct"},
		{"DERP connection", ConnTypeDERP, "derp"},
		{"Peer relay connection", ConnTypePeerRelay, "peer-relay"},
		{"Unknown connection", ConnTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.connType) != tt.expected {
				t.Errorf("ConnectionType = %v, want %v", tt.connType, tt.expected)
			}
		})
	}
}

func TestTestResult(t *testing.T) {
	now := time.Now()
	result := TestResult{
		PeerName:       "test-peer",
		PeerIP:         "100.64.0.1",
		ConnectionType: ConnTypeDirect,
		ConnectionInfo: "direct to 192.168.1.1:41641",
		LatencyMs:      5.5,
		ThroughputMbps: 1000.0,
		Timestamp:      now,
		Error:          nil,
	}

	if result.PeerName != "test-peer" {
		t.Errorf("PeerName = %v, want test-peer", result.PeerName)
	}
	if result.PeerIP != "100.64.0.1" {
		t.Errorf("PeerIP = %v, want 100.64.0.1", result.PeerIP)
	}
	if result.ConnectionType != ConnTypeDirect {
		t.Errorf("ConnectionType = %v, want %v", result.ConnectionType, ConnTypeDirect)
	}
	if result.ConnectionInfo != "direct to 192.168.1.1:41641" {
		t.Errorf("ConnectionInfo = %v, want 'direct to 192.168.1.1:41641'", result.ConnectionInfo)
	}
	if result.LatencyMs != 5.5 {
		t.Errorf("LatencyMs = %v, want 5.5", result.LatencyMs)
	}
	if result.ThroughputMbps != 1000.0 {
		t.Errorf("ThroughputMbps = %v, want 1000.0", result.ThroughputMbps)
	}
	if !result.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", result.Timestamp, now)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestPerfTestPortConstant(t *testing.T) {
	if perfTestPort != 9898 {
		t.Errorf("perfTestPort = %v, want 9898", perfTestPort)
	}
}
