<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>投稿</h1>
        <p>登录用户可以提交文章草稿，审核通过后会发布到站点。</p>
      </div>
      <div class="meta-row">
        <span>自动保存于 15:42</span>
        <span class="status draft">草稿</span>
      </div>
    </section>

    <section class="editor-layout">
      <div class="editor-panel">
        <div class="editor-toolbar">
          <div class="tool-group" aria-label="投稿编辑工具栏">
            <button class="tool" type="button" aria-label="加粗">B</button>
            <button class="tool" type="button" aria-label="斜体">I</button>
            <button class="tool" type="button" aria-label="标题">H</button>
            <button class="tool" type="button" aria-label="引用">"</button>
            <button class="tool" type="button" aria-label="代码">{ }</button>
            <button class="tool" type="button" aria-label="链接">↗</button>
          </div>
          <div class="meta-row">
            <span>Markdown</span>
            <span>实时预览</span>
          </div>
        </div>

        <div class="editor-grid">
          <textarea class="markdown-area" aria-label="投稿 Markdown 编辑区"># 用户评论系统应该怎么设计

登录用户评论、审核、举报、通知和禁言机制，是开放内容站点的基础能力。

## 评论不是简单留言

评论需要和用户系统、通知系统、反垃圾策略一起设计。

## 审核状态

- 待审核
- 已通过
- 已拒绝
- 垃圾评论

> 好的评论区应该帮助内容继续生长，而不是成为后台负担。</textarea>

          <article class="preview-area">
            <h1>用户评论系统应该怎么设计</h1>
            <p>登录用户评论、审核、举报、通知和禁言机制，是开放内容站点的基础能力。</p>
            <h2>评论不是简单留言</h2>
            <p>评论需要和用户系统、通知系统、反垃圾策略一起设计。</p>
            <h2>审核状态</h2>
            <ul>
              <li>待审核</li>
              <li>已通过</li>
              <li>已拒绝</li>
              <li>垃圾评论</li>
            </ul>
            <blockquote>好的评论区应该帮助内容继续生长，而不是成为后台负担。</blockquote>
          </article>
        </div>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>投稿流程</h2>
          </div>
          <div class="stepper">
            <div class="step done">
              <span class="step-index">1</span>
              <div>
                <strong>填写内容</strong>
                <div class="meta-row"><span>当前草稿已自动保存</span></div>
              </div>
            </div>
            <div class="step current">
              <span class="step-index">2</span>
              <div>
                <strong>提交审核</strong>
                <div class="meta-row"><span>编辑会检查质量、格式和安全</span></div>
              </div>
            </div>
            <div class="step">
              <span class="step-index">3</span>
              <div>
                <strong>通过后发布</strong>
                <div class="meta-row"><span>发布后进入公开文章列表</span></div>
              </div>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>文章信息</h2>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="title">标题</label>
              <input class="input" id="title" value="用户评论系统应该怎么设计">
            </div>
            <div class="field">
              <label for="summary">摘要</label>
              <textarea class="input" id="summary">从登录用户评论、审核、举报、通知和禁言机制出发，设计一个可维护的评论系统。</textarea>
            </div>
            <div class="field">
              <label for="category">建议分类</label>
              <select class="input" id="category">
                <option>用户系统</option>
                <option>工程实践</option>
                <option>内容治理</option>
              </select>
            </div>
            <div class="field">
              <label for="tags">标签</label>
              <input class="input" id="tags" value="评论, 用户系统, 审核">
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>封面</h2>
          </div>
          <div class="settings-stack">
            <img
              src="https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=700&q=80"
              alt="投稿封面预览"
              style="border-radius: 8px; aspect-ratio: 16 / 9; object-fit: cover;"
            >
            <button class="button-secondary" type="button">上传封面</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>提交</h2>
          </div>
          <div class="settings-stack">
            <div class="review-note">
              <strong>投稿不会直接公开</strong>
              <p>提交后进入待审核状态。编辑可能会通过、退回修改或拒绝投稿。</p>
            </div>
            <button class="button-secondary" type="button">保存草稿</button>
            <button class="button" type="button">提交审核</button>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
