# lunchbench

CLI benchmark for the [kronk](https://github.com/ardanlabs/kronk) local inference server. Sends a lunch-menu workload to the chat endpoint and reports wall time, TTFT, and tokens/sec.

## Run

Start a kronk server, then:

```
go run .
```

Flags:

```
-server      kronk server url (default http://localhost:11435)
-model       model id (default Qwen3-1.7B-Q4_K_M)
-runs        number of timed runs (default 4)
-max-tokens  max output tokens per run (default 300)
-think       allow Qwen3-style thinking (default true)
-warmup      run one untimed warmup request (default true)
-token       bearer token if server auth is enabled
```

Example:

```
go run . -server http://localhost:11435 -model Qwen3-1.7B-Q4_K_M -runs 4
```
