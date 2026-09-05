# lunchbench

CLI benchmark for the [kronk](https://github.com/ardanlabs/kronk) local inference server. Sends a lunch-menu workload to the chat endpoint and reports wall time, TTFT, and tokens/sec.

## Run

Start a kronk server, then:

```
go run .
```

Flags:

```
-server      kronk server urls, comma-separated to compare backends (default http://localhost:11435)
-model       model id (default Qwen3-1.7B-Q4_K_M)
-runs        number of timed runs per server (default 4)
-max-tokens  max output tokens per run (default 300)
-think       allow Qwen3-style thinking (default true)
-warmup      run one untimed warmup request (default true)
-token       bearer token if server auth is enabled
```

The inference backend (cpu, cuda, vulkan, ...) is fixed per kronk server *process*: it is set at startup (`--processor` flag or `KRONK_PROCESSOR` env var) and cannot be changed per request — clients can only detect it. Lunchbench does that via `GET /v1/kronk/libs` and shows the backend per server. To compare backends, start one server per backend (each with its own `--api-host` port) and pass them all:

```
kronk server start --processor cpu    --api-host localhost:11435
kronk server start --processor vulkan --api-host localhost:11435

go run . -server http://localhost:11435
```

Example:

```
go run . -server http://localhost:11435 -model Qwen3-1.7B-Q4_K_M -runs 4
```
