# sbl Implementation Plan

## 1. Current repo reality

This repository does not contain an implementation yet. It currently has:

- reference docs in `blog-starter-docs/`
- one sample content tree in `test_blog/`
- no Go module
- no templates
- no static assets
- no site config

One important repo-specific issue: the current sample post in `test_blog/` is not buildable if we follow the validation rules from the docs. It references:

- `assets/layout.png`, which does not exist
- `/posts/another-post/`, which does not exist

That means the first implementation pass needs separate valid and invalid fixtures instead of using the current sample as the only happy path.

## 2. Product decision for v1

Build a single Go executable named `sbl`.

Primary command:

```text
sbl build <site-root> [--out <dir>] [--base-url <url>] [--include-drafts] [--clean]
```

Recommended additional command:

```text
sbl validate <site-root> [--base-url <url>] [--include-drafts]
```

`<site-root>` is the folder passed by the user. That folder is treated as the blog source root.

### v1 scope

Deliver these in the first real version:

- post discovery from `content/posts/<slug>/index.md`
- YAML front matter parsing
- strict validation
- Markdown to HTML rendering
- syntax-highlighted code blocks
- Mermaid support
- math support
- fingerprinted asset copying
- generated homepage
- generated archive page
- generated post pages
- generated `404.html` and `50x.html`
- generated `feed.xml`, `sitemap.xml`, and `robots.txt`
- generated SWS config from a checked-in base file plus alias redirects

### explicitly out of scope for the first pass

Defer these until the first end-to-end build is stable:

- watch mode
- built-in preview server
- authoring UI
- standalone `content/pages/`
- taxonomy pages
- search index

`pages/` should stay in the data model design, but not block the first implementation.

## 3. Site-root contract

The builder should accept a minimal folder and still work. To make that practical, `sbl` should ship embedded defaults for templates and baseline CSS.

Expected site-root layout:

```text
<site-root>/
  content/
    posts/
      <slug>/
        index.md
        assets/
  config/
    site.yaml               # optional but recommended
  templates/                # optional overrides
  static/                   # optional additional static files
  deploy/
    sws.base.toml           # optional override, else use embedded default
```

### default behavior

- if `templates/` is missing, use embedded templates
- if `static/` is missing, use embedded default CSS only
- if `deploy/sws.base.toml` is missing, use the embedded starter config
- if `config/site.yaml` is missing, use defaults for local builds except for `base_url`

### config requirement

`base_url` is required for correct feed and sitemap output. It should come from either:

- `config/site.yaml`, or
- `--base-url`

If neither is present, the build should fail with a clear message.

## 4. Concrete CLI behavior

### `sbl build`

Responsibilities:

1. resolve the site root and output directory
2. optionally clean the output directory
3. load config plus embedded defaults
4. discover posts
5. validate content and links
6. render content
7. write the `public/` tree
8. generate the final SWS config
9. print a short build summary

Exit code:

- `0` on success
- non-zero on any validation or rendering failure

### `sbl validate`

Responsibilities:

- run discovery and validation only
- print all failures in one pass if practical
- do not write output files

This command is useful for CI and for authors before a full build.

## 5. Recommended Go package layout

Keep the package split small enough to move fast.

```text
cmd/
  sbl/
    main.go

internal/
  app/
    build.go
    validate.go
  site/
    config.go
    load.go
  content/
    discover.go
    frontmatter.go
    model.go
    validate.go
    links.go
  render/
    markdown.go
    mermaid.go
    math.go
    templates.go
  assets/
    fingerprint.go
    copy.go
  output/
    writer.go
    pages.go
    feed.go
    sitemap.go
    robots.go
  sws/
    base.go
    generate.go

embedded/
  templates/
  static/
  deploy/
```

This is enough separation for testability without turning the repo into a framework.

## 6. Dependency choices

Use a small dependency set.

### Go libraries

- Markdown: `github.com/yuin/goldmark`
- syntax highlighting: `github.com/alecthomas/chroma/v2`
- YAML front matter: `gopkg.in/yaml.v3`
- HTML parsing for post-processing: `golang.org/x/net/html`

### standard library usage

Prefer the standard library for:

- templates
- XML generation
- file walking
- hashing
- path handling
- URL parsing

### Mermaid and math

These are the only parts that should not be forced into pure Go on day one.

Recommended approach:

- keep the core pipeline in Go
- define a renderer adapter interface for Mermaid and math
- back the adapters with external executables in v1
- fail fast with a clear error if the tool is required by content but not installed

Reason: pure-Go Mermaid and KaTeX rendering is not the best place to spend implementation time for the first usable release.

## 7. Data model

Use a simple internal model.

### `SiteConfig`

Fields:

- `Title`
- `BaseURL`
- `Description`
- `Language`
- `Author`
- `Navigation`

### `Post`

Fields:

- `Slug`
- `SourceDir`
- `SourcePath`
- `Title`
- `Date`
- `Updated`
- `Summary`
- `Draft`
- `Tags`
- `Aliases`
- `Description`
- `Image`
- `MarkdownBody`
- `LocalAssetRefs`
- `InternalLinks`

### `RenderedPost`

Fields:

- `Post`
- `CanonicalURL`
- `OutputDir`
- `BodyHTML`
- `TOC`
- `ReadingTime`
- `FingerprintAssets`
- `DiagramAssets`

## 8. Exact build pipeline

This should be the real execution order for `sbl build`.

1. Load config and defaults.
2. Discover all `content/posts/*/index.md` files.
3. Parse front matter and raw Markdown body.
4. Build the full post graph before rendering anything.
5. Validate slugs, dates, aliases, required fields, and duplicate URLs.
6. Validate local asset references and internal links.
7. For each post, extract Mermaid blocks and replace them with stable placeholders.
8. Render the remaining Markdown to HTML with Goldmark.
9. Apply syntax highlighting to non-Mermaid fenced code blocks.
10. Parse the rendered HTML fragment and render math only in text nodes outside `pre` and `code`.
11. Render Mermaid blocks to SVG assets and replace placeholders with `<img>` or `<figure>` markup that points to the fingerprinted SVG files.
12. Copy post-local assets into hashed paths under `public/assets/posts/<slug>/`.
13. Copy shared static assets into hashed paths under `public/assets/`.
14. Render templates for home, archive, each post, 404, and 50x.
15. Write `feed.xml`, `sitemap.xml`, and `robots.txt`.
16. Generate `deploy/sws.toml` from the base config plus alias redirects.

### why this order

- Mermaid must not go through normal code highlighting
- link validation must happen before writing output
- math rendering must not touch code samples
- all asset URLs must be fingerprinted before templates are finalized

## 9. Validation rules to implement first

Hard failures:

- missing `title`, `date`, or `summary`
- invalid `date` or `updated`
- `updated < date`
- duplicate slug
- duplicate canonical URL
- alias collision with another alias or canonical URL
- broken local asset path
- broken internal link to another post
- Mermaid rendering failure
- math rendering failure

Warnings only:

- missing `updated`
- code block without language
- long summary

The validator should accumulate as many errors as possible in one run so authors do not fix them one at a time.

## 10. Output contract

Target output tree for the first implementation:

```text
public/
  index.html
  404.html
  50x.html
  feed.xml
  sitemap.xml
  robots.txt

  archive/
    index.html

  posts/
    <slug>/
      index.html

  assets/
    site.<hash>.css
    posts/
      <slug>/
        <asset-name>.<hash>.<ext>
        diagram-1.<hash>.svg
```

Do not generate flat HTML files like `posts/<slug>.html`.

## 11. Template strategy

Use `html/template` and ship embedded defaults so the first build works without theme work.

Required default templates:

- `base.html`
- `index.html`
- `archive.html`
- `post.html`
- `404.html`
- `50x.html`

Template inputs should be precomputed in Go. Avoid putting logic in templates beyond loops and small conditionals.

## 12. SWS integration plan

Check in a base config and generate the final config during the build.

### build inputs

- embedded default `sws.base.toml`
- or `<site-root>/deploy/sws.base.toml` when present

### generated output

- `<site-root>/deploy/sws.toml`

### generated redirect rules

Always include:

- `/index.html` -> `/`
- `/archive/index.html` -> `/archive/`
- `/posts/{*}/index.html` -> `/posts/$1/`

Generate additional redirects from front matter `aliases`.

Fail the build on any redirect collision.

## 13. Test strategy

Use `testdata/` instead of the current single `test_blog/` tree.

Recommended fixture split:

```text
testdata/
  site-basic/
  site-rich-content/
  site-invalid-missing-asset/
  site-invalid-broken-link/
  site-invalid-duplicate-slug/
```

### test layers

Unit tests:

- front matter parsing
- slug and URL validation
- alias collision detection
- fingerprint naming
- SWS redirect generation

Golden tests:

- rendered post HTML fragment
- homepage HTML
- archive HTML
- feed XML

Integration tests:

- run `sbl build` against `site-basic`
- assert required output files exist
- assert fingerprinted assets are referenced from HTML
- assert `deploy/sws.toml` contains structural redirects and alias redirects

Windows path handling needs explicit test coverage because the current environment is Windows while the generated URLs must always use forward slashes.

## 14. Implementation milestones

### Milestone 0: bootstrap

Deliver:

- `go.mod`
- `cmd/sbl/main.go`
- embedded default templates and CSS
- basic config loader
- initial `testdata/site-basic`

Done when:

- `sbl build testdata/site-basic --base-url https://example.test` writes a minimal home page and one post page

### Milestone 1: discovery and validation

Deliver:

- post discovery
- front matter parsing
- validation engine
- split valid and invalid fixtures

Done when:

- invalid fixtures fail with actionable messages
- valid fixture passes `sbl validate`

### Milestone 2: Markdown rendering

Deliver:

- Goldmark renderer
- code highlighting
- heading IDs
- archive and post templates

Done when:

- sample posts render correctly with highlighted code and stable headings

### Milestone 3: rich content and assets

Deliver:

- Mermaid adapter
- math adapter
- asset fingerprinting
- asset URL rewriting

Done when:

- a post with code, Mermaid, math, and local images builds successfully

### Milestone 4: output completeness

Deliver:

- home page
- archive page
- 404
- 50x
- feed
- sitemap
- robots

Done when:

- `public/` matches the documented shape for the basic fixture

### Milestone 5: SWS integration and hardening

Deliver:

- base SWS config handling
- generated redirect rules
- build summary output
- integration and golden tests

Done when:

- the generated site can be previewed with Static Web Server using the generated `deploy/sws.toml`

## 15. First concrete tasks

If implementation starts now, do this in order:

1. Create the Go module and the `cmd/sbl` CLI skeleton.
2. Add embedded default templates and a tiny default `site.css`.
3. Create `testdata/site-basic/` with one valid post and a minimal `config/site.yaml`.
4. Move the current `test_blog` sample into invalid fixtures or fix it so it is intentionally valid.
5. Implement discovery plus front matter parsing.
6. Implement validation before touching HTML rendering.
7. Implement Markdown rendering and one post template.
8. Add asset fingerprinting.
9. Add Mermaid and math adapters.
10. Add feed, sitemap, robots, 404, 50x, and SWS config generation.

## 16. Risks and decisions to settle early

### Mermaid and math runtime dependencies

Decision needed:

- external tools are acceptable for v1, or
- the project must be pure Go even if feature work slows down

Recommendation: use external renderers in v1.

### fixture policy

Decision needed:

- keep `test_blog/` as an example site only, or
- convert everything into `testdata/` fixtures for automated tests

Recommendation: use `testdata/` for fixtures and keep a separate example site only if needed.

### pages support

Decision needed:

- ship `content/pages/` in the first release, or
- keep the first release post-only

Recommendation: first release should be post-only.

## 17. Definition of done

The first implementation is done when all of the following are true:

- `sbl build <site-root>` succeeds on a valid sample site
- invalid content fails with clear validation messages
- the generated site can be served as static files only
- generated HTML routes use directory-style URLs
- assets are fingerprinted
- a post with code, Mermaid, math, and local assets renders correctly
- `feed.xml`, `sitemap.xml`, `robots.txt`, `404.html`, and `50x.html` are generated
- `deploy/sws.toml` is generated with structural redirects and alias redirects
