package main

import (
	"fmt"
	"runtime"
	"sort"
	"time"
)

type summary struct {
	minTPS float64
	maxTPS float64
	avgTPS float64

	minWall time.Duration
	maxWall time.Duration
	avgWall time.Duration

	avgTTFT  time.Duration
	totalTok int
	avgPTok  float64
}

func summarize(results []runResult) summary {
	var s summary

	n := float64(len(results))
	var wallSum time.Duration
	var ttftSum time.Duration
	var pTokSum int

	s.minTPS = -1
	s.minWall = -1

	for _, r := range results {
		tps := localTPS(r)
		wall := r.wall

		if s.minTPS < 0 || tps < s.minTPS {
			s.minTPS = tps
		}
		if tps > s.maxTPS {
			s.maxTPS = tps
		}
		s.avgTPS += tps / n

		if s.minWall < 0 || wall < s.minWall {
			s.minWall = wall
		}
		if wall > s.maxWall {
			s.maxWall = wall
		}
		wallSum += wall
		ttftSum += r.ttft
		pTokSum += r.promptTokens
		s.totalTok += r.completionTokens
	}

	s.avgWall = time.Duration(float64(wallSum) / n)
	s.avgTTFT = time.Duration(float64(ttftSum) / n)
	s.avgPTok = float64(pTokSum) / n

	return s
}

func localTPS(r runResult) float64 {
	gen := r.wall - r.ttft
	if gen <= 0 {
		return 0
	}
	return float64(r.completionTokens) / gen.Seconds()
}

func printMachineInfo() {
	fmt.Println("MACHINE")
	fmt.Printf("  os/arch     : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  cpus        : %d\n", runtime.NumCPU())
	fmt.Printf("  go version  : %s\n", runtime.Version())
}

func printRunLine(i int, r runResult) {
	fmt.Printf("  run %-2d [%s] wall %6.0fms  ttft %6.0fms  tps %7.2f  in %4d tok  out %4d tok (reasoning %d)\n",
		i+1, r.task, ms(r.wall), ms(r.ttft), localTPS(r), r.promptTokens, r.completionTokens, r.reasoningTokens)
}

func printReport(modelID string, results []runResult, s summary, total time.Duration) {
	tps := make([]float64, len(results))
	for i := range results {
		tps[i] = localTPS(results[i])
	}
	sort.Float64s(tps)
	p95 := tps[len(tps)-1]
	if len(tps) > 2 {
		p95 = tps[int(float64(len(tps))*0.95+0.999)-1]
	}

	fmt.Println()
	fmt.Println("MODEL")
	fmt.Printf("  model id    : %s\n", modelID)

	fmt.Println()
	fmt.Println("RESULTS")
	fmt.Printf("  runs        : %d\n", len(results))
	fmt.Printf("  total time  : %.1fs\n", total.Seconds())
	fmt.Printf("  tps (ours)  : min %.2f | avg %.2f | max %.2f | p95 %.2f\n", s.minTPS, s.avgTPS, s.maxTPS, p95)
	fmt.Printf("  wall (avg)  : %.0fms | min %.0fms | max %.0fms\n", ms(s.avgWall), ms(s.minWall), ms(s.maxWall))
	fmt.Printf("  ttft (avg)  : %.0fms\n", ms(s.avgTTFT))
	fmt.Printf("  tokens      : %.0f avg in | %d total out\n", s.avgPTok, s.totalTok)

	fmt.Println()
	fmt.Println("ANSWERS")
	for _, r := range results {
		fmt.Printf("--- run [%s] model=%s tps=%.2f server-tps=%.2f server-ttft=%.0fms\n", r.task, r.model, localTPS(r), r.serverTPS, r.serverTTFTMS)
		if r.answer == "" {
			fmt.Println("    (no text content)")
			continue
		}
		for _, line := range wrap(r.answer, 96) {
			fmt.Println("    " + line)
		}
	}
}

func ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func wrap(text string, width int) []string {
	var out []string
	for _, para := range splitLines(text) {
		line := ""
		for _, word := range splitWords(para) {
			switch {
			case line == "":
				line = word
			case len(line)+1+len(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	return append(lines, text[start:])
}

func splitWords(text string) []string {
	var words []string
	start := -1
	for i := 0; i < len(text); i++ {
		if text[i] != ' ' && start < 0 {
			start = i
			continue
		}
		if text[i] == ' ' && start >= 0 {
			words = append(words, text[start:i])
			start = -1
		}
	}
	if start >= 0 {
		words = append(words, text[start:])
	}
	return words
}
