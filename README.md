# Magic Cylinder – WebTransport Ping‑Pong Experiment

Magic Cylinder is an experimental WebTransport application built on **HTTP/3** using [quic-go/webtransport-go](https://github.com/quic-go/webtransport-go). It demonstrates a minimal “ping‑pong” style message relay between two server processes launched from the same code base.

## Overview

The flow starts when the client sends an initial Ping to server1. Each server responds and then proactively opens a fresh WebTransport connection to the other server to forward ("echo") the transformed message, producing a chained sequence:

```
client ──Ping──▶ server1 ──Pong/echo──▶ server2 ──Ping/echo──▶ server1 ── ... (continues)
```

Roles:
- **Client**: Sends only the first Ping; then exits.
- **Server1 (port 8443)**: Receives messages, responds, and echoes to Server2.
- **Server2 (port 8444)**: Same behavior, but targets Server1.

Servers are identical binaries; behavior is driven by CLI flags (port, name, target URL).

## Features
- WebTransport over HTTP/3 (QUIC) draft implementation.
- Bidirectional stream per session; each echo uses a NEW session & stream (no long‑lived connection reuse yet).
- Simple message model: Ping/Pong types with sequence incrementing.
- Layered architecture (Controller / Repository / Entity) for testability.
- Single upgrade endpoint: `/webtransport` and basic `/health` endpoint.
- Optional plaintext echo endpoint: `/plain` (HTTP POST with JSON). Choose by setting the peer target URL to `/plain`.

## Prerequisites
- Go 1.21+
- OpenSSL (for local self‑signed certificates)
- macOS/Linux (Windows should work, but paths/permissions may vary)

## Project Structure
```
magic-cylinder/
├── cmd/
│   ├── client/          # Client entry point
│   └── server/          # Server entry point (run twice with different flags)
├── internal/
│   ├── base.go          # Server lifecycle (TLS + start + shutdown)
│   ├── router.go        # Route registration & dependency wiring
│   ├── controller/      # Controller layer (connection + stream handling)
│   ├── repository/      # Repository layer (message build & echo dialing)
│   └── entity/          # Domain models (Message, types, etc.)
├── certs/               # Generated TLS cert/key (after make certs)
├── bin/                 # Built binaries (after make build)
├── Makefile             # Convenience tasks
└── generate-certs.sh    # Helper script for self‑signed certs
```

## Installation & Setup

### 1. Install dependencies
```bash
make deps
```

### 2. Generate self‑signed certificates
```bash
make certs
```
Certificates are written under `certs/` (e.g. `server.crt`, `server.key`).

### 3. Build binaries
```bash
make build
```
Outputs go to `bin/server` and `bin/client`.

### 4. Run servers
Open two terminals:

Terminal A (server1):
```bash
./bin/server -port 8443 -name server1 -target https://localhost:8444/webtransport
```

Terminal B (server2):
```bash
./bin/server -port 8444 -name server2 -target https://localhost:8443/webtransport
```

### 5. Trigger initial Ping
Terminal C:
```bash
./bin/client -server https://localhost:8443/webtransport
```
The client exits after sending; watch both server logs for the ongoing chain.

### Plaintext mode (optional)
If you want the server-to-server echo to use a simple HTTP POST instead of WebTransport, run servers with the `/plain` endpoint as target. The servers still listen with TLS, so use `https://.../plain`.

Terminal A (server1 -> server2 via plaintext HTTP):
```bash
./bin/server -port 8443 -name server1 -target https://localhost:8444/plain
```

Terminal B (server2 -> server1 via plaintext HTTP):
```bash
./bin/server -port 8444 -name server2 -target https://localhost:8443/plain
```

You can trigger the initial ping via WebTransport as usual, or use the client in plaintext mode directly:

Client (plaintext):
```bash
./bin/client -server https://localhost:8443/plain
```

Or test `/plain` directly with curl:

```bash
curl -k -sS https://localhost:8443/plain \
	-H 'Content-Type: application/json' \
	-d '{"type":"ping","content":"Ping via plain","sequence":1,"from":"curl","to":"server"}' | jq .
```
Notes:
- The plaintext echo client accepts both `http://` and `https://` targets. For `https://` with self-signed certs, verification is skipped internally for development.
- Choosing WebTransport vs plaintext is based solely on the target URL you pass to `-target`.

## Example Log Snippet
```
[Controller] ✅ WebTransport connection established
[Controller] Received ping message: Initial ping from client (seq: 1)
[Repository] ProcessPing started … generates Pong (seq: 1)
[Controller] Sent pong message (seq: 1)
[Repository] SendEchoToTarget started (dialing https://localhost:8444/webtransport)
```

## Command Line Flags

Server:
```bash
./bin/server -port <PORT> -name <NAME> -target <TARGET_URL>
```
| Flag    | Description                                  | Example |
|---------|----------------------------------------------|---------|
| -port   | TCP port to listen on                        | 8443    |
| -name   | Logical server name for log output           | server1 |
| -target | WebTransport URL of the peer (omit to disable echo) | https://localhost:8444/webtransport |
| -delay  | Seconds to sleep before each echo (WebTransport or plaintext) | 2 |

For plaintext echo between servers, set `-target` to the `/plain` endpoint, e.g. `https://localhost:8444/plain`.
You can also introduce a delay:
```bash
./bin/server -port 8443 -name server1 -target https://localhost:8444/plain -delay 1
./bin/server -port 8444 -name server2 -target https://localhost:8443/plain -delay 1
```

Client:
```bash
./bin/client -server <SERVER_URL>
```
| Flag   | Description                    | Default |
|--------|--------------------------------|---------|
| -server| WebTransport endpoint to dial  | https://localhost:8443/webtransport |

## Makefile Tasks
```bash
make deps    # Download modules
make certs   # Generate self‑signed TLS cert/key
make build   # Compile server and client
make clean   # Remove bin/ and build artifacts
make test    # (Reserved) Run tests if added later
```

## Design Notes
- Controller focuses on session & stream handling; delegates message transformation to Repository.
- Repository increments a sequence and constructs the next Ping/Pong payload.
- Each echo creates a fresh WebTransport session (simplifies state, increases overhead). Future improvement: session reuse.
- Error handling wraps root errors with context using `fmt.Errorf("… %w", err)`.

## Limitations & Caveats
- Self‑signed certificates: clients skip verification (`InsecureSkipVerify`) – never use this pattern in production.
- No persistent QUIC session reuse; high churn under heavy load.
- No TLS key logging in current code (for Wireshark QUIC decryption you must modify tls.Config to set KeyLogWriter).
- Minimal validation & no authentication – strictly experimental.
- Termination requires manual Ctrl+C.

## Observability & Debugging
Simple `log.Printf` statements are used throughout. To watch encrypted UDP traffic:
```bash
sudo tcpdump -i lo0 -n -vv -X -s 0 'udp and (port 8443 or port 8444)'
```
QUIC is encrypted; for HTTP/3 frame inspection use Wireshark/tshark with a TLS key log file.

## Future Improvements (Ideas)
- Add TLS key logging for QUIC decryption.
- Introduce configurable delay / backoff between echoes.
- Implement session pooling and stream multiplexing.
- Add structured logging (Zap / Zerolog) and log levels.
- Provide unit tests for Controller and Repository via interface mocks.
- Graceful shutdown hooks for in‑flight streams.

## License
Currently unspecified; treat as internal experimental code unless a LICENSE file is added.

---
Feedback and contributions welcome. This repository is intentionally small so individual HTTP/3/WebTransport behaviors remain easy to trace.