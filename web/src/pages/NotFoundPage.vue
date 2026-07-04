<template>
  <main class="page">
    <section class="empty-state">
      <span class="status rejected">404</span>
      <h1>页面不存在</h1>
      <p>这个地址没有对应的内容，可能是文章已迁移、链接拼写错误，或页面尚未发布。</p>
      <div class="header-actions">
        <RouterLink class="button" to="/">返回首页</RouterLink>
        <RouterLink class="button-secondary" to="/archive">查看归档</RouterLink>
      </div>
    </section>
  </main>
</template>
