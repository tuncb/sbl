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
  if (mathNodes.length === 0 && mermaidNodes.length === 0) {
    return;
  }

  const loadedScripts = new Map();
  const loadedStylesheets = new Map();

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
      tag.onerror = () => reject(new Error(`failed to load stylesheet: ${href}`));
      document.head.appendChild(tag);
    });
    loadedStylesheets.set(href, promise);
    return promise;
  }

  function setRenderError(node, error) {
    node.classList.add("sbl-render-error");
    node.setAttribute("title", error instanceof Error ? error.message : String(error));
  }

  async function renderMath() {
    if (mathNodes.length === 0) {
      return;
    }

    await Promise.all([
      loadStylesheet(script.dataset.katexCssUrl),
      loadScript(script.dataset.katexJsUrl),
    ]);

    const katex = window.katex;
    if (!katex || typeof katex.render !== "function") {
      throw new Error("KaTeX did not expose a browser render API");
    }

    for (const node of mathNodes) {
      const source = node.textContent || "";
      try {
        katex.render(source, node, {
          displayMode: node.classList.contains("sbl-math-display"),
          output: "htmlAndMathML",
          throwOnError: false,
          strict: "error",
        });
        node.removeAttribute("title");
      } catch (error) {
        node.textContent = source;
        setRenderError(node, error);
      }
    }
  }

  async function renderMermaid() {
    if (mermaidNodes.length === 0) {
      return;
    }

    await loadScript(script.dataset.mermaidJsUrl);

    const mermaid = window.mermaid;
    if (!mermaid || typeof mermaid.initialize !== "function" || typeof mermaid.render !== "function") {
      throw new Error("Mermaid did not expose a browser render API");
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
        setRenderError(node, error);
      }
    }
  }

  Promise.all([renderMath(), renderMermaid()]).catch((error) => {
    const message = error instanceof Error ? error.message : String(error);
    console.error("sbl client rendering failed:", message);
  });
})();
