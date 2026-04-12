# Decisions

## Milestone 0

- The executable is named `sbl` and starts with two subcommands: `build` and `validate`.
- The first release is post-focused. `content/pages/` stays out of scope until the post pipeline is stable.
- Embedded default templates, CSS, and a base SWS config ship inside the binary so a minimal site root can build without extra theme files.
- `base_url` is required for `build` because feeds and sitemaps need canonical absolute URLs. `validate` can run without it.
- Test fixtures live under `testdata/` and are copied to temporary directories during tests so the builder can freely write `public/` and `deploy/sws.toml`.

## Milestone 1

- Slugs are derived from `content/posts/<slug>/` and must match `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- Validation accumulates multiple content errors in one pass instead of stopping at the first failure.
- Draft posts are excluded by default from the published graph. Links to excluded draft content are treated as broken unless drafts are explicitly included.

## Milestone 2

- Markdown rendering uses Goldmark with auto heading IDs and Chroma-backed syntax highlighting.
- Raw HTML from post source is not trusted. Trusted HTML is injected only by the build pipeline after Markdown rendering.
- Template overrides are optional. Embedded templates are the fallback and define the baseline site layout.

## Milestone 3

- Mermaid rendering uses a built-in SVG renderer in v1 so builds and tests do not depend on external tools.
- Math rendering in v1 converts supported delimiters into semantic HTML wrappers instead of requiring KaTeX at build time.
- Inline math is extracted before Markdown rendering because Goldmark consumes `\(` and `\)` escapes before an HTML-stage math pass can see them.
- All cacheable assets use fingerprinted filenames, including post-local files and generated Mermaid SVGs.

## Milestone 4

- The generated site always uses directory-style HTML routes such as `/posts/<slug>/index.html`.
- RSS, sitemap, robots, and error pages are generated on every build, even for small sites.

## Milestone 5

- `deploy/sws.toml` is generated from a base config plus alias redirects so redirect rules are not hand-maintained.
- Redirect output is deterministic: alias redirects are sorted before being written.
- Custom `--out` directories patch the generated SWS `root` setting so preview config stays aligned with the build output location.

## Documentation

- The top-level README stays brief and focuses on operator-facing usage: what `sbl` builds, required inputs, CLI flags, and the normal validate/build/serve workflow.

## Next Pass

- Standalone pages under `content/pages/<slug>/index.md` are now part of the supported site contract.
- Pages share the Markdown, Mermaid, math, and asset pipeline with posts, but they are excluded from the homepage post list, archive, and RSS feed.
- Pages require `title`; `summary` is optional and is used as page lead text when present.
- Page-local assets are fingerprinted under `/assets/pages/<slug>/...`.
- Real KaTeX rendering uses the local Node toolchain plus the repo's `package.json` dependencies at build time, instead of a pure-Go math renderer.
- KaTeX CSS and fonts are copied into `/assets/vendor/katex-<version>/...` without hashing because the versioned directory already provides stable cache-busting.
- Real Mermaid rendering uses `mermaid-isomorphic` plus Playwright Chromium at build time, instead of the previous SVG source-text placeholder.

## Setup Command

- The CLI now includes `sbl setup` to bootstrap Node dependencies and the Playwright Chromium browser in the repo root.
- `sbl setup` supports partial runs with `--skip-npm` and `--skip-browser` so CI or local environments can reuse cached prerequisites.

## Renderer Errors

- KaTeX and Mermaid failures now report both the source file path and the specific block index, so authors can find invalid content without guessing which page failed.
