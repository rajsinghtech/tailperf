package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"tailscale.com/tsnet"
)

func runServer(ctx context.Context) error {
	srv := &tsnet.Server{
		Hostname: *hostname,
		Dir:      *stateDir,
		AuthKey:  *authKey,
	}
	defer srv.Close()

	log.Printf("Starting tailperf server as %s", *hostname)

	if _, err := srv.Up(ctx); err != nil {
		return fmt.Errorf("failed to bring up tsnet: %w", err)
	}

	listener, err := srv.Listen("tcp", fmt.Sprintf(":%d", perfTestPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	log.Printf("Server listening on :%d", perfTestPort)

	// Run server in background
	errCh := make(chan error, 1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Accept error: %v", err)
					continue
				}
			}

			go handleConnection(ctx, conn)
		}
	}()

	// Wait for context cancellation
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", remoteAddr)

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Read the test type (1 byte: 'T' for throughput test)
	buf := make([]byte, 1)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Printf("Failed to read test type from %s: %v", remoteAddr, err)
		return
	}

	switch buf[0] {
	case 'T': // Throughput test
		handleThroughputTest(ctx, conn, remoteAddr)
	default:
		log.Printf("Unknown test type from %s: %c", remoteAddr, buf[0])
	}
}

func handleThroughputTest(ctx context.Context, conn net.Conn, remoteAddr string) {
	// Remove read deadline for sending
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	totalBytes := int64(0)
	startTime := time.Now()

	// Pre-allocate buffer for random data
	data := make([]byte, 64*1024) // 64 KB chunks

	// Send data until client closes or context is cancelled
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Generate fresh random data for each chunk (security best practice)
		if _, err := rand.Read(data); err != nil {
			log.Printf("Failed to generate random data: %v", err)
			return
		}

		n, err := conn.Write(data)
		if err != nil {
			duration := time.Since(startTime)
			if duration.Seconds() > 0 {
				log.Printf("Throughput test with %s completed: %d bytes in %.2fs (%.2f Mbps)",
					remoteAddr, totalBytes, duration.Seconds(),
					float64(totalBytes)*8/duration.Seconds()/1e6)
			}
			return
		}
		totalBytes += int64(n)
	}
}
