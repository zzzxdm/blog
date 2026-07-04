<script setup lang="ts">
import AdminLayout from "../../components/AdminLayout.vue";
</script>

<template>
  <AdminLayout title="用户管理" description="管理注册用户、作者账号、禁言状态和评论行为。" mobile-title="用户管理" primary-action="邀请作者">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">导出用户</button>
        <button class="button" type="button">邀请作者</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="用户统计">
      <div class="stat-card"><span>注册用户</span><strong>1,248</strong></div>
      <div class="stat-card"><span>已验证邮箱</span><strong>932</strong></div>
      <div class="stat-card"><span>作者</span><strong>8</strong></div>
      <div class="stat-card"><span>禁言中</span><strong>12</strong></div>
    </section>

    <section class="table-panel" aria-label="用户列表">
      <form class="table-toolbar">
        <input class="input" type="search" placeholder="搜索用户名、邮箱、角色" aria-label="搜索用户">
        <select class="input" aria-label="用户状态">
          <option>全部状态</option>
          <option>正常</option>
          <option>待验证</option>
          <option>禁言</option>
          <option>封禁</option>
        </select>
        <select class="input" aria-label="角色">
          <option>全部角色</option>
          <option>注册用户</option>
          <option>作者</option>
          <option>编辑</option>
          <option>管理员</option>
        </select>
      </form>

      <table>
        <thead>
          <tr>
            <th>用户</th>
            <th>角色</th>
            <th>状态</th>
            <th>评论</th>
            <th>收藏</th>
            <th>最近登录</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><strong>林一</strong><div class="meta-row"><span>linyi@example.com</span><span>已验证邮箱</span></div></td>
            <td>注册用户</td>
            <td><span class="status published">正常</span></td>
            <td>42</td>
            <td>18</td>
            <td>今天 15:12</td>
            <td><div class="header-actions"><button class="button-secondary" type="button">查看</button><button class="button-secondary" type="button">禁言</button></div></td>
          </tr>
          <tr>
            <td><strong>管理员</strong><div class="meta-row"><span>author@example.com</span><span>双因素已开启</span></div></td>
            <td>管理员</td>
            <td><span class="status published">正常</span></td>
            <td>128</td>
            <td>6</td>
            <td>刚刚</td>
            <td><button class="button-secondary" type="button">查看</button></td>
          </tr>
          <tr>
            <td><strong>market_user</strong><div class="meta-row"><span>market@example.com</span><span>推广链接举报 3 次</span></div></td>
            <td>注册用户</td>
            <td><span class="status muted">禁言</span></td>
            <td>9</td>
            <td>0</td>
            <td>12 分钟前</td>
            <td><div class="header-actions"><button class="button-secondary" type="button">解除</button><button class="button-secondary" type="button">封禁</button></div></td>
          </tr>
          <tr>
            <td><strong>noise_2048</strong><div class="meta-row"><span>noise@example.com</span><span>未验证邮箱</span></div></td>
            <td>注册用户</td>
            <td><span class="status banned">封禁</span></td>
            <td>21</td>
            <td>0</td>
            <td>今天 09:18</td>
            <td><button class="button-secondary" type="button">查看</button></td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
