package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"sbl/internal/app"
	"sbl/internal/console"
)

var (
	version              = "0.0.5"
	buildFn              = app.Build
	liveFn               = app.Live
	validateFn           = app.Validate
	stdout     io.Writer = os.Stdout
	stderr     io.Writer = os.Stderr
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "-v", "--version":
		return runVersion(args[1:])
	case "build":
		return runBuild(args[1:])
	case "live":
		return runLive(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "version":
		return runVersion(args[1:])
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	default:
		console.Errorf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runBuild(args []string) int {
	flags := newCommandFlagSet("build")
	outDir := flags.String("out", "", "output directory")
	baseURL := flags.String("base-url", "", "site base URL override")
	includeDrafts := flags.Bool("include-drafts", false, "include draft posts")
	clean := flags.Bool("clean", false, "remove output directory before build")
	timings := flags.Bool("timings", false, "print execution timings")
	args, err := normalizeArgs(args, map[string]struct{}{
		"out":      {},
		"base-url": {},
	})
	if err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 2
	}
	if ok, code := parseCommandFlags(flags, args); !ok {
		return code
	}
	if flags.NArg() != 1 {
		console.Errorf(stderr, "build requires exactly one site-root argument\n")
		return 2
	}

	if err := buildFn(app.BuildOptions{
		SiteRoot:      flags.Arg(0),
		OutputDir:     *outDir,
		BaseURL:       *baseURL,
		IncludeDrafts: *includeDrafts,
		Clean:         *clean,
		Stdout:        stdout,
		Timings:       *timings,
	}); err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 1
	}
	return 0
}

func runValidate(args []string) int {
	flags := newCommandFlagSet("validate")
	baseURL := flags.String("base-url", "", "site base URL override")
	includeDrafts := flags.Bool("include-drafts", false, "include draft posts")
	timings := flags.Bool("timings", false, "print execution timings")
	args, err := normalizeArgs(args, map[string]struct{}{
		"base-url": {},
	})
	if err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 2
	}
	if ok, code := parseCommandFlags(flags, args); !ok {
		return code
	}
	if flags.NArg() != 1 {
		console.Errorf(stderr, "validate requires exactly one site-root argument\n")
		return 2
	}

	if err := validateFn(app.ValidateOptions{
		SiteRoot:      flags.Arg(0),
		BaseURL:       *baseURL,
		IncludeDrafts: *includeDrafts,
		Stdout:        stdout,
		Timings:       *timings,
	}); err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 1
	}
	return 0
}

func runLive(args []string) int {
	flags := newCommandFlagSet("live")
	outDir := flags.String("out", "", "output directory")
	baseURL := flags.String("base-url", "", "site base URL override")
	includeDrafts := flags.Bool("include-drafts", false, "include draft posts")
	timings := flags.Bool("timings", false, "print execution timings")
	args, err := normalizeArgs(args, map[string]struct{}{
		"out":      {},
		"base-url": {},
	})
	if err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 2
	}
	if ok, code := parseCommandFlags(flags, args); !ok {
		return code
	}
	if flags.NArg() != 1 {
		console.Errorf(stderr, "live requires exactly one site-root argument\n")
		return 2
	}

	if err := liveFn(app.LiveOptions{
		SiteRoot:      flags.Arg(0),
		OutputDir:     *outDir,
		BaseURL:       *baseURL,
		IncludeDrafts: *includeDrafts,
		Timings:       *timings,
		Stdout:        stdout,
		Stderr:        stderr,
	}); err != nil {
		console.Errorf(stderr, "%v\n", err)
		return 1
	}
	return 0
}

func runVersion(args []string) int {
	flags := newCommandFlagSet("version")
	timings := flags.Bool("timings", false, "print execution timings")
	start := time.Now()
	if ok, code := parseCommandFlags(flags, args); !ok {
		return code
	}
	if flags.NArg() != 0 {
		console.Errorf(stderr, "version does not accept arguments\n")
		return 2
	}

	fmt.Fprintln(stdout, version)
	if *timings {
		fmt.Fprintf(stdout, "timings:\n  total: %s\n", time.Since(start).Round(time.Microsecond))
	}
	return 0
}

func newCommandFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

func parseCommandFlags(flags *flag.FlagSet, args []string) (bool, int) {
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(stdout)
			flags.Usage()
			return false, 0
		}
		console.Errorf(stderr, "%v\n", err)
		return false, 2
	}
	return true, 0
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  sbl [--version]")
	fmt.Fprintln(out, "  sbl [--help]")
	fmt.Fprintln(out, "  sbl build <site-root> [--out <dir>] [--base-url <url>] [--include-drafts] [--clean] [--timings]")
	fmt.Fprintln(out, "  sbl live <site-root> [--out <dir>] [--base-url <url>] [--include-drafts] [--timings]")
	fmt.Fprintln(out, "  sbl validate <site-root> [--base-url <url>] [--include-drafts] [--timings]")
	fmt.Fprintln(out, "  sbl version [--timings]")
}

func normalizeArgs(args []string, valueFlags map[string]struct{}) ([]string, error) {
	reordered := make([]string, 0, len(args))
	positionals := make([]string, 0, 1)

	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			positionals = append(positionals, args[index+1:]...)
			break
		}
		if !isFlagToken(arg) {
			positionals = append(positionals, arg)
			continue
		}

		reordered = append(reordered, arg)
		if !flagNeedsValue(arg, valueFlags) {
			continue
		}
		if index+1 >= len(args) {
			return nil, fmt.Errorf("flag %q requires a value", arg)
		}
		index++
		reordered = append(reordered, args[index])
	}

	if len(positionals) > 1 {
		return nil, fmt.Errorf("expected exactly one site-root argument")
	}

	return append(reordered, positionals...), nil
}

func isFlagToken(arg string) bool {
	return strings.HasPrefix(arg, "-") && arg != "-"
}

func flagNeedsValue(arg string, valueFlags map[string]struct{}) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	name := strings.TrimLeft(arg, "-")
	_, ok := valueFlags[name]
	return ok
}
