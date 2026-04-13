# sbl

`sbl` is a Go-based static blog builder.

It reads a site folder, discovers Markdown content under `content/posts/<slug>/index.md` and `content/pages/<slug>/index.md`, validates content, renders HTML, fingerprints assets, and writes a static site that can be served from `public/`.

The shipped `sbl` binary is self-contained. Default templates, CSS, client render bootstrap, vendored KaTeX and Mermaid assets, and the base SWS config are compiled into the executable. To build a site, only the site folder itself needs to be present on disk.

Generated output includes:

- home page
- archive page
- post pages
- standalone pages
- `404.html`
- `50x.html`
- `feed.xml`
- `sitemap.xml`
- `robots.txt`
- `deploy/sws.toml`

## Commands

Build a site:

```text
sbl build <site-root> [--out <dir>] [--base-url <url>] [--include-drafts] [--clean]
```

Validate content without writing output:

```text
sbl validate <site-root> [--base-url <url>] [--include-drafts]
```

Print the CLI version:

```text
sbl version
```

## Options

`build`

- `--out <dir>`: write output to a custom directory instead of `<site-root>/public`
- `--base-url <url>`: override `config/site.yaml` `base_url`
- `--include-drafts`: include draft posts in validation and build output
- `--clean`: remove the output directory before building

`validate`

- `--base-url <url>`: override `config/site.yaml` `base_url`
- `--include-drafts`: include draft posts in validation

`version`

- prints the current CLI version string

## Site Layout

Minimum expected input:

```text
<site-root>/
  config/
    site.yaml
  content/
    posts/
      <slug>/
        index.md
        assets/
    pages/
      <slug>/
        index.md
        assets/
```

`site.yaml` should define at least:

```yaml
title: "My Blog"
base_url: "https://example.com"
```

## Tooling

KaTeX and Mermaid ship as self-hosted browser assets committed in the repo.

Pages that contain math or Mermaid fences load those assets client-side from `/assets/vendor/...`.

Builds do not require Node, npm, or a browser install.
Builds do not require checked-out repo templates, static assets, or deploy defaults next to the executable.

## Quick Start

1. Build the `sbl` binary:

```text
go build ./cmd/sbl
```

2. Create a new site folder:

```text
my-site/
  config/
    site.yaml
  content/
    posts/
      hello-world/
        index.md
```

3. Add `config/site.yaml`:

```yaml
title: "My Blog"
base_url: "https://example.com"
description: "My static site built with sbl."
language: "en"
navigation:
  - label: "Archive"
    url: "/archive/"
```

4. Add `content/posts/hello-world/index.md`:

```md
---
title: "Hello World"
date: 2026-04-12
summary: "My first post."
---

## Welcome

This site was built with `sbl`.
```

5. Validate and build the site:

```text
./sbl validate ./my-site
./sbl build ./my-site --clean
```

Flags may appear before or after `<site-root>`.

6. The generated site will be written to:

```text
my-site/public/
```

7. Preview it with Static Web Server:

```text
static-web-server -w ./my-site/deploy/sws.toml
```

## Usage

1. Create a site folder with `config/site.yaml`, posts under `content/posts/`, and optional standalone pages under `content/pages/`.
2. Run:

```text
sbl validate ./my-site
sbl build ./my-site --clean
```

3. Serve the generated `public/` directory with any static file server.
4. If you use Static Web Server, use the generated `deploy/sws.toml`.

Example:

```text
static-web-server -w ./my-site/deploy/sws.toml
```

## Releases

CI runs on pushes to `main` and on pull requests targeting `main`.

Release publishing runs when you push a tag that starts with `v`, for example:

```text
git tag v0.0.1
git push origin v0.0.1
```

The release workflow only publishes if the tagged commit is reachable from `main`.

Each release uploads one compressed archive per target platform:

- Windows x64: `sbl-windows-x86_64-vX.Y.Z.zip`
- Linux x64: `sbl-linux-x86_64-vX.Y.Z.tar.gz`
- macOS arm64: `sbl-macos-arm64-vX.Y.Z.tar.gz`
