package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Problem struct {
	question string
	answer   string
}

func shuffleProblems(probs []Problem) {
	n := len(probs)
	rand.Shuffle(n, func(i, j int) {
		probs[i], probs[j] = probs[j], probs[i]
	})
}

func parseRows(records [][]string) []Problem {
	problems := make([]Problem, len(records))
	for i, r := range records {
		problems[i] = Problem{question: r[0], answer: strings.TrimSpace(r[1])}
	}
	return problems
}

func main() {
	file := flag.String("csv", "problems.csv", "csv file with questions & answers")
	limit := flag.Int("limit", 30, "time limit (seconds)")

	flag.Parse()

	f, err := os.Open(*file)

	if err != nil {
		if pathErr, ok := errors.AsType[*os.PathError](err); ok {
			fmt.Fprintf(os.Stderr, "file not found: %s\n", pathErr.Path)
		} else {
			fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		}
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing csv file: %v\n", err)
		return
	}

	problems := parseRows(rows)
	answerCh := make(chan string)
	nCorrect := 0
	timer := time.NewTimer(time.Duration(*limit) * time.Second)

	// timer := time.NewTimer(limit * time.Second)
	handleQuestion := func(n int, q string) {
		var resp string
		fmt.Printf("Problem #%d %s = ", n, q)
		_, err := fmt.Scanln(&resp)
		if err != nil {
			fmt.Printf("error reading input: %v", err)
		}
		answerCh <- resp
	}

loop:
	for i, prob := range problems {
		go handleQuestion(i, prob.question)
		select {
		case resp := <-answerCh:
			if resp == prob.answer {
				nCorrect++
			}
		case <-timer.C:
			fmt.Println("\ntime's up")
			break loop

		}
	}
	fmt.Printf("Answered %d / %d correct", nCorrect, len(problems))
}
