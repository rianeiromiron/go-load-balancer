# 🚀 Concurrent HTTP Load Balancer & Worker Pool in Go

A robust, production-grade concurrent task dispatcher and load balancer built in Go, demonstrating idiomatic concurrency patterns, channel multiplexing, backpressure handling, and graceful shutdown.

## 🏗️ Architecture Overview

- **Bounded Worker Pool:** Fixed number of worker goroutines processing heavy operations concurrently.
- **Request Dispatching via Channels:** Incoming HTTP requests are converted into work items and placed on a buffered channel queue.
- **Per-Request Response Channels:** Each HTTP handler listens on a dedicated one-shot channel to receive its specific result from the worker.
- **Backpressure / Load Shedding:** Uses non-blocking `select` with `default` to immediately return `HTTP 503 Service Unavailable` when queue capacity is reached.
- **Graceful Shutdown:** Intercepts OS signals (`SIGINT`, `SIGTERM`) to cleanly finish in-flight tasks and close network connections.

## 🚀 How to Run

1. Clone the repository:
   ```bash
   git clone https://github.com/tu-usuario/go-load-balancer.git
   cd go-load-balancer