package app_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"sbl/internal/app"
	"sbl/internal/tooling"
)

func TestSetupRunsNPMThenBrowserInstall(t *testing.T) {
	var calls []string
	var stdout bytes.Buffer

	err := app.Setup(app.SetupOptions{
		Stdout: &stdout,
		Runner: func(name string, args []string, dir string, out, errOut io.Writer) error {
			calls = append(calls, name+" "+strings.Join(args, " ")+" @ "+dir)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}

	wantDir := tooling.ModuleRoot()
	want := []string{
		"npm install @ " + wantDir,
		"npx playwright install chromium @ " + wantDir,
	}
	if len(calls) != len(want) {
		t.Fatalf("unexpected call count: got %d want %d", len(calls), len(want))
	}
	for index := range want {
		if calls[index] != want[index] {
			t.Fatalf("unexpected call %d: got %q want %q", index, calls[index], want[index])
		}
	}
	if !strings.Contains(stdout.String(), "setup complete") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestSetupSupportsSkipFlags(t *testing.T) {
	tests := []struct {
		name        string
		opts        app.SetupOptions
		wantCommand string
		wantCalls   int
	}{
		{
			name:        "skip npm",
			opts:        app.SetupOptions{SkipNPM: true},
			wantCommand: "npx playwright install chromium",
			wantCalls:   1,
		},
		{
			name:        "skip browser",
			opts:        app.SetupOptions{SkipBrowser: true},
			wantCommand: "npm install",
			wantCalls:   1,
		},
		{
			name:      "skip both",
			opts:      app.SetupOptions{SkipNPM: true, SkipBrowser: true},
			wantCalls: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var calls []string
			err := app.Setup(app.SetupOptions{
				SkipNPM:     test.opts.SkipNPM,
				SkipBrowser: test.opts.SkipBrowser,
				Runner: func(name string, args []string, dir string, out, errOut io.Writer) error {
					calls = append(calls, name+" "+strings.Join(args, " "))
					return nil
				},
			})
			if err != nil {
				t.Fatalf("Setup returned error: %v", err)
			}
			if len(calls) != test.wantCalls {
				t.Fatalf("unexpected call count: got %d want %d", len(calls), test.wantCalls)
			}
			if test.wantCalls == 1 && calls[0] != test.wantCommand {
				t.Fatalf("unexpected command: got %q want %q", calls[0], test.wantCommand)
			}
		})
	}
}
