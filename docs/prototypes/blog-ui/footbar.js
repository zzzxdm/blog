(() => {
  const SITE_START_DATE = new Date(2020, 7, 21, 0, 0, 0);

  function getRuntimeText(now = new Date()) {
    let years = now.getFullYear() - SITE_START_DATE.getFullYear();
    let anniversary = new Date(SITE_START_DATE);
    anniversary.setFullYear(SITE_START_DATE.getFullYear() + years);

    if (anniversary > now) {
      years -= 1;
      anniversary = new Date(SITE_START_DATE);
      anniversary.setFullYear(SITE_START_DATE.getFullYear() + years);
    }

    const totalSeconds = Math.max(0, Math.floor((now - anniversary) / 1000));
    const days = Math.floor(totalSeconds / 86400);
    const hours = Math.floor((totalSeconds % 86400) / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    return `本站已运行 ${years} 年 ${days} 天 ${hours} 小时 ${minutes} 分钟 ${seconds} 秒`;
  }

  class SiteFootbar extends HTMLElement {
    connectedCallback() {
      this.render();
      this.updateRuntime();
      this.timer = window.setInterval(() => this.updateRuntime(), 1000);
    }

    disconnectedCallback() {
      if (this.timer) {
        window.clearInterval(this.timer);
      }
    }

    render() {
      const year = new Date().getFullYear();

      this.innerHTML = `
        <footer class="footbar">
          <div class="footbar-inner">
            <div class="footbar-copy">
              <div>Copyright © ${year} <strong>云间笔记</strong>.</div>
              <div data-runtime></div>
              <div>总访问量: 47087 人次 <span>|</span> 访客人数: 28294 人 <span>|</span> 字数统计: 46.8k 字</div>
            </div>
            <div class="footbar-tools">
              <nav class="social-links" aria-label="站点链接">
                <a class="social-link" href="https://github.com/" aria-label="GitHub">GH</a>
                <a class="social-link" href="mailto:hello@example.com" aria-label="邮箱">✉</a>
              </nav>
            </div>
          </div>
        </footer>
      `;
    }

    updateRuntime() {
      const runtime = this.querySelector("[data-runtime]");
      if (runtime) {
        runtime.textContent = getRuntimeText();
      }
    }
  }

  if (!customElements.get("site-footbar")) {
    customElements.define("site-footbar", SiteFootbar);
  }
})();
