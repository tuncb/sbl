package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type mermaidJob struct {
	ID      string `json:"id"`
	Diagram string `json:"diagram"`
}

type mermaidResult struct {
	ID    string `json:"id"`
	SVG   string `json:"svg"`
	Error string `json:"error"`
}

const mermaidEvalScript = `
import { createMermaidRenderer } from 'mermaid-isomorphic';
import { readFileSync } from 'node:fs';

const input = JSON.parse(readFileSync(0, 'utf8'));
const renderer = createMermaidRenderer();
const rendered = await renderer(
  input.map((item) => item.diagram),
  {
    prefix: 'sbl',
    mermaidOptions: {
      fontFamily: 'Arial, sans-serif'
    }
  }
);

const output = rendered.map((item, index) => {
  if (item.status !== 'fulfilled') {
    return {
      id: input[index].id,
      error: item.reason instanceof Error ? item.reason.message : String(item.reason)
    };
  }
  return {
    id: input[index].id,
    svg: item.value.svg
  };
});

process.stdout.write(JSON.stringify(output));
`

func renderMermaidDiagrams(blocks []MermaidBlock) (map[string]mermaidResult, error) {
	jobs := make([]mermaidJob, 0, len(blocks))
	for _, block := range blocks {
		if strings.TrimSpace(block.Source) == "" {
			return nil, fmt.Errorf("render Mermaid block %d: Mermaid diagram is empty", block.Index)
		}
		jobs = append(jobs, mermaidJob{
			ID:      block.Placeholder,
			Diagram: block.Source,
		})
	}

	payload, err := json.Marshal(jobs)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("node", "--input-type=module", "--eval", mermaidEvalScript)
	cmd.Dir = moduleRoot()
	cmd.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("render Mermaid with node: %w: %s", err, stderr.String())
	}

	var results []mermaidResult
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		return nil, fmt.Errorf("decode Mermaid output: %w", err)
	}

	out := make(map[string]mermaidResult, len(results))
	for _, result := range results {
		out[result.ID] = result
	}
	return out, nil
}
