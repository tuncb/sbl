# Blog Implementation Plan (Static-Web-Server edition)

## Goal

Build a static-first blog where:

- content is written in Markdown
- Go performs all content processing and HTML generation
- the deployed site is a plain `public/` directory
- Static Web Server (SWS) serves the generated files in production
- the runtime has no custom application logic

The design rule is:

> keep all complexity in the build pipeline, and make production serving a file-hosting problem.

---

## 1. Architecture at a glance

### Source layer

- `content/posts/<slug>/index.md`
- `content/posts/<slug>/assets/...`
- `templates/...`
- `static/...`
- `config/site.toml` (or yaml/json if you prefer)

### Build layer (Go)

The Go build command is responsible for:

- discovering content
- parsing front matter
- validating metadata and links
- rendering Markdown to HTML
- highlighting code blocks
- rendering Mermaid diagrams to SVG assets
- rendering equations to HTML
- copying and fingerprinting assets
- applying HTML templates
- generating feeds, sitemap, robots, 404, and 50x pages
- generating redirect rules from content aliases
- writing `public/`
- writing the final `deploy/sws.toml`

### Runtime layer (SWS)

SWS is responsible for:

- serving files from `public/`
- applying cache headers
- serving custom `404.html` and `50x.html`
- doing directory trailing-slash redirects
- applying explicit redirect rules for old URLs
- compression and optional health checks

There is no custom HTTP routing code in production.

---

## 2. Canonical URL policy

Lock this down first.

### Canonical public URLs

- homepage: `/`
- archive: `/archive/`
- post: `/posts/<slug>/`
- standalone page: `/pages/<slug>/`
- feed: `/feed.xml`
- sitemap: `/sitemap.xml`
- robots: `/robots.txt`

### Rules

- all HTML pages except `/` use trailing-slash canonical URLs
- every canonical page URL maps to a real directory containing `index.html`
- do not use rewrites for normal page routing
- redirect explicit `index.html` requests to the directory URL
- generate redirects for old slugs from front matter

This keeps the on-disk layout and the public URL layout identical.

---

## 3. Proposed repository layout

```text
/cmd/
  /buildsite/
    main.go

/internal/
  /config/
  /content/
  /markdown/
  /render/
  /assets/
  /output/
  /feeds/
  /redirects/

/content/
  /posts/
    /my-post/
      index.md
      /assets/
        image.png

/templates/
  base.html
  index.html
  post.html
  archive.html
  page.html
  404.html
  50x.html
  partials/

/static/
  site.css
  icons/
  vendor/
    katex/

/deploy/
  sws.base.toml

/public/
  ...generated output...
```

---

## 4. Exact responsibilities by subsystem

## 4.1 Content discovery and validation

### Input model

Each post lives in one folder:

```text
content/posts/<slug>/
  index.md
  assets/
```

### Front matter fields for v1

Required:

- `title`
- `date`
- `summary`

Optional:

- `updated`
- `draft`
- `aliases`
- `tags`
- `description`
- `image`

### Validation rules

Implement these as hard build failures:

- folder name is the canonical slug
- slug must be URL-safe and unique
- required fields exist
- `updated >= date` when present
- aliases do not collide with current URLs
- local asset links resolve
- internal post/page links resolve
- no duplicate canonical URLs
- draft posts are excluded from production output

---

## 4.2 Markdown and rich content pipeline

### Order of operations

1. Parse front matter
2. Read raw Markdown body
3. Extract Mermaid fenced blocks
4. Render Mermaid blocks to SVG files under hashed asset paths
5. Replace Mermaid source blocks with placeholders referring to the generated SVG assets
6. Render Markdown to HTML with Goldmark
7. Highlight fenced code blocks
8. Render math delimiters to HTML
9. Sanitize or trust only the final post body fragment
10. Apply HTML templates

### Output policy

- code blocks stay inside the HTML page
- Mermaid diagrams become regular SVG files under `/assets/posts/<slug>/...`
- equations are rendered into the HTML page
- metadata remains escaped text
- only the final article body is treated as trusted HTML

### Why external SVG files for Mermaid

- simpler caching
- easier CSP
- fewer surprises from inline SVG styling
- easy to fingerprint for long-term caching

---

## 4.3 Asset pipeline

### Goals

- every cacheable asset path should be safe to cache for a year
- changing an asset should change its URL

### Rule

All copied assets get fingerprinted.

Examples:

```text
static/site.css                      -> public/assets/site.3f2c9b61.css
content/posts/go-http/assets/a.png   -> public/assets/posts/go-http/a.91d2f304.png
content/posts/go-http/assets/b.pdf   -> public/assets/posts/go-http/b.7d20b9ab.pdf
Mermaid diagram #1                   -> public/assets/posts/go-http/diagram-1.55cb0d6a.svg
```

### Consequences

- HTML pages are short-lived and revalidated
- `/assets/**` can be `immutable`
- no query-string cache busting is needed

---

## 4.4 Template system

### Required templates

- `base.html`
- `index.html`
- `post.html`
- `archive.html`
- `page.html`
- `404.html`
- `50x.html`

### Data passed to templates

Common site data:

- site title
- base URL
- nav links
- current page metadata
- canonical URL
- stylesheet URL
- feed URL

Post data:

- title
- date / updated
- summary
- rendered body HTML
- reading time
- previous/next links
- social image URL if present

---

## 4.5 Output generator

The output generator writes the exact `public/` tree.

It must:

- create real directories for page routes
- write `index.html` into each route directory
- write root files (`feed.xml`, `sitemap.xml`, `robots.txt`)
- write `404.html` and `50x.html` at the root
- copy hashed assets into `/assets/`
- generate the final SWS configuration file with redirects

---

## 5. Exact `public/` structure

```text
public/
  index.html
  404.html
  50x.html
  feed.xml
  sitemap.xml
  robots.txt
  favicon.ico
  favicon.svg
  site.webmanifest

  archive/
    index.html

  posts/
    <slug>/
      index.html

  pages/
    <slug>/
      index.html

  assets/
    site.<hash>.css
    posts/
      <slug>/
        <asset-name>.<hash>.<ext>
        diagram-1.<hash>.svg
        diagram-2.<hash>.svg
    vendor/
      katex-<version>/
        katex.min.css
        fonts/
          ...
    icons/
      ...
```

### Notes

- `posts/` and `pages/` contain only directories with `index.html`
- do not write `posts/<slug>.html`
- do not rely on rewrites to reach post pages
- all long-lived browser-cached files live under `/assets/`
- `feed.xml`, `sitemap.xml`, and `robots.txt` stay top-level for convention and discoverability

---

## 6. Redirect model

### Redirect sources

Redirects come from three places:

1. structural cleanup redirects
2. `aliases` front matter on content
3. optional host canonicalization (for example `www` -> bare domain)

### Structural redirects to generate

- `/index.html` -> `/`
- `/archive/index.html` -> `/archive/`
- `/posts/<slug>/index.html` -> `/posts/<slug>/`
- `/pages/<slug>/index.html` -> `/pages/<slug>/`

### Content alias redirects

Example front matter:

```yaml
aliases:
  - /2024/12/my-old-post/
  - /notes/old-slug/
```

The build should turn that into SWS redirect rules pointing to the canonical post URL.

### Collision rule

Fail the build if:

- an alias matches another canonical URL
- two aliases collide
- an alias points to a draft or missing page

---

## 7. SWS configuration strategy

Keep two files in mind:

- `deploy/sws.base.toml` checked into git
- generated `deploy/sws.toml` or `dist/sws.toml` written by the build

The build step should:

1. load the checked-in base config
2. append generated redirect rules from content aliases
3. write the final SWS config used for deployment

This avoids hand-editing redirect rules.

---

## 8. 404 and 50x pages

## 404 page requirements

- plain text-first design
- same base template as the rest of the site
- clear status message
- link back to home and archive
- `noindex` response header via SWS custom headers

Suggested copy:

- title: `Page not found`
- body: `The page you asked for does not exist or has moved.`
- links: `Home`, `Archive`

## 50x page requirements

- same style as 404
- no dynamic details
- plain retry advice

Suggested copy:

- title: `Temporary server problem`
- body: `Please try again in a moment.`

---

## 9. Phased implementation order

## Phase 1 — foundation

Deliver:

- Go module
- `cmd/buildsite`
- site config loader
- one template set
- one sample post
- one CSS file

Exit:

- build writes homepage + one post page + archive + 404

## Phase 2 — content model and validation

Deliver:

- front matter parser
- content discovery
- URL/slug validator
- asset/link validation
- draft filtering

Exit:

- invalid content fails clearly

## Phase 3 — Markdown rendering

Deliver:

- Goldmark renderer
- code highlighting
- heading IDs
- TOC extraction if desired

Exit:

- code samples render correctly

## Phase 4 — Mermaid and math

Deliver:

- Mermaid extraction
- SVG generation
- math rendering
- asset hashing + URL rewriting

Exit:

- diagrams and equations render correctly in built pages

## Phase 5 — output generator

Deliver:

- canonical page directories
- feed
- sitemap
- robots
- 404 + 50x
- favicon/webmanifest

Exit:

- `public/` matches the target output tree exactly

## Phase 6 — SWS deployment config

Deliver:

- checked-in base config
- generated final `sws.toml`
- redirect generation from aliases
- cache header policy

Exit:

- local deploy can be served with `static-web-server -w deploy/sws.toml`

## Phase 7 — tests and CI

Deliver:

- validation tests
- golden tests for rendered HTML
- redirect generation tests
- smoke test that checks required output files exist
- CI build that fails on content errors

Exit:

- build is reproducible and deployable

---

## 10. Developer workflow

### Local author workflow

1. edit Markdown
2. run build
3. preview with SWS
4. commit content + generated redirects if needed

### Useful commands

```text
make build
make preview
make clean
make test
```

Example preview command:

```text
static-web-server -w deploy/sws.toml
```

Optional later improvement:

- a watch mode that rebuilds on file changes
- a dev script that starts both the builder and SWS preview

---

## 11. Important edge cases to handle early

- request URIs and filesystem paths are not the same thing
- explicit `index.html` URLs should redirect to the canonical directory URL
- all asset URLs referenced from HTML must already be fingerprinted
- alias redirects must not conflict with real pages
- hidden files are ignored by SWS by default, so do not rely on dot-directories like `.well-known/` unless you intentionally change that behavior
- do not use SWS fallback-page mode for this blog because real missing pages should stay 404

---

## 12. Definition of done for v1

The system is done when:

- a Markdown post with code, Mermaid, and math builds successfully
- the generated site is fully navigable with no rewrites for normal page routing
- SWS serves the site using only static files and config
- cache headers are correct for HTML vs fingerprinted assets
- 404 and 50x pages are wired
- feed and sitemap are generated
- old URLs can be preserved via generated redirects

