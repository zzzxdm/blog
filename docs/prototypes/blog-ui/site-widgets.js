(() => {
  const searchItems = [
    {
      title: "如何设计一个内容长期增长的博客系统",
      url: "./article.html",
      category: "工程实践",
      summary: "内容模型、发布流程、SEO、缓存、搜索和数据治理。"
    },
    {
      title: "Vue3 内容站的缓存与 SEO 边界",
      url: "./article.html",
      category: "Vue3",
      summary: "接口缓存、服务端 meta、页面更新和性能优化。"
    },
    {
      title: "用户评论系统应该怎么设计",
      url: "./submit.html",
      category: "用户系统",
      summary: "登录用户评论、审核、举报、站内信和禁言机制。"
    },
    {
      title: "为什么博客后台需要文章版本历史",
      url: "./article.html",
      category: "内容治理",
      summary: "版本记录、回滚、审核协作和内容资产保护。"
    },
    {
      title: "让旧文章继续被搜索引擎找到",
      url: "./archive.html",
      category: "SEO",
      summary: "重定向、canonical、Sitemap 和结构化数据。"
    }
  ];

  class SiteSearch extends HTMLElement {
    connectedCallback() {
      this.render();
      this.input = this.querySelector("[data-search-input]");
      this.results = this.querySelector("[data-search-results]");
      this.closeButton = this.querySelector("[data-search-close]");

      this.renderResults(searchItems);
      this.input.addEventListener("input", () => this.filter());
      this.closeButton.addEventListener("click", () => this.close());
      this.addEventListener("click", (event) => {
        if (event.target === this) {
          this.close();
        }
      });

      document.addEventListener("click", (event) => {
        const opener = event.target.closest("[data-search-open]");
        if (opener) {
          event.preventDefault();
          this.open();
        }
      });

      document.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
          this.close();
        }
      });
    }

    render() {
      this.className = "search-overlay";
      this.innerHTML = `
        <section class="search-dialog" role="dialog" aria-modal="true" aria-label="站内搜索">
          <div class="search-dialog-header">
            <input class="input" data-search-input type="search" placeholder="搜索文章、专题或标签" aria-label="搜索关键词">
            <button class="search-close" data-search-close type="button" aria-label="关闭搜索">×</button>
          </div>
          <div class="search-results" data-search-results></div>
        </section>
      `;
    }

    open() {
      this.classList.add("open");
      this.input.focus();
    }

    close() {
      this.classList.remove("open");
    }

    filter() {
      const keyword = this.input.value.trim().toLowerCase();
      const filtered = searchItems.filter((item) => {
        const text = `${item.title} ${item.category} ${item.summary}`.toLowerCase();
        return text.includes(keyword);
      });
      this.renderResults(filtered);
    }

    renderResults(items) {
      if (!items.length) {
        this.results.innerHTML = '<div class="search-result"><strong>没有找到结果</strong><p>换个关键词试试。</p></div>';
        return;
      }

      this.results.innerHTML = items.map((item) => `
        <a class="search-result" href="${item.url}">
          <div class="meta-row"><span class="tag">${item.category}</span></div>
          <strong>${item.title}</strong>
          <p>${item.summary}</p>
        </a>
      `).join("");
    }
  }

  class SiteBacktop extends HTMLElement {
    connectedCallback() {
      this.innerHTML = '<button class="backtop" type="button" aria-label="回到顶部">↑</button>';
      this.button = this.querySelector(".backtop");
      this.button.addEventListener("click", () => window.scrollTo({ top: 0, behavior: "smooth" }));
      window.addEventListener("scroll", () => this.update(), { passive: true });
      this.update();
    }

    update() {
      this.button.classList.toggle("visible", window.scrollY > 360);
    }
  }

  document.addEventListener("click", (event) => {
    const backButton = event.target.closest("[data-back]");
    if (!backButton) {
      return;
    }

    event.preventDefault();
    const fallback = backButton.getAttribute("data-fallback") || "./archive.html";

    if (window.history.length > 1) {
      window.history.back();
    } else {
      window.location.href = fallback;
    }
  });

  function setArchiveView(view, root = document) {
    root.querySelectorAll("[data-archive-view-button]").forEach((button) => {
      button.classList.toggle("active", button.getAttribute("data-archive-view-button") === view);
    });

    root.querySelectorAll("[data-archive-view]").forEach((section) => {
      section.hidden = section.getAttribute("data-archive-view") !== view;
    });

    root.querySelectorAll("[data-page-link]").forEach((link) => {
      const url = new URL(link.getAttribute("href"), window.location.href);
      url.searchParams.set("view", view);
      link.setAttribute("href", `${url.pathname.split("/").pop()}${url.search}`);
    });
  }

  function getArchivePage() {
    const params = new URLSearchParams(window.location.search);
    const page = Number.parseInt(params.get("page") || "1", 10);
    return Number.isFinite(page) && page > 0 ? page : 1;
  }

  function setArchivePage(page, root = document, shouldPushState = false, viewOverride = null) {
    const currentPage = Math.max(1, Math.min(page, 11));
    const view = viewOverride || getInitialArchiveView();
    const indicator = root.querySelector("[data-page-indicator]");

    if (indicator) {
      indicator.textContent = `第 ${currentPage} 页`;
    }

    root.querySelectorAll("[data-page-number]").forEach((link) => {
      const linkPage = Number.parseInt(link.getAttribute("data-page-number"), 10);
      link.classList.toggle("current", linkPage === currentPage);
    });

    root.querySelectorAll("[data-page-prev]").forEach((link) => {
      const prevPage = Math.max(currentPage - 1, 1);
      link.setAttribute("href", `archive.html?page=${prevPage}&view=${view}`);
      link.classList.toggle("disabled", currentPage <= 1);
    });

    root.querySelectorAll("[data-page-next]").forEach((link) => {
      const nextPage = Math.min(currentPage + 1, 11);
      link.setAttribute("href", `archive.html?page=${nextPage}&view=${view}`);
      link.classList.toggle("disabled", currentPage >= 11);
    });

    if (shouldPushState) {
      const url = new URL(window.location.href);
      url.searchParams.set("page", String(currentPage));
      url.searchParams.set("view", view);
      window.history.pushState({}, "", url);
    }
  }

  function getInitialArchiveView() {
    const params = new URLSearchParams(window.location.search);
    const view = params.get("view") || window.localStorage.getItem("archive:view");
    return view === "list" ? "list" : "grid";
  }

  document.addEventListener("DOMContentLoaded", () => {
    if (document.querySelector("[data-archive-view]")) {
      const initialView = getInitialArchiveView();
      window.localStorage.setItem("archive:view", initialView);
      setArchiveView(initialView);
      setArchivePage(getArchivePage(), document, false, initialView);
    }
  });

  document.addEventListener("click", (event) => {
    const viewButton = event.target.closest("[data-archive-view-button]");
    if (!viewButton) {
      return;
    }

    const view = viewButton.getAttribute("data-archive-view-button");
    const root = viewButton.closest("main") || document;
    window.localStorage.setItem("archive:view", view);
    setArchiveView(view, root);
    setArchivePage(getArchivePage(), root, true, view);
  });

  document.addEventListener("click", (event) => {
    const pageLink = event.target.closest("[data-page-link]");
    if (!pageLink) {
      return;
    }

    event.preventDefault();

    if (pageLink.classList.contains("disabled")) {
      return;
    }

    const url = new URL(pageLink.getAttribute("href"), window.location.href);
    const page = Number.parseInt(url.searchParams.get("page") || pageLink.getAttribute("data-page-number") || "1", 10);
    const view = url.searchParams.get("view") || getInitialArchiveView();
    const root = pageLink.closest("main") || document;

    setArchivePage(page, root, true, view);
    root.querySelector(".archive-viewbar")?.scrollIntoView({ behavior: "smooth", block: "start" });
  });

  if (!customElements.get("site-search")) {
    customElements.define("site-search", SiteSearch);
  }

  if (!customElements.get("site-backtop")) {
    customElements.define("site-backtop", SiteBacktop);
  }
})();
