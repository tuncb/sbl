# Exact `public/` Output Structure

This is the target on-disk structure for the generated site.

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
    first-post/
      index.html
    go-http-routing/
      index.html
    writing-math-in-posts/
      index.html

  pages/
    about/
      index.html

  assets/
    site.3f2c9b61.css

    posts/
      go-http-routing/
        network-stack.91d2f304.webp
        example-snippet.7d20b9ab.png
        diagram-1.55cb0d6a.svg
      writing-math-in-posts/
        diagram-1.a4200af1.svg

    vendor/
      katex-<version>/
        katex.min.css
        fonts/
          KaTeX_Main-Regular.woff2
          KaTeX_Math-Italic.woff2
          ...

    icons/
      apple-touch-icon.2f8c4d10.png
      og-default.4d53d2f0.png
```

## URL mapping rules

- `/` -> `public/index.html`
- `/archive/` -> `public/archive/index.html`
- `/posts/<slug>/` -> `public/posts/<slug>/index.html`
- `/pages/<slug>/` -> `public/pages/<slug>/index.html`
- `/assets/...` -> same path under `public/assets/...`

## Why it is shaped this way

- every HTML route has a real directory on disk
- trailing-slash URLs match the directory layout naturally
- fingerprinted assets can be cached aggressively
- top-level machine-readable files stay where tools expect them
- SWS does not need rewrites for normal page routing

## Do not generate these in v1

Avoid these unless you intentionally change the design:

- `public/posts/<slug>.html`
- `public/archive.html`
- `public/post/index.html` plus a rewrite from some unrelated route
- unhashed long-lived assets under `/assets/`

