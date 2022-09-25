package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %v <raw_benchmark_output>\n", os.Args[0])
	}
	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}
	lines := strings.Split(string(buf), "\n")

	medians := extractMedians(lines)
	structured := structureStats(medians)
	showTable(structured)
}

type benchmark struct {
	isGEOS    bool
	inputSize int
	op        string
}

func structureStats(medians map[string]time.Duration) map[benchmark]time.Duration {
	benches := make(map[benchmark]time.Duration)
	for name, median := range medians {
		slashParts := strings.Split(name, "/")
		eqParts := strings.Split(slashParts[1], "=")
		inputSize, err := strconv.Atoi(eqParts[1])
		if err != nil {
			panic(err)
		}
		opWithPrefix := strings.Split(slashParts[2], "-")[0]
		opParts := strings.Split(opWithPrefix, "_")
		isGEOS := opParts[0] == "GEOS"
		opName := opParts[1]

		benches[benchmark{
			isGEOS:    isGEOS,
			inputSize: inputSize,
			op:        opName,
		}] = median
	}
	return benches
}

func extractMedians(lines []string) map[string]time.Duration {
	stats := make(map[string][]time.Duration)
	for _, line := range lines {
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		parts := strings.Split(line, "\t")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) != 3 {
			panic(line)
		}

		name := parts[0]
		nsPerOp, err := strconv.Atoi(strings.Split(parts[2], " ")[0])
		if err != nil {
			panic(err)
		}
		stats[name] = append(stats[name], time.Duration(nsPerOp))
	}

	medians := make(map[string]time.Duration)
	for name, results := range stats {
		sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })
		if n := len(results); n%2 == 0 {
			r1 := results[n/2]
			r2 := results[n/2-1]
			medians[name] = (r1 + r2) / 2
		} else {
			medians[name] = results[(n-1)/2]
		}
	}
	return medians
}

func showTable(benches map[benchmark]time.Duration) {
	for _, op := range []string{
		"Intersection",
		"Union",
		"Difference",
		"SymmetricDifference",
	} {
		fmt.Println()
		fmt.Printf("**Operation:** %v\n", op)
		fmt.Println()
		fmt.Println("| Input Size | Simple Features | GEOS | Ratio |")
		fmt.Println("| ---        | ---             | ---  | ---   |")

		for i := 2; i <= 14; i++ {
			n := 1 << i
			sf := benches[benchmark{false, n, op}]
			ge := benches[benchmark{true, n, op}]
			fmt.Printf(
				"| 2<sup>%d</sup> | %s | %s | %.1f |\n",
				i, roundDuration(sf), roundDuration(ge),
				float64(sf)/float64(ge),
			)
		}
	}
}

func roundDuration(d time.Duration) time.Duration {
	var round time.Duration = time.Nanosecond
	for round < d {
		round *= 10
	}
	round /= 1000
	return d.Round(round)
}
