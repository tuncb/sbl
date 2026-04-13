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

- Mermaid rendering keeps source as safe HTML placeholders and uses self-hosted browser assets instead of a build-time renderer.
- Math rendering converts supported delimiters into semantic HTML wrappers and lets self-hosted KaTeX enhance them in the browser.
- Inline math is extracted before Markdown rendering because Goldmark consumes `\(` and `\)` escapes before an HTML-stage math pass can see them.
- All cacheable assets use fingerprinted filenames except versioned vendor assets under `/assets/vendor/...`.

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
- KaTeX and Mermaid ship as vendored browser assets committed in the repo and copied to `/assets/vendor/...` during build.
- Mermaid fences are emitted as safe `<pre>` placeholders and rendered client-side, so builds no longer require Node or a headless browser.
- Math blocks are emitted as semantic wrappers and rendered client-side with self-hosted KaTeX assets.
- Inline math is still extracted before Markdown rendering because Goldmark consumes `\(` and `\)` escapes before an HTML-stage math pass can see them.
- KaTeX CSS, fonts, and JS, plus Mermaid JS, live under versioned vendor directories without hashing because the versioned path already provides cache-busting.

## README Setup

- The README now includes a concrete quick-start that shows the exact folders, minimal config, sample content, and the build/preview commands for creating a new site.
