(function () {
  const script =
    document.currentScript ||
    document.querySelector("script[data-sbl-render]");
  if (!script) {
    return;
  }

  const mathNodes = Array.from(
    document.querySelectorAll(".sbl-math-inline, .sbl-math-display")
  );
  const mermaidNodes = Array.from(document.querySelectorAll("pre.sbl-mermaid"));
  const codeNodes = Array.from(
    document.querySelectorAll('code[class*="language-"]')
  );
  if (
    mathNodes.length === 0 &&
    mermaidNodes.length === 0 &&
    codeNodes.length === 0
  ) {
    return;
  }

  const loadedScripts = new Map();
  const loadedStylesheets = new Map();

  function errorMessage(error) {
    if (error instanceof Error && error.message) {
      return error.message;
    }
    return String(error);
  }

  function logRenderFailure(kind, stage, error, source) {
    console.error(`sbl ${kind} ${stage} failure:`, {
      message: errorMessage(error),
      source,
    });
  }

  function setButtonState(button, text, durationMs) {
    const original = button.textContent;
    button.textContent = text;
    button.disabled = true;
    window.setTimeout(() => {
      button.textContent = original;
      button.disabled = false;
    }, durationMs);
  }

  async function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return;
    }

    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "");
    textarea.style.position = "absolute";
    textarea.style.left = "-9999px";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    textarea.remove();
  }

  function createActionButton(label, onClick) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "sbl-render-action";
    button.textContent = label;
    button.addEventListener("click", onClick);
    return button;
  }

  function createCopyButton(source) {
    const button = createActionButton("Copy source", async () => {
      try {
        await copyText(source);
        setButtonState(button, "Copied", 1200);
      } catch (error) {
        console.error("sbl copy failed:", errorMessage(error));
        setButtonState(button, "Copy failed", 1600);
      }
    });
    return button;
  }

  function createFallbackContent(options) {
    const {
      title,
      summary,
      detail,
      source,
      inline,
    } = options;

    if (inline) {
      const container = document.createElement("span");
      container.className = "sbl-render-fallback sbl-render-fallback-inline";
      container.setAttribute("role", "note");
      container.title = detail ? `${summary} ${detail}` : summary;

      const label = document.createElement("span");
      label.className = "sbl-render-fallback-label";
      label.textContent = title;

      const code = document.createElement("code");
      code.className = "sbl-render-fallback-source-inline";
      code.textContent = source;

      container.append(label, code, createCopyButton(source));
      return container;
    }

    const container = document.createElement("div");
    container.className = "sbl-render-fallback sbl-render-fallback-block";
    container.setAttribute("role", "note");

    const header = document.createElement("div");
    header.className = "sbl-render-fallback-header";

    const heading = document.createElement("strong");
    heading.className = "sbl-render-fallback-title";
    heading.textContent = title;

    const summaryNode = document.createElement("p");
    summaryNode.className = "sbl-render-fallback-summary";
    summaryNode.textContent = summary;

    header.append(heading, summaryNode);

    if (detail) {
      const detailNode = document.createElement("p");
      detailNode.className = "sbl-render-fallback-detail";
      detailNode.textContent = detail;
      header.appendChild(detailNode);
    }

    const actions = document.createElement("div");
    actions.className = "sbl-render-fallback-actions";
    actions.appendChild(createCopyButton(source));

    const details = document.createElement("details");
    details.className = "sbl-render-fallback-details";

    const detailsSummary = document.createElement("summary");
    detailsSummary.textContent = "Show source";

    const pre = document.createElement("pre");
    pre.className = "sbl-render-fallback-source-block";

    const code = document.createElement("code");
    code.textContent = source;
    pre.appendChild(code);

    details.append(detailsSummary, pre);
    container.append(header, actions, details);
    return container;
  }

  function replaceWithFallback(node, options) {
    node.replaceWith(createFallbackContent(options));
  }

  function blockFailureOptions(kind, stage, source, error) {
    const detail = errorMessage(error);
    if (kind === "math") {
      if (stage === "load") {
        return {
          title: "Math unavailable",
          summary: "The math renderer could not be loaded.",
          detail,
          source,
          inline: false,
        };
      }
      return {
        title: "Math failed to render",
        summary: "This expression could not be rendered.",
        detail,
        source,
        inline: false,
      };
    }

    if (stage === "load") {
      return {
        title: "Diagram unavailable",
        summary: "The Mermaid renderer could not be loaded.",
        detail,
        source,
        inline: false,
      };
    }
    return {
      title: "Diagram failed to render",
      summary: "This Mermaid diagram could not be rendered.",
      detail,
      source,
      inline: false,
    };
  }

  function inlineFailureOptions(source, error) {
    return {
      title: "Math failed",
      summary: "This inline expression could not be rendered.",
      detail: errorMessage(error),
      source,
      inline: true,
    };
  }

  function loadScript(src) {
    if (!src) {
      return Promise.resolve();
    }
    if (loadedScripts.has(src)) {
      return loadedScripts.get(src);
    }

    const existing = document.querySelector(`script[src="${src}"]`);
    if (existing) {
      const promise = Promise.resolve();
      loadedScripts.set(src, promise);
      return promise;
    }

    const promise = new Promise((resolve, reject) => {
      const tag = document.createElement("script");
      tag.src = src;
      tag.defer = true;
      tag.onload = resolve;
      tag.onerror = () => reject(new Error(`failed to load script: ${src}`));
      document.head.appendChild(tag);
    });
    loadedScripts.set(src, promise);
    return promise;
  }

  function loadStylesheet(href) {
    if (!href) {
      return Promise.resolve();
    }
    if (loadedStylesheets.has(href)) {
      return loadedStylesheets.get(href);
    }

    const existing = document.querySelector(`link[href="${href}"]`);
    if (existing) {
      const promise = Promise.resolve();
      loadedStylesheets.set(href, promise);
      return promise;
    }

    const promise = new Promise((resolve, reject) => {
      const tag = document.createElement("link");
      tag.rel = "stylesheet";
      tag.href = href;
      tag.onload = resolve;
      tag.onerror = () =>
        reject(new Error(`failed to load stylesheet: ${href}`));
      document.head.appendChild(tag);
    });
    loadedStylesheets.set(href, promise);
    return promise;
  }

  async function renderMath() {
    if (mathNodes.length === 0) {
      return;
    }

    try {
      await Promise.all([
        loadStylesheet(script.dataset.katexCssUrl),
        loadScript(script.dataset.katexJsUrl),
      ]);
    } catch (error) {
      for (const node of mathNodes) {
        const source = node.textContent || "";
        const isInline = node.classList.contains("sbl-math-inline");
        logRenderFailure("math", "load", error, source);
        replaceWithFallback(
          node,
          isInline
            ? inlineFailureOptions(source, error)
            : blockFailureOptions("math", "load", source, error)
        );
      }
      return;
    }

    const katex = window.katex;
    if (!katex || typeof katex.render !== "function") {
      const error = new Error("KaTeX did not expose a browser render API");
      for (const node of mathNodes) {
        const source = node.textContent || "";
        const isInline = node.classList.contains("sbl-math-inline");
        logRenderFailure("math", "load", error, source);
        replaceWithFallback(
          node,
          isInline
            ? inlineFailureOptions(source, error)
            : blockFailureOptions("math", "load", source, error)
        );
      }
      return;
    }

    for (const node of mathNodes) {
      const source = node.textContent || "";
      const isInline = node.classList.contains("sbl-math-inline");
      try {
        katex.render(source, node, {
          displayMode: node.classList.contains("sbl-math-display"),
          output: "htmlAndMathML",
          throwOnError: true,
          strict: "error",
        });
        node.removeAttribute("title");
      } catch (error) {
        logRenderFailure("math", "render", error, source);
        replaceWithFallback(
          node,
          isInline
            ? inlineFailureOptions(source, error)
            : blockFailureOptions("math", "render", source, error)
        );
      }
    }
  }

  async function renderMermaid() {
    if (mermaidNodes.length === 0) {
      return;
    }

    try {
      await loadScript(script.dataset.mermaidJsUrl);
    } catch (error) {
      for (const node of mermaidNodes) {
        const source = node.textContent || "";
        logRenderFailure("mermaid", "load", error, source);
        replaceWithFallback(
          node,
          blockFailureOptions("mermaid", "load", source, error)
        );
      }
      return;
    }

    const mermaid = window.mermaid;
    if (
      !mermaid ||
      typeof mermaid.initialize !== "function" ||
      typeof mermaid.render !== "function"
    ) {
      const error = new Error("Mermaid did not expose a browser render API");
      for (const node of mermaidNodes) {
        const source = node.textContent || "";
        logRenderFailure("mermaid", "load", error, source);
        replaceWithFallback(
          node,
          blockFailureOptions("mermaid", "load", source, error)
        );
      }
      return;
    }

    mermaid.initialize({
      startOnLoad: false,
      securityLevel: "strict",
    });

    for (let index = 0; index < mermaidNodes.length; index += 1) {
      const node = mermaidNodes[index];
      const source = node.textContent || "";
      try {
        const result = await mermaid.render(`sbl-mermaid-${index + 1}`, source);
        const rendered = document.createElement("div");
        rendered.className = "sbl-mermaid-rendered";
        if (typeof result === "string") {
          rendered.innerHTML = result;
        } else {
          rendered.innerHTML = result.svg;
          if (typeof result.bindFunctions === "function") {
            result.bindFunctions(rendered);
          }
        }
        node.replaceWith(rendered);
      } catch (error) {
        logRenderFailure("mermaid", "render", error, source);
        replaceWithFallback(
          node,
          blockFailureOptions("mermaid", "render", source, error)
        );
      }
    }
  }

  async function renderCode() {
    if (codeNodes.length === 0) {
      return;
    }

    window.Prism = window.Prism || {};
    window.Prism.manual = true;

    try {
      await loadScript(script.dataset.prismCoreJsUrl);
      await loadScript(script.dataset.prismAutoloaderJsUrl);
    } catch (error) {
      console.error("sbl code load failure:", errorMessage(error));
      return;
    }

    const prism = window.Prism;
    if (!prism || typeof prism.highlightAllUnder !== "function") {
      console.error(
        "sbl code load failure:",
        "Prism did not expose a browser highlight API"
      );
      return;
    }

    if (
      prism.plugins &&
      prism.plugins.autoloader &&
      typeof script.dataset.prismLanguagesPath === "string"
    ) {
      prism.plugins.autoloader.languages_path =
        script.dataset.prismLanguagesPath;
    }

    prism.highlightAllUnder(document);
  }

  Promise.all([renderMath(), renderMermaid(), renderCode()]).catch((error) => {
    console.error("sbl client rendering failed:", errorMessage(error));
  });
})();
