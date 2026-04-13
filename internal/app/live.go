package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"time"
)

const (
	defaultLivePollInterval = 500 * time.Millisecond
	defaultLiveDebounce     = 200 * time.Millisecond
)

type LiveOptions struct {
	SiteRoot      string
	OutputDir     string
	BaseURL       string
	IncludeDrafts bool
	Context       context.Context
	Stdout        io.Writer
	Stderr        io.Writer
	PollInterval  time.Duration
	Debounce      time.Duration
}

type liveSnapshot map[string]liveEntry

type liveEntry struct {
	Exists  bool
	IsDir   bool
	Size    int64
	ModTime int64
}

type liveWatchTarget struct {
	RelPath   string
	Recursive bool
}

var liveWatchTargets = []liveWatchTarget{
	{RelPath: "config", Recursive: true},
	{RelPath: "content", Recursive: true},
	{RelPath: "templates", Recursive: true},
	{RelPath: "static", Recursive: true},
	{RelPath: filepath.ToSlash(filepath.Join("deploy", "sws.base.toml"))},
}

func Live(opts LiveOptions) error {
	siteRoot, err := filepath.Abs(opts.SiteRoot)
	if err != nil {
		return err
	}
	if err := validateLiveSiteRoot(siteRoot); err != nil {
		return err
	}

	outputDir := resolveOutputDir(siteRoot, opts.OutputDir)
	if err := validateOutputDir(siteRoot, outputDir); err != nil {
		return err
	}

	ctx := opts.Context
	stopSignals := func() {}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
		stopSignals = cancel
	}
	defer stopSignals()

	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = stdout
	}

	pollInterval := opts.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultLivePollInterval
	}
	debounce := opts.Debounce
	if debounce <= 0 {
		debounce = defaultLiveDebounce
	}

	snapshot, err := scanLiveInputs(siteRoot)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "watching %s\n", siteRoot)
	runLiveBuild(stdout, stderr, BuildOptions{
		SiteRoot:      siteRoot,
		OutputDir:     outputDir,
		BaseURL:       opts.BaseURL,
		IncludeDrafts: opts.IncludeDrafts,
		Stdout:        stdout,
	})

	lastSeen := snapshot
	var pending liveSnapshot
	var pendingSince time.Time

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			current, err := scanLiveInputs(siteRoot)
			if err != nil {
				fmt.Fprintf(stderr, "live watch scan failed: %v\n", err)
				continue
			}

			if snapshotsEqual(current, lastSeen) {
				pending = nil
				pendingSince = time.Time{}
				continue
			}

			if pending == nil || !snapshotsEqual(current, pending) {
				pending = current
				pendingSince = time.Now()
				continue
			}

			if time.Since(pendingSince) < debounce {
				continue
			}

			fmt.Fprintln(stdout, "changes detected, rebuilding...")
			runLiveBuild(stdout, stderr, BuildOptions{
				SiteRoot:      siteRoot,
				OutputDir:     outputDir,
				BaseURL:       opts.BaseURL,
				IncludeDrafts: opts.IncludeDrafts,
				Stdout:        stdout,
			})
			lastSeen = pending
			pending = nil
			pendingSince = time.Time{}
		}
	}
}

func validateLiveSiteRoot(siteRoot string) error {
	info, err := os.Stat(siteRoot)
	if os.IsNotExist(err) {
		return fmt.Errorf("site root does not exist: %s", siteRoot)
	}
	if err != nil {
		return fmt.Errorf("stat site root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("site root is not a directory: %s", siteRoot)
	}
	return nil
}

func runLiveBuild(stdout, stderr io.Writer, opts BuildOptions) {
	if err := Build(opts); err != nil {
		fmt.Fprintf(stderr, "live build failed: %v\n", err)
	}
}

func scanLiveInputs(siteRoot string) (liveSnapshot, error) {
	snapshot := make(liveSnapshot)
	for _, target := range liveWatchTargets {
		if err := scanLiveTarget(siteRoot, target, snapshot); err != nil {
			return nil, err
		}
	}
	return snapshot, nil
}

func scanLiveTarget(siteRoot string, target liveWatchTarget, snapshot liveSnapshot) error {
	absPath := filepath.Join(siteRoot, filepath.FromSlash(target.RelPath))
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		snapshot[target.RelPath] = liveEntry{}
		return nil
	}
	if err != nil {
		return err
	}

	recordLiveEntry(siteRoot, absPath, info, snapshot)
	if !target.Recursive || !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(absPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == absPath {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		recordLiveEntry(siteRoot, path, info, snapshot)
		return nil
	})
}

func recordLiveEntry(siteRoot, fullPath string, info os.FileInfo, snapshot liveSnapshot) {
	rel, err := filepath.Rel(siteRoot, fullPath)
	if err != nil {
		return
	}
	snapshot[filepath.ToSlash(rel)] = liveEntry{
		Exists:  true,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime().UnixNano(),
	}
}

func snapshotsEqual(left, right liveSnapshot) bool {
	return reflect.DeepEqual(left, right)
}
