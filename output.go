package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/jszwec/csvutil"
	"io"
	"os"
	"path/filepath"
)

type OutputType string

const (
	OutputStdout OutputType = "stdout"
	OutputCSV    OutputType = "csv"
	OutputJSON   OutputType = "json"
)

type Presenter interface {
	Write(pc PackageChange) error
	io.Closer
}

func getPresenter(opts RunOptions) (Presenter, error) {
	var file *os.File
	var err error
	if opts.OutputPath != "" {
		dir := filepath.Dir(opts.OutputPath)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", dir, err)
		}
		file, err = os.Create(opts.OutputPath)
		if err != nil {
			return nil, fmt.Errorf("create file %s: %w", opts.OutputPath, err)
		}
	}

	switch opts.OutputType {
	case OutputCSV:
		writer := csv.NewWriter(file)
		return &csvPresenter{
			writer: writer,
			closer: file,
		}, nil
	case OutputJSON:
		encoder := json.NewEncoder(file)
		return &jsonPresenter{
			encoder: encoder,
			closer:  file,
		}, nil
	case OutputStdout:
		fallthrough
	default:
		return stdoutPresenter{}, nil
	}
}

type stdoutPresenter struct{}

func (p stdoutPresenter) Write(pc PackageChange) error {
	authorName := pc.Author
	if authorName == "" {
		authorName = pc.Email
	} else if pc.Email != "" {
		authorName += " (" + pc.Email + ")"
	}

	eventName := ""
	versionStr := ""
	devPkgStr := ""
	if pc.Package.IsDev {
		devPkgStr = "DEV-"
	}

	var col color.Attribute

	switch pc.Event {
	case PackageAdded:
		eventName = "added"
		versionStr = fmt.Sprintf("Ver: %s", pc.Package.Version)
		col = color.FgGreen
	case PackageUpdated:
		eventName = "updated"
		versionStr = fmt.Sprintf("Ver: %s -> %s)", pc.OldVersion, pc.Package.Version)
		col = color.FgYellow
	case PackageDeleted:
		eventName = "deleted"
		versionStr = fmt.Sprintf("Last ver: %s", pc.Package.Version)
		col = color.FgRed
	}

	printFunc := color.New(col)
	printStr := "[%s] User %s %s %spackage %s. %s. Commit: %s\n"
	printArgs := []any{
		pc.Time.Format("2006-01-02 15:04:05"),
		authorName,
		eventName,
		devPkgStr,
		pc.Package.Name,
		versionStr,
		pc.Commit,
	}
	if _, err := printFunc.Printf(printStr, printArgs...); err != nil {
		fmt.Printf(printStr, printArgs...)
	}

	return nil
}

func (p stdoutPresenter) Close() error {
	return nil
}

type csvPresenter struct {
	writer  *csv.Writer
	changes []PackageChange
	closer  io.Closer
}

func (p *csvPresenter) Write(pc PackageChange) error {
	p.changes = append(p.changes, pc)
	return nil
}

func (p *csvPresenter) Close() error {
	encoder := csvutil.NewEncoder(p.writer)
	if err := encoder.Encode(p.changes); err != nil {
		return fmt.Errorf("csv encode: %w", err)
	}
	p.writer.Flush()
	return p.closer.Close()
}

type jsonPresenter struct {
	encoder *json.Encoder
	changes []PackageChange
	closer  io.Closer
}

func (p *jsonPresenter) Write(pc PackageChange) error {
	p.changes = append(p.changes, pc)
	return nil
}

func (p *jsonPresenter) Close() error {
	if err := p.encoder.Encode(p.changes); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return p.closer.Close()
}

func printErrf(format string, a ...any) {
	format += "\n"
	if _, err := color.New(color.FgHiRed).Printf(format, a...); err != nil {
		fmt.Printf(format, a...)
	}
}

func printDebugf(format string, a ...any) {
	format += "\n"
	if _, err := color.New(color.FgHiWhite).Printf(format, a...); err != nil {
		fmt.Printf(format, a...)
	}
}
