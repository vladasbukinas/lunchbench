# lunchbench

CLI benchmark for the [kronk](https://github.com/ardanlabs/kronk) local inference server. Sends a fake lunch-menu workload to an OpenAI-compatible chat endpoint over SSE and reports wall time, TTFT, and tokens/sec.

## Commands

- Build/run: `go run .` (or `go build .` — the `lunchbench` binary in the root is a local build artifact, gitignored)
- The benchmark needs a running kronk server (default `http://localhost:11435`; start with `kronk server start`). Verify quickly with `curl http://localhost:11435/v1/liveness`.
- Flags: `-server`, `-model` (default `Qwen3-1.7B-Q4_K_M`), `-runs`, `-max-tokens`, `-think`, `-warmup`, `-token`.
- No tests, CI, or lint config exist.

## Structure

Single `main` package:
- `main.go` — flags, server/model prechecks, run loop
- `benchmark.go` — tasks, SSE streaming request/response handling
- `restaurants.go` — hardcoded menu data (`menu()`) used as the prompt payload
- `report.go` — aggregation and output formatting

## Gotchas

- The kronk SDK (`github.com/ardanlabs/kronk/sdk/kronk/model`) is used only for `model.D` (an ordered map) and `model.ChatResponse` types — it's the sole dependency despite a large transitive tree in `go.mod`.
- Thinking mode is toggled via the non-standard `chat_template_kwargs: {enable_thinking: ...}` field (Qwen3-style), not a standard OpenAI field.
- Server metrics (`tokens_per_second`, `time_to_first_token_ms`) come from the final usage-only SSE chunk (a chunk with no choices); the report also computes local TPS as `completionTokens / (wall - ttft)`.
