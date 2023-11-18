package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/sync/errgroup"
	"os/exec"
)

type RunOptions struct {
	RepositoryPath     string
	FilePath           string
	CommitsCount       uint
	StrategyName       string
	CaptureEvents      []ChangeEvent
	CaptureDevPackages bool
	OutputType         OutputType
	OutputPath         string
}

func Run(ctx context.Context, opts RunOptions) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errWg, ctx := errgroup.WithContext(ctx)

	var strat Strategy
	var err error
	if opts.StrategyName != "" {
		strat, err = findStrategyByName(opts.StrategyName)
		if err != nil {
			return fmt.Errorf("find strategy by name %s: %w", opts.StrategyName, err)
		}
	} else {
		strat, err = findStrategyByPath(opts.FilePath)
		if err != nil {
			return fmt.Errorf("find strategy by path %s: %w", opts.FilePath, err)
		}
	}

	printDebugf(`find pkg changes in %s using "%s" strategy`, opts.FilePath, strat.Name)

	repo, err := git.PlainOpen(opts.RepositoryPath)
	if err != nil {
		return fmt.Errorf("open git repository %s: %w", opts.RepositoryPath, err)
	}

	printDebugf("repository %s opened", opts.RepositoryPath)

	captureAdded, captureUpdated, captureDeleted := getCaptureOpts(opts.CaptureEvents)

	printDebugf("capture added:%t, updated:%t, deleted:%t", captureAdded, captureUpdated, captureDeleted)

	output, err := getPresenter(opts)
	if err != nil {
		return fmt.Errorf("get presenter: %w", err)
	}
	defer output.Close()

	printDebugf("output type: %s", opts.OutputType)
	if opts.OutputPath != "" {
		printDebugf("output path: %s", opts.OutputPath)
	}

	var prevPackages map[string]Package

	commitCh := make(chan string, min(int(opts.CommitsCount), 500))
	errWg.Go(func() error {
		if err := getLastCommits(ctx, commitCh, opts); err != nil {
			return fmt.Errorf("get last commits: %w", err)
		}
		return nil
	})

	for commit := range commitCh {
		hash := plumbing.NewHash(commit)

		ref, err := repo.CommitObject(hash)
		if err != nil {
			printErrf("get object of commit %s: %s", commit, err)
			continue
		}

		file, err := ref.File(opts.FilePath)
		if err != nil {
			printErrf("get file on commit %s: %s", commit, err)
			continue
		}

		reader, err := file.Reader()
		if err != nil {
			printErrf("open file on commit %s: %w", commit, err)
			continue
		}

		packages, err := strat.GetPackages(reader, opts.CaptureDevPackages)
		if err != nil {
			printErrf("get packages on commit %s: %w", commit, err)
			continue
		}
		if prevPackages == nil { // skip first commit
			prevPackages = packages
			continue
		}

		for pkgName := range mergePkgKeys(packages, prevPackages) {
			pkg, exist := packages[pkgName]
			pkgPrev, existPrev := prevPackages[pkgName]

			var err error
			if captureUpdated && exist && existPrev && pkg.Version != pkgPrev.Version {
				err = output.Write(PackageChange{
					Package:    pkg,
					OldVersion: pkgPrev.Version,
					Author:     ref.Author.Name,
					Email:      ref.Author.Email,
					Time:       ref.Author.When,
					Commit:     commit,
					Event:      PackageUpdated,
				})
			} else if captureAdded && exist && !existPrev {
				err = output.Write(PackageChange{
					Package: pkg,
					Author:  ref.Author.Name,
					Email:   ref.Author.Email,
					Time:    ref.Author.When,
					Commit:  commit,
					Event:   PackageAdded,
				})
			} else if captureDeleted && !exist && existPrev {
				err = output.Write(PackageChange{
					Package: pkgPrev,
					Author:  ref.Author.Name,
					Email:   ref.Author.Email,
					Time:    ref.Author.When,
					Commit:  commit,
					Event:   PackageDeleted,
				})
			}
			if err != nil {
				return fmt.Errorf("write result to output: %w", err)
			}
		}

		prevPackages = packages
	}

	if err = errWg.Wait(); err != nil {
		return err
	}

	if opts.OutputPath != "" {
		printDebugf("result was saved to %s", opts.OutputPath)
	}

	return nil
}

func getCaptureOpts(changeEvents []ChangeEvent) (bool, bool, bool) {
	captureAdded := false
	captureUpdated := false
	captureDeleted := false

	for _, event := range changeEvents {
		switch event {
		case PackageAdded:
			captureAdded = true
		case PackageUpdated:
			captureUpdated = true
		case PackageDeleted:
			captureDeleted = true
		}
	}

	if !captureAdded && !captureUpdated && !captureDeleted {
		captureAdded = true
		captureUpdated = true
		captureDeleted = true
	}

	return captureAdded, captureUpdated, captureDeleted
}

func mergePkgKeys(a, b map[string]Package) map[string]struct{} {
	keys := make(map[string]struct{}, 0)
	for pkg := range a {
		keys[pkg] = struct{}{}
	}
	for pkg := range b {
		keys[pkg] = struct{}{}
	}
	return keys
}

func getLastCommits(ctx context.Context, ch chan<- string, opts RunOptions) error {
	defer close(ch)

	app := "git"
	args := []string{"log", "-n", fmt.Sprintf("%d", opts.CommitsCount), "--pretty=format:%H", "--reverse", "--", opts.FilePath}

	cmd := exec.Command(app, args...)
	cmd.Dir = opts.RepositoryPath

	stdout, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("getting cmd output: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		commit := scanner.Text()
		ch <- commit
	}

	return nil
}
