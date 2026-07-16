export interface MarkdownHeading {
  id: string;
  level: number;
  text: string;
}

export interface RenderMarkdownOptions {
  allowImages?: boolean;
}

export function renderMarkdown(markdown: string, options: RenderMarkdownOptions = {}): string {
  const allowImages = options.allowImages !== false;
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

    html.push(`<p>${renderInline(paragraph.join(" "), allowImages)}</p>`);
    paragraph = [];
  }

  function flushList() {
    if (!list.length) {
      return;
    }

    html.push(`<ul>${list.map((item) => `<li>${renderInline(item, allowImages)}</li>`).join("")}</ul>`);
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
      html.push(`<h${level} id="${escapeAttribute(slugify(text))}">${renderInline(text, allowImages)}</h${level}>`);
      return;
    }

    if (trimmed.startsWith(">")) {
      flushParagraph();
      flushList();
      html.push(`<blockquote>${renderInline(trimmed.replace(/^>\s?/, ""), allowImages)}</blockquote>`);
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

  return sanitizeRenderedHtml(html.join("\n"));
}

/** Safer markdown rendering for user comments (no images + final HTML sanitize). */
export function renderCommentMarkdown(markdown: string): string {
  return renderMarkdown(stripRawHtmlTags(markdown), { allowImages: false });
}

export function extractMarkdownHeadings(markdown: string): MarkdownHeading[] {
  const headings: MarkdownHeading[] = [];
  let inCode = false;

  markdown
    .replace(/\r\n/g, "\n")
    .split("\n")
    .forEach((line) => {
      const trimmed = line.trim();
      if (trimmed.startsWith("```")) {
        inCode = !inCode;
        return;
      }
      if (inCode) {
        return;
      }

      const heading = /^(#{1,3})\s+(.+)$/.exec(trimmed);
      if (!heading) {
        return;
      }

      const text = heading[2].trim();
      headings.push({
        id: slugify(text),
        level: heading[1].length + 1,
        text
      });
    });

  return headings;
}

function renderInline(value: string, allowImages: boolean): string {
  const codeSpans: string[] = [];
  let rendered = escapeHtml(value).replace(/`([^`]+)`/g, (_match, code: string) => {
    codeSpans.push(`<code>${code}</code>`);
    return `\u0000${codeSpans.length - 1}\u0000`;
  });

  if (allowImages) {
    rendered = rendered.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+(['"])(.*?)\3)?\)/g, (_match, alt: string, src: string, _quote: string, title: string) => {
      const titleAttribute = title ? ` title="${escapeAttribute(title)}"` : "";
      return `<img src="${escapeAttribute(safeHref(src))}" alt="${escapeAttribute(alt)}"${titleAttribute}>`;
    });
  } else {
    rendered = rendered.replace(/!\[([^\]]*)\]\(([^)\s]+)(?:\s+(['"]).*?\3)?\)/g, (_match, alt: string) => escapeHtml(alt || ""));
  }

  rendered = rendered.replace(/\[([^\]]+)\]\(([^)\s]+)(?:\s+(['"]).*?\3)?\)/g, (_match, label: string, href: string) => {
    return `<a href="${escapeAttribute(safeHref(href))}" rel="nofollow noopener noreferrer" target="_blank">${label}</a>`;
  });
  rendered = rendered.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  rendered = rendered.replace(/\*([^*]+)\*/g, "<em>$1</em>");
  rendered = rendered.replace(/\u0000(\d+)\u0000/g, (_match, index: string) => codeSpans[Number(index)] ?? "");

  return rendered;
}

function safeHref(value: string): string {
  const href = value.trim();
  const lower = href.toLowerCase();

  if (
    href.startsWith("/") ||
    href.startsWith("#") ||
    lower.startsWith("http://") ||
    lower.startsWith("https://") ||
    lower.startsWith("mailto:")
  ) {
    // Block sneaky schemes after whitespace / control chars, e.g. "java\tscript:".
    if (/^[\w.+-]+(?:\s|%|\\)/i.test(href) && !/^(https?|mailto):/i.test(href) && !href.startsWith("/") && !href.startsWith("#")) {
      return "#";
    }
    return href;
  }

  return "#";
}

/** Final HTML pass for defense-in-depth against renderer bugs. */
export function sanitizeRenderedHtml(html: string): string {
  return html
    .replace(/<\/?(script|style|iframe|object|embed|link|meta|form|input|button|textarea|select)\b[^>]*>/gi, "")
    .replace(/\son[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    .replace(/\s(href|src)\s*=\s*("|')\s*javascript:[^"']*\2/gi, ' $1="#"')
    .replace(/\s(href|src)\s*=\s*javascript:[^\s>]+/gi, ' $1="#"')
    .replace(/\s(href|src)\s*=\s*("|')\s*data:[^"']*\2/gi, ' $1="#"')
    .replace(/\s(href|src)\s*=\s*data:[^\s>]+/gi, ' $1="#"');
}

function stripRawHtmlTags(value: string): string {
  return value.replace(/<\/?[a-zA-Z][^>]*>/g, "");
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
