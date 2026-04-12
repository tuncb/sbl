package app

import (
	"fmt"
	"io"
	"os/exec"

	"sbl/internal/tooling"
)

type SetupOptions struct {
	SkipNPM     bool
	SkipBrowser bool
	Stdout      io.Writer
	Stderr      io.Writer
	Runner      CommandRunner
}

type CommandRunner func(name string, args []string, dir string, stdout, stderr io.Writer) error

func Setup(opts SetupOptions) error {
	runner := opts.Runner
	if runner == nil {
		runner = runCommand
	}
	root := tooling.ModuleRoot()

	if !opts.SkipNPM {
		if opts.Stdout != nil {
			fmt.Fprintln(opts.Stdout, "running npm install")
		}
		if err := runner("npm", []string{"install"}, root, opts.Stdout, opts.Stderr); err != nil {
			return fmt.Errorf("setup npm dependencies: %w", err)
		}
	}

	if !opts.SkipBrowser {
		if opts.Stdout != nil {
			fmt.Fprintln(opts.Stdout, "running npx playwright install chromium")
		}
		if err := runner("npx", []string{"playwright", "install", "chromium"}, root, opts.Stdout, opts.Stderr); err != nil {
			return fmt.Errorf("setup playwright browser: %w", err)
		}
	}

	if opts.Stdout != nil {
		fmt.Fprintf(opts.Stdout, "setup complete in %s\n", root)
	}
	return nil
}

func runCommand(name string, args []string, dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
