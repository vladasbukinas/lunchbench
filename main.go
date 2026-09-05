package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("\nERROR: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	server := flag.String("server", "http://localhost:11435", "kronk server base url")
	modelName := flag.String("model", "Qwen3-1.7B-Q4_K_M", "model id to benchmark")
	runs := flag.Int("runs", 4, "number of timed runs")
	maxTokens := flag.Int("max-tokens", 300, "max output tokens per run")
	token := flag.String("token", "", "bearer token if server auth is enabled")
	think := flag.Bool("think", true, "allow the model to reason before answering (Qwen3-style thinking)")
	warmup := flag.Bool("warmup", true, "run one untimed warmup request first")
	flag.Parse()

	base := strings.TrimRight(*server, "/")

	if err := checkServer(base); err != nil {
		return err
	}

	loaded, err := checkModel(base, *modelName)
	if err != nil {
		return err
	}
	if !loaded {
		fmt.Printf("WARNING: model %q not in server model list; continuing (server may auto-load it)\n", *modelName)
	}

	fmt.Println("LUNCHBENCH - Kronk local inference benchmark")
	printMachineInfo()
	fmt.Printf("  server      : %s\n", base)
	fmt.Printf("  model       : %s\n", *modelName)
	fmt.Printf("  runs        : %d (max_tokens %d, thinking %t)\n", *runs, *maxTokens, *think)

	m := menu()
	b := newBench(base, *token, *modelName, *think, renderMenu(m))

	ctx := context.Background()

	if *warmup {
		fmt.Print("\nwarmup running... ")
		wctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		if _, err := b.run(wctx, tasks()[0], *maxTokens); err != nil {
			cancel()
			return fmt.Errorf("warmup failed: %w", err)
		}
		cancel()
		fmt.Println("done")
	}

	fmt.Println("\nbenchmark runs:")
	fmt.Println(strings.Repeat("-", 100))

	start := time.Now()
	results := make([]runResult, 0, *runs)
	tk := tasks()

	for i := 0; i < *runs; i++ {
		r, err := b.run(ctx, tk[i%len(tk)], *maxTokens)
		if err != nil {
			return err
		}
		results = append(results, r)
		printRunLine(i, r)
	}

	total := time.Since(start)

	printReport(*modelName, results, summarize(results), total)

	return nil
}

func checkServer(server string) error {
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(server + "/v1/liveness")
	if err != nil {
		return fmt.Errorf("kronk server not reachable at %s; start it with: kronk server start", server)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kronk server health check failed: status %d at %s", resp.StatusCode, server)
	}

	return nil
}

func checkModel(server string, modelName string) (bool, error) {
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(server + "/v1/models")
	if err != nil {
		return true, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return true, nil
	}

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return true, nil
	}

	for _, m := range list.Data {
		if m.ID == modelName {
			return true, nil
		}
	}

	return false, nil
}
