package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

var (
	serverMode   = flag.Bool("server", false, "Run in server mode (requires -authkey)")
	clientMode   = flag.Bool("client", false, "Run client using tsnet (creates separate Tailscale node, requires -authkey)")
	hostname     = flag.String("hostname", os.Getenv("TAILPERF_HOSTNAME"), "Tailscale hostname (default: tailperf-<pid> for server/client)")
	authKey      = flag.String("authkey", os.Getenv("TAILPERF_AUTHKEY"), "Tailscale auth key (required for server and -client modes)")
	interval     = flag.Duration("interval", 30*time.Second, "Test interval in client mode")
	testDuration = flag.Duration("duration", 10*time.Second, "Duration of each throughput test")
	stateDir     = flag.String("state", os.Getenv("TAILPERF_STATE_DIR"), "State directory (default: ~/.tailperf/<hostname>)")
	peerFilter   = flag.String("filter", getEnvOrDefault("TAILPERF_FILTER", "tailperf"), "Only test peers with this string in their hostname (empty = test all)")
	targetPeer   = flag.String("target", os.Getenv("TAILPERF_TARGET"), "Test only this specific peer (exact hostname match, takes precedence over -filter)")
)

func main() {
	flag.Parse()

	// Validate flags
	if *serverMode && *clientMode {
		log.Fatal("Cannot use both -server and -client flags")
	}

	// Validate durations
	if *interval < time.Second {
		log.Fatal("Interval must be at least 1 second")
	}
	if *testDuration < time.Second {
		log.Fatal("Test duration must be at least 1 second")
	}
	if *testDuration > 5*time.Minute {
		log.Fatal("Test duration cannot exceed 5 minutes")
	}

	// Setup for modes that need tsnet (server or client)
	if *serverMode || *clientMode {
		if *authKey == "" {
			mode := "Server"
			if *clientMode {
				mode = "Client"
			}
			log.Fatalf("%s mode requires -authkey flag", mode)
		}

		if *hostname == "" {
			*hostname = fmt.Sprintf("tailperf-%d", os.Getpid())
		}

		if *stateDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("Failed to get home directory: %v", err)
			}
			*stateDir = fmt.Sprintf("%s/.tailperf/%s", home, *hostname)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	if *serverMode {
		if err := runServer(ctx); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else if *clientMode {
		if err := runTsnetClient(ctx); err != nil {
			log.Fatalf("Client error: %v", err)
		}
	} else {
		if err := runLocalClient(ctx); err != nil {
			log.Fatalf("Client error: %v", err)
		}
	}
}
