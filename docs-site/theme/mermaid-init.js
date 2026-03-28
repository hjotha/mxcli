// Render mermaid diagrams in mdbook.
// mdbook renders ```mermaid blocks as <code class="language-mermaid"> inside <pre>.
// This script loads mermaid from CDN and converts them to rendered SVG diagrams.
(function () {
  var blocks = document.querySelectorAll('code.language-mermaid');
  if (blocks.length === 0) return;

  var script = document.createElement('script');
  script.src = 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js';
  script.onload = function () {
    mermaid.initialize({ startOnLoad: false, theme: 'default' });

    blocks.forEach(function (code, i) {
      var pre = code.parentElement;
      var container = document.createElement('div');
      container.className = 'mermaid';
      container.textContent = code.textContent;
      pre.parentElement.replaceChild(container, pre);
    });

    mermaid.run();
  };
  document.head.appendChild(script);
})();
