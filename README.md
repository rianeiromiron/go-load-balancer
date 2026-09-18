# 🚀 Go Channels: Concurrent Load Balancer & Worker Pool

This project demonstrates a simple but practical concurrent HTTP load balancer in Go using channels, goroutines, and a worker pool pattern. It models a system where incoming HTTP requests are queued and processed by a limited set of workers, while preserving backpressure and graceful shutdown behavior.

## ✨ Features

- Fixed-size worker pool with concurrent processing
- Buffered request queue using Go channels
- Per-request response channel for direct client response
- Backpressure via non-blocking `select` with `default`
- HTTP 503 response when the queue is saturated
- Graceful shutdown on `Ctrl + C`, `SIGINT`, or `SIGTERM`

## 🏗️ Architecture

The design is intentionally simple and educational:

- `colaTrabajos` is a buffered channel that stores pending work items.
- A fixed number of goroutines act as workers.
- Each HTTP request creates a unique response channel.
- The router enqueues work and waits for the worker result.
- If the queue is full, the request is rejected immediately instead of blocking.

## ▶️ Run the project

1. Clone the repository:

   ```bash
   git clone https://github.com/rianeiromiron/go-channels-load_balancer.git
   cd go-channels-load_balancer
   ```

2. Start the application:

   ```bash
   go run .
   ```

3. Open a browser or call the endpoint:

   ```bash
   curl "http://localhost:8080/procesar?tarea=mi-tarea"
   ```

## 🌐 HTTP endpoint

The server exposes:

- `GET /procesar?tarea=<valor>`

Example:

```bash
curl "http://localhost:8080/procesar?tarea=procesar-documento"
```

If the queue is full, the server responds with:

- `503 Service Unavailable`

## 🧠 Notes on concurrency

This project is useful for understanding:

- channel-based communication between goroutines
- worker-pool patterns
- bounded queues and load shedding
- graceful shutdown in a long-running service

## 📁 Project structure

```text
.
├── main.go
├── go.mod
├── README.md
```

## 🛠️ Requirements

- Go 1.20 or newer
- A terminal for running the server

## 📌 Example behavior

When the server starts, it launches workers and begins listening on port `8080`. Each request is assigned an ID, placed into the queue, and processed by the next available worker. The response from the worker is sent back to the specific client request that created the response channel.

This makes the example a good reference for channel-driven concurrency and simple API request dispatching in Go.
