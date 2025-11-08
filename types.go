package main

import (
	"time"
)

type ConnectionType string

const (
	ConnTypeDirect    ConnectionType = "direct"
	ConnTypeDERP      ConnectionType = "derp"
	ConnTypePeerRelay ConnectionType = "peer-relay"
	ConnTypeUnknown   ConnectionType = "unknown"
)

type TestResult struct {
	PeerName       string
	PeerIP         string
	ConnectionType ConnectionType
	ConnectionInfo string
	LatencyMs      float64
	ThroughputMbps float64
	Timestamp      time.Time
	Error          error
}

const (
	perfTestPort = 9898
)
