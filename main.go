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
	server := flag.String("server", "http://localhost:11435", "kronk server base urls (comma-separated to compare inference backends)")
	modelName := flag.String("model", "Qwen3-1.7B-Q4_K_M", "model id to benchmark")
	runs := flag.Int("runs", 4, "number of timed runs")
	maxTokens := flag.Int("max-tokens", 300, "max output tokens per run")
	token := flag.String("token", "", "bearer token if server auth is enabled")
	think := flag.Bool("think", true, "allow the model to reason before answering (Qwen3-style thinking)")
	warmup := flag.Bool("warmup", true, "run one untimed warmup request first")
	flag.Parse()

	servers := splitList(*server)
	if len(servers) == 0 {
		return fmt.Errorf("no server urls provided")
	}

	fmt.Println("LUNCHBENCH - Kronk local inference benchmark")
	printMachineInfo()
	fmt.Printf("  model       : %s\n", *modelName)
	fmt.Printf("  runs        : %d per server (max_tokens %d, thinking %t)\n", *runs, *maxTokens, *think)

	m := menu()
	menuTxt := renderMenu(m)

	results := make([]serverRun, 0, len(servers))
	for _, base := range servers {
		sr, err := benchmarkServer(base, *token, *modelName, *think, menuTxt, *runs, *maxTokens, *warmup)
		if err != nil {
			return err
		}
		results = append(results, sr)
	}

	if len(results) > 1 {
		printComparison(results)
	}

	return nil
}

func benchmarkServer(base string, token string, modelName string, think bool, menuTxt string, numRuns int, maxTokens int, warmup bool) (serverRun, error) {
	var sr serverRun

	if err := checkServer(base); err != nil {
		return sr, err
	}

	loaded, err := checkModel(base, modelName)
	if err != nil {
		return sr, err
	}
	if !loaded {
		fmt.Printf("WARNING: model %q not in server model list; continuing (server may auto-load it)\n", modelName)
	}

	sr.base = base
	sr.processor = detectProcessor(base, token)

	fmt.Printf("\nSERVER %s\n", base)
	fmt.Printf("  inference   : %s\n", sr.processor)
	fmt.Printf("  model       : %s\n", modelName)
	fmt.Printf("  runs        : %d (max_tokens %d, thinking %t)\n", numRuns, maxTokens, think)

	b := newBench(base, token, modelName, think, menuTxt)

	ctx := context.Background()

	if warmup {
		fmt.Print("\nwarmup running... ")
		wctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		if _, err := b.run(wctx, tasks()[0], maxTokens); err != nil {
			cancel()
			return sr, fmt.Errorf("warmup failed: %w", err)
		}
		cancel()
		fmt.Println("done")
	}

	fmt.Println("\nbenchmark runs:")
	fmt.Println(strings.Repeat("-", 100))

	start := time.Now()
	results := make([]runResult, 0, numRuns)
	tk := tasks()

	for i := 0; i < numRuns; i++ {
		r, err := b.run(ctx, tk[i%len(tk)], maxTokens)
		if err != nil {
			return sr, err
		}
		results = append(results, r)
		printRunLine(i, r)
	}

	sr.total = time.Since(start)
	sr.results = results
	sr.sum = summarize(results)

	printReport(modelName, sr.processor, results, sr.sum, sr.total)

	return sr, nil
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.TrimRight(p, "/"))
		}
	}
	return out
}

func detectProcessor(server string, token string) string {
	client := http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(http.MethodGet, server+"/v1/kronk/libs", nil)
	if err != nil {
		return "unknown"
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "unknown"
	}

	var out struct {
		Processor string `json:"processor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.Processor == "" {
		return "unknown"
	}

	return out.Processor
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
