export function renderMarkdown(markdown: string): string {
  const lines = markdown.replace(/\r\n/g, "\n").split("\n");
  const html: string[] = [];
  let paragraph: string[] = [];
  let list: string[] = [];
  let code: string[] = [];
  let inCode = false;

  function flushParagraph() {
    if (!paragraph.length) {
      return;
    }

    html.push(`<p>${renderInline(paragraph.join(" "))}</p>`);
    paragraph = [];
  }

  function flushList() {
    if (!list.length) {
      return;
    }

    html.push(`<ul>${list.map((item) => `<li>${renderInline(item)}</li>`).join("")}</ul>`);
    list = [];
  }

  lines.forEach((line) => {
    const trimmed = line.trim();

    if (trimmed.startsWith("```")) {
      if (inCode) {
        html.push(`<div class="code-block"><pre><code>${escapeHtml(code.join("\n"))}</code></pre></div>`);
        code = [];
        inCode = false;
      } else {
        flushParagraph();
        flushList();
        inCode = true;
      }
      return;
    }

    if (inCode) {
      code.push(line);
      return;
    }

    if (!trimmed) {
      flushParagraph();
      flushList();
      return;
    }

    const heading = /^(#{1,3})\s+(.+)$/.exec(trimmed);
    if (heading) {
      flushParagraph();
      flushList();
      const level = heading[1].length + 1;
      const text = heading[2].trim();
      html.push(`<h${level} id="${slugify(text)}">${renderInline(text)}</h${level}>`);
      return;
    }

    if (trimmed.startsWith(">")) {
      flushParagraph();
      flushList();
      html.push(`<blockquote>${renderInline(trimmed.replace(/^>\s?/, ""))}</blockquote>`);
      return;
    }

    const unordered = /^[-*]\s+(.+)$/.exec(trimmed);
    if (unordered) {
      flushParagraph();
      list.push(unordered[1]);
      return;
    }

    paragraph.push(trimmed);
  });

  if (inCode) {
    html.push(`<div class="code-block"><pre><code>${escapeHtml(code.join("\n"))}</code></pre></div>`);
  }

  flushParagraph();
  flushList();

  return html.join("\n");
}

function renderInline(value: string): string {
  const codeSpans: string[] = [];
  let rendered = escapeHtml(value).replace(/`([^`]+)`/g, (_match, code: string) => {
    codeSpans.push(`<code>${code}</code>`);
    return `\u0000${codeSpans.length - 1}\u0000`;
  });

  rendered = rendered.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_match, label: string, href: string) => {
    return `<a href="${escapeAttribute(safeHref(href))}" rel="nofollow noopener">${label}</a>`;
  });
  rendered = rendered.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  rendered = rendered.replace(/\*([^*]+)\*/g, "<em>$1</em>");
  rendered = rendered.replace(/\u0000(\d+)\u0000/g, (_match, index: string) => codeSpans[Number(index)] ?? "");

  return rendered;
}

function safeHref(value: string): string {
  const href = value.trim();
  if (
    href.startsWith("/") ||
    href.startsWith("#") ||
    href.startsWith("http://") ||
    href.startsWith("https://") ||
    href.startsWith("mailto:")
  ) {
    return href;
  }

  return "#";
}

function slugify(value: string): string {
  const slug = value
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "");

  return slug || "section";
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function escapeAttribute(value: string): string {
  return escapeHtml(value).replace(/`/g, "&#96;");
}
