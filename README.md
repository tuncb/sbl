# sbl

`sbl` is a Go-based static blog builder.

It reads a site folder, discovers Markdown content under `content/posts/<slug>/index.md` and `content/pages/<slug>/index.md`, validates content, renders HTML, fingerprints assets, and writes a static site that can be served from `public/`.

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

## Options

`build`

- `--out <dir>`: write output to a custom directory instead of `<site-root>/public`
- `--base-url <url>`: override `config/site.yaml` `base_url`
- `--include-drafts`: include draft posts in validation and build output
- `--clean`: remove the output directory before building

`validate`

- `--base-url <url>`: override `config/site.yaml` `base_url`
- `--include-drafts`: include draft posts in validation

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

Real math rendering uses KaTeX through the local Node toolchain.

Run this once in the repo before building:

```text
npm install
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
