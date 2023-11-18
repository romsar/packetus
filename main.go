package main

import (
	"context"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
)

var repositoryPath, filePath string
var commitsCount uint
var strategyName string
var changeEvents []string
var captureDevPackages bool
var jsonResultPath, csvResultPath string

func init() {
	flag.UintVar(&commitsCount, "commits", 100, "processed commits count")
	flag.StringVar(&strategyName, "strategy", "", "strategy name: composer, npm")
	flag.StringSliceVar(
		&changeEvents,
		"change-events",
		[]string{"added", "updated", "deleted"},
		"change events that will be processed separated by comma: added, updated, deleted",
	)
	flag.BoolVar(&captureDevPackages, "dev", true, "capture dev packages")
	flag.StringVar(&jsonResultPath, "json", "", "json result path")
	flag.StringVar(&csvResultPath, "csv", "", "csv result path")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	args := os.Args[1:]
	if len(args) >= 1 {
		repositoryPath = args[0]
	}
	if repositoryPath == "" {
		fmt.Println("repository path argument is required")
		os.Exit(1)
	}
	if len(args) >= 2 {
		filePath = strings.TrimLeft(args[1], "./")
	}
	if filePath == "" {
		fmt.Println("file path argument is required")
		os.Exit(1)
	}

	outputType := OutputStdout
	outputPath := ""
	if csvResultPath != "" {
		outputPath = csvResultPath
		outputType = OutputCSV
	} else if jsonResultPath != "" {
		outputPath = jsonResultPath
		outputType = OutputJSON
	}

	captureEvents := make([]ChangeEvent, len(changeEvents))
	for i, event := range changeEvents {
		captureEvents[i] = ChangeEvent(event)
	}

	opts := RunOptions{
		RepositoryPath:     repositoryPath,
		FilePath:           filePath,
		CommitsCount:       commitsCount,
		StrategyName:       strategyName,
		CaptureEvents:      captureEvents,
		CaptureDevPackages: captureDevPackages,
		OutputType:         outputType,
		OutputPath:         outputPath,
	}

	if err := Run(ctx, opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
