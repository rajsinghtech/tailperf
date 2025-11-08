package main

import (
	"testing"

	"tailscale.com/ipn/ipnstate"
)

func TestGetConnectionType(t *testing.T) {
	tests := []struct {
		name             string
		peerStatus       *ipnstate.PeerStatus
		expectedType     ConnectionType
		expectedInfoPart string
	}{
		{
			name: "Direct connection",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr: "192.168.1.100:41641",
			},
			expectedType:     ConnTypeDirect,
			expectedInfoPart: "direct to",
		},
		{
			name: "Peer relay connection",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr:   "",
				PeerRelay: "100.64.0.5:41641:0",
			},
			expectedType:     ConnTypePeerRelay,
			expectedInfoPart: "peer-relay via",
		},
		{
			name: "DERP connection",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr:   "",
				PeerRelay: "",
				Relay:     "nyc",
			},
			expectedType:     ConnTypeDERP,
			expectedInfoPart: "derp-",
		},
		{
			name: "Unknown connection",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr:   "",
				PeerRelay: "",
				Relay:     "",
			},
			expectedType:     ConnTypeUnknown,
			expectedInfoPart: "unknown",
		},
		{
			name: "Direct takes precedence over peer relay",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr:   "192.168.1.100:41641",
				PeerRelay: "100.64.0.5:41641:0",
				Relay:     "nyc",
			},
			expectedType:     ConnTypeDirect,
			expectedInfoPart: "direct to",
		},
		{
			name: "Peer relay takes precedence over DERP",
			peerStatus: &ipnstate.PeerStatus{
				CurAddr:   "",
				PeerRelay: "100.64.0.5:41641:0",
				Relay:     "nyc",
			},
			expectedType:     ConnTypePeerRelay,
			expectedInfoPart: "peer-relay via",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connType, connInfo := getConnectionType(tt.peerStatus)

			if connType != tt.expectedType {
				t.Errorf("getConnectionType() type = %v, want %v", connType, tt.expectedType)
			}

			if tt.expectedInfoPart != "" && len(connInfo) > 0 {
				// Just check that the expected part is in the string
				if connInfo[:len(tt.expectedInfoPart)] != tt.expectedInfoPart &&
				   !containsStr(connInfo, tt.expectedInfoPart) {
					t.Errorf("getConnectionType() info = %v, want to contain %v", connInfo, tt.expectedInfoPart)
				}
			}
		})
	}
}

// Helper function for substring checking in tests
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
	       len(s) > len(substr) && contains(s, substr)
}
