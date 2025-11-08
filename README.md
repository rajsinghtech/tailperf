# tailperf

Network performance testing tool for Tailscale networks. Monitors latency, throughput, and connection types between nodes.

## Modes

- **Server Mode** (`-server`): Standalone Tailscale node that listens for performance tests
- **Client Mode** (default): Uses your existing Tailscale connection to test servers
- **Tsnet Client Mode** (`-client`): Standalone Tailscale client node for testing

## Features

- Continuous performance monitoring at configurable intervals
- Connection type detection (direct, DERP, peer-relay)
- Latency measurement via Tailscale's ping API
- TCP throughput testing
- No client configuration needed

## Quick Start

### Binary

```bash
go build -o tailperf
```

### Docker

**Using Docker Compose:**
```bash
# Copy and configure environment variables
cp .env.example .env
# Edit .env with your Tailscale auth key

# Deploy with Docker Compose
docker compose up -d
```

**Using Docker Run:**

Server mode:
```bash
docker run -d \
  --name tailperf-server \
  -e TAILPERF_AUTHKEY=tskey-auth-xxxxx \
  -e TAILPERF_HOSTNAME=tailperf-server \
  -e TAILPERF_STATE_DIR=/var/lib/tailperf \
  -v tailperf-data:/var/lib/tailperf \
  ghcr.io/rajsinghtech/tailperf:latest -server
```

Client mode (tsnet):
```bash
docker run -d \
  --name tailperf-client \
  -e TAILPERF_AUTHKEY=tskey-auth-xxxxx \
  -e TAILPERF_HOSTNAME=tailperf-client \
  -e TAILPERF_STATE_DIR=/var/lib/tailperf \
  -v tailperf-client-data:/var/lib/tailperf \
  ghcr.io/rajsinghtech/tailperf:latest -client -interval=20s -duration=5s
```

## Usage

### Server Mode

```bash
./tailperf -server -authkey=tskey-auth-xxx...
```

Creates a standalone Tailscale node and listens on port 9898 for performance tests.

### Client Mode (Default)

```bash
./tailperf
```

Uses your existing Tailscale connection to test servers. No auth key needed.

### Tsnet Client Mode

```bash
./tailperf -client -authkey=tskey-auth-xxx...
```

Creates a standalone Tailscale client node. Useful for machines without Tailscale installed.

### Configuration

Supports both command-line flags and environment variables:

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-server` | - | false | Run in server mode |
| `-client` | - | false | Run in tsnet client mode |
| `-authkey` | `TAILPERF_AUTHKEY` | - | Tailscale auth key (required for server/client modes) |
| `-hostname` | `TAILPERF_HOSTNAME` | `tailperf-<pid>` | Custom hostname |
| `-state` | `TAILPERF_STATE_DIR` | `~/.tailperf/<hostname>` | State directory |
| `-interval` | - | 30s | Test interval |
| `-duration` | - | 10s | Test duration |
| `-filter` | `TAILPERF_FILTER` | tailperf | Only test peers with this substring in hostname |
| `-target` | `TAILPERF_TARGET` | - | Test only this specific peer (exact match, overrides `-filter`) |

**Important:** Copy `.env.example` to `.env` and configure your auth key before using Docker.

## Example Output

```
=== Performance Test Results (2025-01-08 10:30:45) ===

Peer: tailperf-server-1 (100.64.0.2)
  Connection: direct - direct to 192.168.1.100:41641 (5.23ms)
  Latency: 5.23 ms
  Throughput: 3692.80 Mbps

Peer: tailperf-server-2 (100.64.0.3)
  Connection: derp - derp-nyc (25.45ms)
  Latency: 25.45 ms
  Throughput: 234.56 Mbps

---
```


## Connection Types

- **Direct**: Peer-to-peer connection (fastest)
- **DERP**: Relayed via Tailscale's servers
- **Peer Relay**: Tunneled through another peer


## Troubleshooting

**Client mode requires Tailscale daemon:**
- Ensure Tailscale is running: `tailscale status`
- Connect to your tailnet: `tailscale up`

**Server/client modes require auth key:**
- Generate at https://login.tailscale.com/admin/settings/keys
- Store in `.env` file (recommended) or pass via `-authkey` flag

**No peers found:**
- Verify servers are running with `-server` flag
- Check `tailscale status` shows the peers online

## License

MIT
