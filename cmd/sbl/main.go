package main

import (
	"flag"
	"fmt"
	"os"

	"sbl/internal/app"
)

var (
	buildFn    = app.Build
	validateFn = app.Validate
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	switch args[0] {
	case "build":
		return runBuild(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printUsage(os.Stderr)
		return 2
	}
}

func runBuild(args []string) int {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	outDir := flags.String("out", "", "output directory")
	baseURL := flags.String("base-url", "", "site base URL override")
	includeDrafts := flags.Bool("include-drafts", false, "include draft posts")
	clean := flags.Bool("clean", false, "remove output directory before build")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "build requires exactly one site-root argument")
		return 2
	}

	if err := buildFn(app.BuildOptions{
		SiteRoot:      flags.Arg(0),
		OutputDir:     *outDir,
		BaseURL:       *baseURL,
		IncludeDrafts: *includeDrafts,
		Clean:         *clean,
		Stdout:        os.Stdout,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runValidate(args []string) int {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	baseURL := flags.String("base-url", "", "site base URL override")
	includeDrafts := flags.Bool("include-drafts", false, "include draft posts")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "validate requires exactly one site-root argument")
		return 2
	}

	if err := validateFn(app.ValidateOptions{
		SiteRoot:      flags.Arg(0),
		BaseURL:       *baseURL,
		IncludeDrafts: *includeDrafts,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  sbl build <site-root> [--out <dir>] [--base-url <url>] [--include-drafts] [--clean]")
	fmt.Fprintln(out, "  sbl validate <site-root> [--base-url <url>] [--include-drafts]")
}
