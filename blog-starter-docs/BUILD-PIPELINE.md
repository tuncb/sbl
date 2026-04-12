# Build Pipeline Outline

## Suggested flow

1. Discover posts under `content/posts/*/index.md`.
2. Parse YAML front matter and validate required fields.
3. Read the Markdown body.
4. Extract Mermaid fenced blocks and replace them with placeholders.
5. Render the remaining Markdown to an HTML fragment with Goldmark.
6. Syntax-highlight fenced code blocks.
7. Render each Mermaid placeholder to SVG and inject the result back into the HTML fragment.
8. Render math expressions to KaTeX HTML.
9. Wrap the article fragment with Go templates.
10. Write the page to `public/posts/<slug>/index.html`.
11. Copy static files and post-local assets.
12. Build the index page, archive page, feed, sitemap, and 404 page.

## Important ordering

- Mermaid must be handled before normal code highlighting, or at least excluded from the highlighting path.
- Math must run after code blocks are already HTML so code samples are not interpreted as equations.
- Raw HTML should stay disallowed in post source; only the build step should inject trusted HTML such as KaTeX output or inline SVG.

## Recommended outputs

```text
public/
  index.html
  archive/index.html
  posts/building-a-small-blog/index.html
  assets/
    site.css
    katex/
      katex.min.css
      fonts/
    diagrams/
      <hash>.svg
```

## Good hard-fail checks

- duplicate slug
- invalid date
- missing summary
- missing asset file
- Mermaid render failure
- KaTeX parse failure
- broken internal link
