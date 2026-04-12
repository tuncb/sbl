package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type katexJob struct {
	ID          string `json:"id"`
	Expression  string `json:"expression"`
	DisplayMode bool   `json:"displayMode"`
}

type katexResult struct {
	ID   string `json:"id"`
	HTML string `json:"html"`
}

const katexEvalScript = `
import katex from 'katex';
import { readFileSync } from 'node:fs';

const input = JSON.parse(readFileSync(0, 'utf8'));
const output = input.map((item) => ({
  id: item.id,
  html: katex.renderToString(item.expression, {
    displayMode: item.displayMode,
    throwOnError: true,
    output: 'htmlAndMathml',
    strict: 'error'
  })
}));
process.stdout.write(JSON.stringify(output));
`

func renderKaTeX(jobs []katexJob) (map[string]string, error) {
	if len(jobs) == 0 {
		return map[string]string{}, nil
	}

	payload, err := json.Marshal(jobs)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("node", "--input-type=module", "--eval", katexEvalScript)
	cmd.Dir = moduleRoot()
	cmd.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("render KaTeX with node: %w: %s", err, stderr.String())
	}

	var results []katexResult
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		return nil, fmt.Errorf("decode KaTeX output: %w", err)
	}

	out := make(map[string]string, len(results))
	for _, result := range results {
		out[result.ID] = result.HTML
	}
	return out, nil
}
