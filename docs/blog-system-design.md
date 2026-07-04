# 现代化博客系统设计方案

## 1. 项目概述

本方案用于设计并实现一个现代化博客系统，适用于个人博客、团队博客、技术内容站点或轻量内容社区。

配套开发计划见：[博客系统开发计划](./blog-development-plan.md)。

系统目标包括：

- 提供良好的写作、发布和内容管理体验
- 支持高性能访问和良好的 SEO 表现
- 支持分类、标签、搜索、评论、订阅等博客常用功能
- 具备清晰的后台管理能力
- 具备后续扩展为多作者平台、知识库或内容社区的基础

### 1.1 成熟系统参考

本方案参考以下成熟博客和内容系统的设计经验：

| 系统 | 可借鉴能力 |
| --- | --- |
| WordPress | 主题体系、插件生态、媒体库、页面管理、菜单管理、评论审核、导入导出 |
| Ghost | 简洁写作体验、会员订阅、Newsletter、内容预览、SEO 友好、出版工作流 |
| Medium / Hashnode | 阅读体验、作者主页、关注订阅、推荐分发、社交分享、互动数据 |
| Hugo / Jekyll | 静态化部署、高性能访问、Markdown 内容组织、版本化内容管理 |
| Strapi / Sanity | Headless CMS、结构化内容模型、API 优先、多端内容分发 |

因此，系统不只需要完成文章增删改查，还需要补充内容生产、运营增长、数据治理、主题扩展、迁移备份和稳定性保障等能力。

## 2. 用户角色

| 角色 | 说明 | 主要权限 |
| --- | --- | --- |
| 访客 | 未登录用户 | 浏览文章、搜索文章、查看分类和标签，可按配置提交访客评论 |
| 注册用户 | 已登录普通用户 | 评论、回复、点赞、点踩、收藏、订阅、投稿、查看站内信、管理自己的评论和投稿 |
| 作者 | 内容创作者 | 创建和管理自己的文章 |
| 编辑 | 内容运营人员 | 管理文章、分类、标签、评论 |
| 管理员 | 系统管理者 | 管理所有内容、用户、权限和系统配置 |

## 3. 功能设计

### 3.1 前台功能

#### 首页

- 最新文章列表
- 推荐文章
- 热门文章
- 分类导航
- 标签入口
- 作者信息展示
- 友情链接或站点信息

#### 文章列表页

- 按发布时间倒序展示文章
- 支持分页或无限滚动
- 支持按分类筛选
- 支持按标签筛选
- 支持按热度、时间排序

#### 文章详情页

- 文章标题、摘要、正文
- 作者、发布时间、更新时间
- 分类、标签
- 阅读量、点赞数、点踩数、评论数
- 点赞 / 点踩反馈
- 文章目录导航
- 代码高亮
- 图片预览
- 上一篇 / 下一篇
- 相关文章推荐
- 分享链接

#### 分类与标签

- 分类列表
- 分类详情页
- 标签云
- 标签详情页
- 分类和标签的文章聚合

#### 搜索

- 文章标题搜索
- 文章正文搜索
- 标签搜索
- 搜索结果高亮
- 热门搜索词统计

#### 评论

- 文章评论
- 评论回复
- 评论点赞
- 评论举报
- 评论通知
- 登录用户评论
- 可配置访客评论
- 评论审核
- 敏感词过滤
- IP 限流
- 可选第三方登录评论

#### 用户账号

- 注册
- 登录
- 退出登录
- 邮箱验证
- 找回密码
- 第三方登录，可选
- 用户资料页
- 我的评论
- 我的收藏
- 我的投稿
- 站内信与消息通知
- 未读消息提醒
- 账号注销

#### 用户投稿

- 登录用户创建投稿
- 投稿草稿自动保存
- 投稿封面上传
- 投稿分类和标签建议
- 投稿预览
- 提交审核
- 查看审核状态
- 查看退回原因
- 修改后重新提交
- 审核通过后展示为正式文章

#### 订阅

- RSS 订阅
- 邮件订阅，后续可选
- 新文章邮件推送，后续可选
- 专栏订阅，后续可选

#### SEO

- 自定义页面标题
- 自定义页面描述
- Open Graph 分享信息
- Sitemap
- Robots.txt
- 结构化数据 JSON-LD

#### 阅读体验

- 阅读进度条
- 文章目录吸顶
- 深色模式
- 字号调整
- 代码块复制
- 图片懒加载
- 文章点赞和点踩
- 文章收藏
- 稍后阅读
- 作者主页

#### 运营增长

- RSS 入口
- 邮件订阅入口后续再开启
- 相关推荐
- 热门内容榜单
- 分享卡片
- 站内公告
- 友情链接
- 自定义落地页

### 3.2 后台功能

#### 登录与权限

- 管理员登录
- 普通用户登录
- 多作者账号
- 角色权限控制
- 登录日志
- 操作日志

#### 用户管理

- 用户列表
- 用户搜索
- 用户状态管理
- 角色分配
- 禁言
- 封禁
- 重置密码
- 查看用户评论记录
- 查看用户登录记录

#### 站内信管理

- 后台给单个用户发送站内信
- 按角色或用户状态群发站内信
- 投稿审核、评论回复等系统事件自动生成消息
- 支持未读、已读、归档和删除状态
- 支持消息优先级、发送记录和失败重试
- 支持关闭用户互发私信，先保留后台到用户的运营消息能力

#### 文章管理

- 新建文章
- 编辑文章
- 草稿保存
- 定时发布
- 文章上下架
- 文章归档
- Markdown 编辑
- 封面图上传
- SEO 信息编辑

#### 投稿审核

- 投稿列表
- 投稿详情预览
- 审核通过并发布
- 退回修改
- 拒绝投稿
- 编辑投稿内容
- 添加审核意见
- 查看投稿人历史评论和投稿记录
- 投稿通过后可选择是否将投稿人升级为作者

#### 分类与标签管理

- 分类新增、编辑、删除
- 分类排序
- 标签新增、编辑、删除
- 标签合并
- 分类和标签 slug 管理

#### 评论管理

- 评论列表
- 评论审核
- 评论删除
- 评论置顶
- 评论恢复
- 评论举报处理
- 批量审核
- 按用户、文章、状态筛选
- 屏蔽关键词
- IP 限制

#### 媒体资源管理

- 图片上传
- 文件上传
- 图片压缩
- 资源复用
- CDN 地址管理

#### 数据统计

- PV / UV
- 文章阅读量
- 点赞数
- 点踩数
- 评论数
- 搜索词统计
- 来源渠道统计
- 热门文章排行

#### 系统设置

- 站点名称
- Logo
- 站点描述
- 备案信息
- 社交链接
- 邮件服务配置
- 评论开关
- 投稿开关
- 投稿频率限制
- 投稿说明配置
- 主题配置

#### 内容工作流

- 文章版本历史
- 修改记录对比
- 文章预览链接
- 审核流转
- 退回修改
- 发布审批
- 定时任务管理

#### 页面与导航管理

- 自定义页面
- 顶部导航菜单
- 底部导航菜单
- 友情链接管理
- 重定向规则管理
- 404 页面配置

#### 导入导出与备份

- Markdown 导入
- WordPress 导入
- Ghost 导入
- 文章批量导出
- 媒体资源导出
- 数据库备份
- 备份恢复

## 4. 技术架构

### 4.1 推荐技术选型

面向个人或小团队项目，推荐使用前后端分离架构：

| 模块 | 技术 |
| --- | --- |
| 前台与后台 | Vue 3 + Vite + Vue Router + Pinia |
| 开发语言 | TypeScript + Go |
| UI | Tailwind CSS 或 Naive UI / Element Plus |
| 后端 API | Go + Gin |
| 数据库 | PostgreSQL |
| 数据访问 | GORM 或 sqlc / pgx |
| 缓存 | Redis |
| 队列 | Asynq 或 Go worker + Redis |
| 搜索 | PostgreSQL 全文搜索 |
| 编辑器 | Markdown Editor，例如 Milkdown / ByteMD |
| 鉴权 | Gin 中间件 + Session / JWT |
| 文件存储 | MinIO |
| 部署 | Docker Compose + Nginx |
| 监控 | Sentry |

如果项目需要支持更高并发或更复杂的业务，可以在保持 Vue3 + Go/Gin 主体不变的前提下做服务拆分：

| 模块 | 技术 |
| --- | --- |
| 前台 | Vue 3，可升级 Nuxt 3 或预渲染以强化 SEO |
| 管理后台 | Vue 3 + Naive UI / Element Plus |
| 后端 API | Go + Gin，按模块拆分服务 |
| 数据库 | PostgreSQL / MySQL |
| 缓存 | Redis |
| 搜索 | PostgreSQL 全文搜索，必要时再评估专用搜索服务 |
| 队列 | Asynq / RabbitMQ |
| 对象存储 | MinIO |
| 网关 | Nginx |
| CI/CD | GitHub Actions / GitLab CI |

### 4.2 整体架构图

```text
用户浏览器
   |
   v
CDN / Nginx
   |
   v
Vue3 前台 / 后台静态资源
   |
   +--> 前台页面：CSR + 关键页面预渲染 / 服务端 meta 注入
   +--> 管理后台：CSR
   |
   v
Go + Gin API
   |
   +--> PostgreSQL：文章、用户、评论、分类、标签、全文搜索
   +--> Redis：缓存、限流、会话、热点数据
   +--> Object Storage：图片、附件
   +--> Email Service：验证码、站内信邮件提醒、后续 Newsletter
   +--> Analytics：访问统计、行为统计
   +--> Worker：定时发布、邮件发送、导入导出
```

## 5. 核心模块设计

### 5.1 内容模块

内容模块负责文章的创建、编辑、发布、归档和展示。

文章状态建议包括：

- `draft`：草稿
- `submitted`：用户投稿待审核
- `rejected`：投稿被拒绝或退回
- `scheduled`：定时发布
- `published`：已发布
- `archived`：已归档

文章内容建议使用 Markdown 存储，渲染时转换为 HTML。对于需要组件化内容的场景，可以升级为 MDX。

用户投稿建议复用文章模型，但状态必须隔离：

- 注册用户创建的投稿默认只对本人和后台审核人员可见
- 注册用户只能编辑自己的 `draft`、`submitted`、`rejected` 投稿
- 投稿进入 `submitted` 后，用户不能直接发布
- 编辑或管理员审核通过后，将投稿状态改为 `published`
- 审核退回时状态改为 `rejected`，并写入审核意见
- 投稿发布后，保留投稿人与审核人的记录

文章互动建议：

- 登录用户可以对已发布文章点赞或点踩
- 同一用户对同一篇文章只能保留一个反馈状态：`like` 或 `dislike`
- 用户重复点击当前反馈时取消反馈，切换反馈时更新原记录
- 文章表保留 `like_count` 和 `dislike_count` 冗余计数字段，列表和详情页直接读取
- 反馈接口需要限流，并记录审计上下文，避免刷赞或恶意点踩

### 5.2 用户与权限模块

使用 RBAC 权限模型：

- 用户属于一个或多个角色
- 角色绑定权限
- 后台接口根据权限控制访问

用户系统建议支持：

- 邮箱注册
- 邮箱验证
- 密码登录
- 找回密码
- 第三方登录，可选
- Session 或 JWT 鉴权
- 用户资料维护
- 用户头像
- 用户收藏文章
- 用户点赞或点踩文章
- 用户评论记录
- 站内信与消息通知
- 账号注销

用户状态建议包括：

- `pending`：待验证
- `active`：正常
- `muted`：禁言
- `banned`：封禁
- `deleted`：已注销

评论权限建议：

- MVP 默认注册用户才能评论
- 访客评论作为站点配置项，默认关闭
- 被禁言用户不能评论和回复
- 被封禁用户不能登录和互动
- 作者、编辑、管理员可以管理评论
- 用户可以删除自己的评论，但删除后保留审计记录

投稿权限建议：

- 注册用户可以创建投稿草稿
- 注册用户可以提交自己的投稿审核
- 注册用户不能直接发布文章
- 注册用户不能修改已发布投稿
- 编辑和管理员可以审核投稿
- 编辑和管理员可以编辑投稿内容后发布
- 管理员可以将持续高质量投稿用户升级为作者

典型权限包括：

- `post:create`
- `post:update`
- `post:delete`
- `post:publish`
- `post:react`
- `submission:create`
- `submission:update:own`
- `submission:submit`
- `submission:review`
- `submission:publish`
- `comment:create`
- `comment:reply`
- `comment:delete:own`
- `comment:moderate`
- `message:read:own`
- `message:create`
- `message:broadcast`
- `category:manage`
- `tag:manage`
- `user:manage`
- `setting:manage`

### 5.3 评论模块

评论模块需要重点处理安全和内容质量：

- 评论默认进入待审核状态
- 管理员可配置是否开启自动审核
- 登录用户可发表评论和回复
- 用户可删除自己的评论
- 支持评论点赞
- 支持评论举报
- 支持评论回复通知
- 支持作者回复标识
- 对提交评论接口做频率限制
- 对评论内容做 XSS 清理
- 支持敏感词过滤
- 支持按 IP、用户、文章维度限制

评论状态建议包括：

- `pending`：待审核
- `approved`：已通过
- `rejected`：已拒绝
- `spam`：垃圾评论
- `deleted`：用户或管理员删除

评论发布流程：

1. 用户登录后提交评论。
2. 系统进行频率限制、敏感词检测和 XSS 清理。
3. 如果开启自动审核，评论进入 `approved` 状态。
4. 如果关闭自动审核，评论进入 `pending` 状态。
5. 评论通过后通知文章作者和被回复用户。

评论展示规则：

- 默认只展示 `approved` 评论
- 用户可看到自己提交的待审核评论
- 管理员可看到所有状态评论
- 被删除评论在楼层中显示为“该评论已删除”
- 评论按时间或热度排序

### 5.4 搜索模块

搜索模块直接使用 PostgreSQL 全文搜索，减少系统依赖和运维成本。

推荐实现方式：

- 在文章表或独立搜索表维护 `tsvector` 字段
- 对 `tsvector` 建 GIN 索引
- 标题、摘要、正文和标签分别设置权重
- 使用 `websearch_to_tsquery` 或 `plainto_tsquery` 处理关键词
- 搜索结果按相关性、发布时间和热度综合排序
- 文章发布、更新、下架时同步更新全文索引字段
- 搜索结果支持标题、摘要、正文和标签加权

只有当数据量、复杂分词、多语言检索或搜索分析能力明显超过 PostgreSQL 能力边界时，再评估专用搜索服务。

### 5.5 媒体模块

媒体资源不建议直接长期存储在应用服务器本地磁盘。

推荐方案：

- 图片上传到对象存储
- 通过 CDN 分发
- 上传时校验文件类型和大小
- 对图片生成缩略图
- 保存媒体资源元数据到数据库

### 5.6 统计模块

统计模块可以分为基础统计和增强统计。

基础统计：

- 文章阅读量
- 点赞数
- 点踩数
- 评论数
- 热门文章

增强统计：

- PV / UV
- 来源渠道
- 搜索词
- 用户访问路径
- 设备类型

阅读量建议使用 Redis 缓冲写入，定期批量同步到数据库，降低数据库写压力。

### 5.7 内容工作流模块

成熟博客系统通常不是直接编辑后立即覆盖线上内容，需要支持完整内容生命周期：

- 自动保存草稿
- 文章版本历史
- 版本对比和回滚
- 发布前预览
- 审核、退回、发布审批
- 定时发布和定时下架
- 作者、编辑、管理员之间的协作记录

预览链接应使用带过期时间的 token，避免未发布内容被公开访问。

### 5.8 投稿审核模块

投稿审核模块用于支持“登录用户 -> 投稿 -> 待审核 -> 通过后发布”的社区贡献流程。

前台投稿流程：

1. 登录用户进入投稿页。
2. 填写标题、摘要、正文、分类、标签和封面图。
3. 系统自动保存为投稿草稿。
4. 用户预览内容后提交审核。
5. 投稿进入 `submitted` 状态。
6. 用户可在个人中心查看审核进度和审核意见。

后台审核流程：

1. 编辑或管理员查看投稿列表。
2. 打开投稿详情，检查内容质量、格式、图片和 SEO 信息。
3. 可选择通过发布、退回修改或拒绝投稿。
4. 通过后投稿变成正式文章，并进入公开文章列表。
5. 退回或拒绝时必须填写审核意见。
6. 系统通知投稿人审核结果。

投稿审核规则：

- 投稿默认不进入搜索索引
- 投稿默认不生成公开页面
- 投稿通过后才触发缓存刷新、搜索索引同步、RSS 更新和订阅推送
- 被禁言或封禁用户不能投稿
- 短时间多次投稿需要限流
- 管理员可关闭投稿入口

### 5.9 站内信模块

站内信用于承载站内事件提醒和后台运营消息。它和邮件、Web Push 不同，优先保证登录用户在站内可追溯地看到重要事件。

建议分阶段实现：

- 第一阶段：系统消息和管理员消息，覆盖评论回复、评论审核、投稿审核、账号状态变更和站点公告。
- 第二阶段：后台群发，支持按角色、用户状态、注册时间和活跃度筛选收件人。
- 第三阶段：可选开放用户互发私信，但默认关闭，避免个人博客早期承担聊天、举报和骚扰治理成本。

用户侧能力：

- 收件箱展示全部、未读、系统、审核和管理员消息
- 消息详情展示发送人、时间、正文和关联文章或投稿
- 支持标记已读、全部已读、归档和删除
- 顶部导航显示未读数量
- 重要审核消息可在我的投稿中同步展示

后台侧能力：

- 给指定用户发送站内信
- 按角色或筛选条件批量发送
- 设置消息类型、优先级和跳转目标
- 查看发送记录、送达人数、已读率和失败数
- 支持定时发送和撤回未读消息，撤回需要记录审计日志

消息类型建议：

- `system`：系统事件，例如评论通过、评论回复、账号状态变更
- `submission`：投稿审核结果和退回修改意见
- `announcement`：站点公告或运营通知
- `admin`：管理员给用户的定向消息
- `private`：用户互发私信，后续可选

### 5.10 主题与导航模块

参考 WordPress 和 Ghost，主题和导航需要从业务数据中解耦。

建议设计：

- 主题配置独立存储
- 支持浅色 / 深色模式
- 支持站点 Logo、字体、主色、布局配置
- 支持顶部导航、底部导航和社交链接配置
- 支持自定义页面和落地页
- 支持友情链接、关于页、归档页

早期可以只做内置主题配置，不急于实现完整插件市场。

### 5.11 订阅与 Newsletter 模块

参考 Ghost 的会员和 Newsletter 设计，订阅能力后置实现，MVP 只保留 RSS：

- 第一阶段支持 RSS
- 第二阶段支持邮箱订阅和退订
- 第三阶段支持新文章邮件推送、订阅分组和专栏订阅
- 后续可扩展会员、付费订阅和内容权限

邮件发送建议通过异步队列处理，避免发布文章时阻塞主流程。

### 5.12 导入导出与迁移模块

成熟博客系统需要考虑迁移成本，避免内容被平台锁定。

建议支持：

- Markdown 文件导入导出
- WordPress XML 导入
- Ghost JSON 导入
- 媒体资源批量导出
- 站点配置导出
- 数据库备份恢复

导入任务应记录执行状态、成功数量、失败数量和失败原因，方便重试。

### 5.13 审计与操作日志模块

后台的关键操作需要可追踪：

- 登录成功和失败
- 创建、修改、删除文章
- 发布、下架、归档文章
- 审核评论
- 修改系统配置
- 修改用户权限
- 执行导入导出和备份恢复

审计日志不建议允许普通后台用户删除，只能按保留周期归档。

## 6. 数据库设计

### 6.1 用户表 users

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| username | varchar | 用户名 |
| display_name | varchar | 展示名称 |
| email | varchar | 邮箱 |
| password_hash | varchar | 密码哈希 |
| avatar | varchar | 头像地址 |
| bio | text | 个人简介 |
| role | varchar | 角色 |
| status | varchar | 状态 |
| email_verified_at | timestamp | 邮箱验证时间 |
| last_login_at | timestamp | 最近登录时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.2 文章表 posts

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| title | varchar | 标题 |
| slug | varchar | URL 标识 |
| summary | text | 摘要 |
| content | text | Markdown 正文 |
| cover_image | varchar | 封面图 |
| status | varchar | 状态 |
| visibility | varchar | 可见性：公开、私密、会员 |
| author_id | uuid | 作者 ID |
| submitter_id | uuid | 投稿人 ID，可为空 |
| category_id | uuid | 分类 ID |
| view_count | int | 阅读量 |
| like_count | int | 点赞数 |
| dislike_count | int | 点踩数 |
| comment_count | int | 评论数 |
| reading_time | int | 预计阅读分钟数 |
| is_featured | boolean | 是否推荐 |
| allow_comment | boolean | 是否允许评论 |
| seo_title | varchar | SEO 标题 |
| seo_description | text | SEO 描述 |
| canonical_url | varchar | 规范链接 |
| search_vector | tsvector | 全文搜索向量，建议建立 GIN 索引 |
| review_note | text | 审核意见 |
| reviewed_by | uuid | 审核人 ID |
| reviewed_at | timestamp | 审核时间 |
| submitted_at | timestamp | 提交审核时间 |
| published_at | timestamp | 发布时间 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.3 分类表 categories

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| name | varchar | 分类名称 |
| slug | varchar | URL 标识 |
| description | text | 分类描述 |
| sort_order | int | 排序 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.4 标签表 tags

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| name | varchar | 标签名称 |
| slug | varchar | URL 标识 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.5 文章标签关联表 post_tags

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| post_id | uuid | 文章 ID |
| tag_id | uuid | 标签 ID |

### 6.6 评论表 comments

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| post_id | uuid | 文章 ID |
| user_id | uuid | 用户 ID，可为空 |
| parent_id | uuid | 父评论 ID，可为空 |
| content | text | 评论内容 |
| status | varchar | 状态 |
| like_count | int | 点赞数 |
| reply_count | int | 回复数 |
| is_author_reply | boolean | 是否作者回复 |
| deleted_at | timestamp | 删除时间 |
| reviewed_by | uuid | 审核人 ID |
| reviewed_at | timestamp | 审核时间 |
| ip_address | varchar | IP 地址 |
| user_agent | text | 浏览器信息 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.7 媒体资源表 media_assets

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| filename | varchar | 文件名 |
| url | varchar | 访问地址 |
| alt_text | varchar | 图片替代文本 |
| mime_type | varchar | 文件类型 |
| size | int | 文件大小 |
| width | int | 图片宽度 |
| height | int | 图片高度 |
| uploader_id | uuid | 上传人 ID |
| created_at | timestamp | 创建时间 |

### 6.8 系统设置表 site_settings

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| key | varchar | 配置键 |
| value | jsonb | 配置值 |
| updated_at | timestamp | 更新时间 |

### 6.9 文章版本表 post_revisions

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| post_id | uuid | 文章 ID |
| title | varchar | 版本标题 |
| content | text | 版本正文 |
| summary | text | 版本摘要 |
| editor_id | uuid | 编辑人 ID |
| change_note | text | 修改说明 |
| created_at | timestamp | 创建时间 |

### 6.10 订阅者表 subscribers

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| email | varchar | 邮箱 |
| name | varchar | 昵称 |
| status | varchar | 状态 |
| source | varchar | 来源 |
| confirmed_at | timestamp | 确认订阅时间 |
| unsubscribed_at | timestamp | 退订时间 |
| created_at | timestamp | 创建时间 |

### 6.11 导航菜单表 navigation_items

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| label | varchar | 展示名称 |
| url | varchar | 链接地址 |
| position | varchar | 位置：header、footer |
| sort_order | int | 排序 |
| is_external | boolean | 是否外链 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.12 重定向表 redirects

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| source_path | varchar | 原路径 |
| target_path | varchar | 目标路径 |
| status_code | int | HTTP 状态码 |
| enabled | boolean | 是否启用 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.13 审计日志表 audit_logs

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| actor_id | uuid | 操作人 ID |
| action | varchar | 操作类型 |
| target_type | varchar | 操作对象类型 |
| target_id | uuid | 操作对象 ID |
| metadata | jsonb | 操作上下文 |
| ip_address | varchar | IP 地址 |
| created_at | timestamp | 创建时间 |

### 6.14 任务表 jobs

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| type | varchar | 任务类型 |
| status | varchar | 任务状态 |
| payload | jsonb | 任务参数 |
| result | jsonb | 执行结果 |
| error_message | text | 失败原因 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 6.15 用户会话表 user_sessions

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| user_id | uuid | 用户 ID |
| session_token_hash | varchar | 会话 token 哈希 |
| ip_address | varchar | IP 地址 |
| user_agent | text | 浏览器信息 |
| expires_at | timestamp | 过期时间 |
| created_at | timestamp | 创建时间 |

### 6.16 密码重置表 password_reset_tokens

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| user_id | uuid | 用户 ID |
| token_hash | varchar | 重置 token 哈希 |
| expires_at | timestamp | 过期时间 |
| used_at | timestamp | 使用时间 |
| created_at | timestamp | 创建时间 |

### 6.17 用户收藏表 user_bookmarks

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| user_id | uuid | 用户 ID |
| post_id | uuid | 文章 ID |
| created_at | timestamp | 收藏时间 |

### 6.18 文章反馈表 post_reactions

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| user_id | uuid | 用户 ID |
| post_id | uuid | 文章 ID |
| type | varchar | 反馈类型：like / dislike |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

建议对 `(user_id, post_id)` 建唯一索引，确保同一用户对同一文章只有一个有效反馈。

### 6.19 评论点赞表 comment_reactions

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| user_id | uuid | 用户 ID |
| comment_id | uuid | 评论 ID |
| type | varchar | 反应类型，默认 like |
| created_at | timestamp | 创建时间 |

### 6.20 评论举报表 comment_reports

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| comment_id | uuid | 评论 ID |
| reporter_id | uuid | 举报人 ID |
| reason | varchar | 举报原因 |
| description | text | 补充说明 |
| status | varchar | 处理状态 |
| handled_by | uuid | 处理人 ID |
| handled_at | timestamp | 处理时间 |
| created_at | timestamp | 创建时间 |

### 6.21 通知表 notifications

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| user_id | uuid | 接收用户 ID |
| type | varchar | 通知类型 |
| title | varchar | 标题 |
| content | text | 内容 |
| target_type | varchar | 目标类型 |
| target_id | uuid | 目标 ID |
| read_at | timestamp | 阅读时间 |
| created_at | timestamp | 创建时间 |

### 6.22 站内信表 site_messages

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| sender_id | uuid | 发送人 ID，系统消息可为空 |
| created_by | uuid | 后台创建人 ID |
| message_type | varchar | 消息类型，system / submission / announcement / admin / private |
| priority | varchar | 优先级，normal / important / urgent |
| title | varchar | 标题 |
| content | text | 正文 |
| target_type | varchar | 关联目标类型，例如 post / comment / submission / user |
| target_id | uuid | 关联目标 ID |
| send_scope | varchar | 发送范围，single / role / filtered / broadcast |
| scheduled_at | timestamp | 定时发送时间 |
| sent_at | timestamp | 实际发送时间 |
| revoked_at | timestamp | 撤回时间 |
| created_at | timestamp | 创建时间 |

### 6.23 站内信收件人表 site_message_recipients

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| message_id | uuid | 站内信 ID |
| recipient_id | uuid | 接收用户 ID |
| delivery_status | varchar | 送达状态，pending / delivered / failed |
| read_at | timestamp | 阅读时间 |
| archived_at | timestamp | 归档时间 |
| deleted_at | timestamp | 用户侧删除时间 |
| created_at | timestamp | 创建时间 |

## 7. API 设计

### 7.1 前台 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| POST | `/api/auth/logout` | 退出登录 |
| GET | `/api/auth/me` | 获取当前用户 |
| POST | `/api/auth/verify-email` | 邮箱验证 |
| POST | `/api/auth/forgot-password` | 发送找回密码邮件 |
| POST | `/api/auth/reset-password` | 重置密码 |
| GET | `/api/posts` | 获取文章列表 |
| GET | `/api/posts/:slug` | 获取文章详情 |
| GET | `/api/categories` | 获取分类列表 |
| GET | `/api/categories/:slug/posts` | 获取分类下文章 |
| GET | `/api/tags` | 获取标签列表 |
| GET | `/api/tags/:slug/posts` | 获取标签下文章 |
| GET | `/api/search` | 搜索文章 |
| GET | `/api/posts/:id/reaction` | 获取当前用户对文章的反馈 |
| PUT | `/api/posts/:id/reaction` | 设置文章点赞或点踩 |
| DELETE | `/api/posts/:id/reaction` | 取消文章点赞或点踩 |
| POST | `/api/posts/:id/bookmark` | 收藏文章 |
| GET | `/api/posts/:id/comments` | 获取文章评论 |
| POST | `/api/posts/:id/comments` | 提交文章评论 |
| POST | `/api/comments/:id/replies` | 回复评论 |
| DELETE | `/api/comments/:id` | 删除自己的评论 |
| POST | `/api/comments/:id/like` | 点赞评论 |
| POST | `/api/comments/:id/report` | 举报评论 |
| GET | `/api/me/comments` | 获取我的评论 |
| GET | `/api/me/bookmarks` | 获取我的收藏 |
| GET | `/api/me/submissions` | 获取我的投稿 |
| POST | `/api/submissions` | 创建投稿草稿 |
| PUT | `/api/submissions/:id` | 更新自己的投稿 |
| POST | `/api/submissions/:id/submit` | 提交投稿审核 |
| DELETE | `/api/submissions/:id` | 删除自己的未发布投稿 |
| GET | `/api/me/notifications` | 获取我的通知 |
| GET | `/api/me/messages` | 获取我的站内信 |
| GET | `/api/me/messages/:id` | 获取站内信详情 |
| POST | `/api/me/messages/:id/read` | 标记站内信已读 |
| POST | `/api/me/messages/read-all` | 全部标记已读 |
| POST | `/api/me/messages/:id/archive` | 归档站内信 |
| DELETE | `/api/me/messages/:id` | 删除自己的站内信 |
| POST | `/api/messages` | 发送用户私信，可选开启 |
| PUT | `/api/me/profile` | 更新个人资料 |
| POST | `/api/me/avatar` | 上传或更换头像 |
| PUT | `/api/me/password` | 修改当前账号密码 |
| GET | `/api/me/sessions` | 获取当前账号登录设备 |
| DELETE | `/api/me/sessions/:id` | 移除指定登录设备 |
| POST | `/api/me/export` | 创建个人数据导出任务 |
| DELETE | `/api/me` | 申请注销当前账号 |
| GET | `/api/navigation` | 获取导航菜单 |
| POST | `/api/subscribers` | 邮箱订阅，后续可选开启 |
| POST | `/api/subscribers/unsubscribe` | 退订，后续可选开启 |
| GET | `/api/sitemap.xml` | 站点地图 |

文章详情页按 `slug` 访问，但详情响应必须返回文章 `id`，供点赞、点踩、收藏和评论接口使用。

### 7.2 后台 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/admin/auth/login` | 管理员登录 |
| GET | `/api/admin/users` | 获取用户列表 |
| GET | `/api/admin/users/export` | 导出用户数据 |
| POST | `/api/admin/users/invitations` | 邀请作者或编辑 |
| GET | `/api/admin/users/:id` | 获取用户详情 |
| PUT | `/api/admin/users/:id/status` | 更新用户状态 |
| PUT | `/api/admin/users/:id/role` | 更新用户角色 |
| POST | `/api/admin/users/:id/password-reset` | 发送重置密码链接 |
| GET | `/api/admin/users/:id/sessions` | 查看用户登录设备 |
| GET | `/api/admin/posts` | 获取文章列表 |
| POST | `/api/admin/posts` | 创建文章 |
| PUT | `/api/admin/posts/:id` | 更新文章 |
| DELETE | `/api/admin/posts/:id` | 删除文章 |
| POST | `/api/admin/posts/:id/publish` | 发布文章 |
| POST | `/api/admin/posts/:id/preview` | 生成预览链接 |
| GET | `/api/admin/posts/:id/revisions` | 获取文章版本 |
| POST | `/api/admin/posts/:id/revisions/:revisionId/restore` | 恢复文章版本 |
| GET | `/api/admin/submissions` | 获取投稿列表 |
| GET | `/api/admin/submissions/:id` | 获取投稿详情 |
| PUT | `/api/admin/submissions/:id` | 编辑投稿内容 |
| POST | `/api/admin/submissions/:id/approve` | 审核通过并发布 |
| POST | `/api/admin/submissions/:id/reject` | 拒绝或退回投稿 |
| GET | `/api/admin/comments` | 获取评论列表 |
| PUT | `/api/admin/comments/:id/status` | 更新评论状态 |
| DELETE | `/api/admin/comments/:id` | 删除评论 |
| GET | `/api/admin/comment-reports` | 获取评论举报列表 |
| PUT | `/api/admin/comment-reports/:id/status` | 处理评论举报 |
| GET | `/api/admin/messages` | 获取站内信发送记录 |
| POST | `/api/admin/messages` | 给指定用户发送站内信 |
| POST | `/api/admin/messages/broadcast` | 按范围群发站内信 |
| POST | `/api/admin/messages/:id/revoke` | 撤回未读站内信 |
| GET | `/api/admin/messages/:id/statistics` | 获取站内信送达与已读统计 |
| GET | `/api/admin/media` | 获取媒体资源列表 |
| POST | `/api/admin/media` | 上传媒体资源 |
| GET | `/api/admin/media/:id` | 获取媒体资源详情 |
| PATCH | `/api/admin/media/:id` | 更新媒体资源元数据 |
| DELETE | `/api/admin/media/:id` | 删除未使用媒体资源 |
| GET | `/api/admin/navigation` | 获取导航菜单 |
| PUT | `/api/admin/navigation` | 更新导航菜单 |
| GET | `/api/admin/redirects` | 获取重定向规则 |
| POST | `/api/admin/redirects` | 创建重定向规则 |
| POST | `/api/admin/import` | 创建导入任务 |
| POST | `/api/admin/export` | 创建导出任务 |
| GET | `/api/admin/jobs/:id` | 获取任务状态 |
| GET | `/api/admin/audit-logs` | 获取审计日志 |
| GET | `/api/admin/statistics` | 获取统计数据 |
| GET | `/api/admin/settings` | 获取系统配置 |
| PUT | `/api/admin/settings` | 更新系统配置 |

## 8. 页面路由设计

### 8.1 前台路由

| 路径 | 页面 |
| --- | --- |
| `/` | 首页 |
| `/posts` | 文章列表 |
| `/posts/[slug]` | 文章详情 |
| `/categories` | 分类列表 |
| `/categories/[slug]` | 分类详情 |
| `/tags` | 标签列表 |
| `/tags/[slug]` | 标签详情 |
| `/topics` | 专题列表 |
| `/topics/[slug]` | 专题详情 |
| `/search` | 搜索页 |
| `/login` | 登录页 |
| `/register` | 注册页 |
| `/account` | 个人中心 |
| `/account/comments` | 我的评论 |
| `/account/bookmarks` | 我的收藏 |
| `/account/submissions` | 我的投稿 |
| `/account/messages` | 站内信 |
| `/account/settings` | 账号设置 |
| `/submit` | 用户投稿 |
| `/about` | 关于页 |
| `/archive` | 归档页 |
| `/authors/[id]` | 作者主页 |
| `/preview/[token]` | 文章预览页 |
| `/rss.xml` | RSS |

### 8.2 后台路由

| 路径 | 页面 |
| --- | --- |
| `/admin/login` | 后台登录 |
| `/admin` | 后台概览 |
| `/admin/users` | 用户管理 |
| `/admin/posts` | 文章管理 |
| `/admin/submissions` | 投稿审核 |
| `/admin/posts/new` | 新建文章 |
| `/admin/posts/[id]/edit` | 编辑文章 |
| `/admin/categories` | 分类管理 |
| `/admin/tags` | 标签管理 |
| `/admin/comments` | 评论管理 |
| `/admin/comment-reports` | 评论举报管理 |
| `/admin/messages` | 站内信管理 |
| `/admin/media` | 媒体管理 |
| `/admin/navigation` | 导航管理 |
| `/admin/statistics` | 数据统计 |
| `/admin/redirects` | 重定向管理 |
| `/admin/import-export` | 导入导出 |
| `/admin/audit-logs` | 审计日志 |
| `/admin/settings` | 系统设置 |

### 8.3 原型覆盖状态

当前静态原型已覆盖：

- 前台：首页、归档、专题、文章详情、投稿、登录、个人中心概览、我的评论、我的收藏、我的投稿、站内信、账号设置
- 后台：文章管理、写作编辑器、投稿审核、评论管理、用户管理、站内信管理、媒体库、导航管理、数据统计、系统设置

后续进入详细设计前建议继续补齐：

- `/admin`：后台概览仪表盘，可以复用统计核心指标，但需要增加待办、最近操作和系统状态
- `/admin/categories`、`/admin/tags`：分类与标签管理，支持 slug、排序、合并、删除前影响检查
- `/admin/redirects`：重定向管理，当前只在导航页中展示摘要，不足以覆盖批量规则管理
- `/admin/import-export`：导入导出任务页，展示任务队列、进度、失败原因和下载入口
- `/admin/audit-logs`：审计日志页，支持按操作者、动作、对象和时间筛选
- `/search`、`/about`、`/authors/[id]`、`/preview/[token]`：前台独立页面原型

## 9. 缓存与性能设计

### 9.1 页面渲染策略

- 首页：Vue3 构建静态资源，首屏数据由 Go API 提供，可按需预渲染
- 文章详情页：建议由 Go 注入基础 HTML、SEO meta 和结构化数据，再交给 Vue 激活交互
- 分类页和标签页：客户端渲染为主，可对热门分类做预渲染
- 搜索页：服务端查询或客户端请求 API
- 后台管理页：CSR

### 9.2 缓存策略

- 文章详情缓存
- 热门文章缓存
- 分类和标签缓存
- 搜索结果短时间缓存
- 阅读量使用 Redis 缓冲
- 图片和静态资源使用 CDN 缓存
- RSS、Sitemap 等低频更新内容使用短周期缓存
- 后台接口默认不缓存，避免权限数据泄露

### 9.3 缓存失效

以下操作需要触发缓存刷新：

- 发布文章
- 更新文章
- 删除文章
- 修改分类
- 修改标签
- 修改站点配置
- 修改导航菜单
- 修改重定向规则
- 导入文章或批量更新文章

### 9.4 性能指标

建议把性能指标作为验收标准：

- 首页首屏加载时间小于 2 秒
- 文章详情页 Lighthouse Performance 大于 90
- 文章详情页 LCP 小于 2.5 秒
- 图片使用懒加载和响应式尺寸
- 代码块、评论等非首屏模块延迟加载
- 静态资源启用 gzip 或 brotli 压缩
- 数据库慢查询需要记录和优化

### 9.5 SEO 与可访问性

- 每篇文章必须有唯一 `slug`
- 文章更新时保留历史 URL 或配置 301 重定向
- 支持 canonical URL，避免重复内容
- 自动生成 Sitemap
- 自动生成 RSS
- 图片必须支持 alt 文本
- 页面语义结构使用正确的 h1、h2、article、nav、main 标签
- 深色模式需要保证文本对比度
- 交互控件需要支持键盘访问

## 10. 安全设计

- 密码使用 bcrypt 或 argon2 哈希
- 邮箱验证链接和密码重置链接需要设置过期时间
- 会话 token 只保存哈希值
- 登录失败次数过多时触发临时锁定或验证码
- 后台接口必须鉴权
- 管理员账号支持双因素认证
- 管理后台启用 CSRF 保护
- 评论、文章渲染内容需要做 XSS 清理
- 投稿正文、摘要、标题和图片说明需要做 XSS 清理
- 页面启用 CSP 安全策略
- 上传文件限制类型和大小
- 上传文件重命名，避免路径注入
- 登录、评论、文章互动接口启用限流
- 投稿创建和提交审核接口启用限流
- 评论举报接口启用限流
- 禁言和封禁状态需要在评论接口强制校验
- 禁言和封禁状态需要在投稿接口强制校验
- 文章预览链接需要设置过期时间
- 邮箱订阅需要确认机制，退订链接需要一次性 token
- 敏感操作记录审计日志
- 数据库连接使用最小权限账号
- 生产环境使用 HTTPS

### 10.1 隐私与合规

- 明示统计脚本和 Cookie 使用情况
- 订阅邮件必须包含退订入口
- 支持删除订阅者邮箱
- 避免在日志中记录明文密码、token 和完整个人隐私数据
- 备份文件需要加密存储
- 审计日志设置合理保留周期

## 11. 部署设计

### 11.1 基础部署

```text
Nginx
   |
   +--> Vue3 Static Web
   |
   +--> Go Gin API
   |
   +--> Static Assets / CDN

PostgreSQL
Redis
Object Storage
Worker
```

### 11.2 Docker Compose 服务

建议拆分以下服务：

- `web`：Vue3 前台和后台静态资源
- `api`：Go + Gin API 服务
- `postgres`：数据库
- `redis`：缓存
- `worker`：Go 后台任务服务
- `nginx`：反向代理

### 11.3 CI/CD

推荐流水线：

1. 提交代码到 Git 仓库
2. 执行 lint 和测试
3. 构建 Docker 镜像
4. 推送镜像到镜像仓库
5. 部署到服务器
6. 执行数据库迁移
7. 健康检查

### 11.4 备份与恢复

备份策略：

- PostgreSQL 每日全量备份
- 关键业务表支持按小时增量备份
- 对象存储开启版本管理
- 备份文件加密存储
- 至少保留最近 7 天备份

恢复策略：

- 定期演练数据库恢复
- 定期校验备份文件可用性
- 建立恢复时间目标 RTO
- 建立恢复点目标 RPO
- 恢复过程需要记录审计日志

## 12. 日志与监控

### 12.1 日志

- 应用访问日志
- API 错误日志
- 登录日志
- 后台操作日志
- 任务执行日志

### 12.2 监控

- 应用错误监控：Sentry
- 服务可用性监控
- 数据库连接数监控
- Redis 使用率监控
- 接口响应时间监控
- 磁盘和内存使用率监控

### 12.3 告警

- 应用 5xx 错误率异常
- 数据库连接池耗尽
- Redis 不可用
- PostgreSQL 全文搜索慢查询异常
- 邮件发送失败率异常
- 定时发布任务失败
- 磁盘空间不足
- 备份任务失败

## 13. 开发迭代规划

### 第一阶段：MVP

- 首页
- 文章列表
- 文章详情
- 分类和标签
- 后台登录
- 用户注册登录
- 文章发布
- Markdown 编辑
- 登录用户评论
- 评论审核
- 基础 SEO
- 基础导航配置
- 基础备份脚本

### 第二阶段：增强体验

- 评论
- 评论回复
- 评论点赞
- 文章点赞和点踩
- 评论举报
- 搜索
- 图片上传
- 文章草稿
- 登录用户投稿
- 投稿审核
- 文章版本历史
- 定时发布
- 阅读量统计
- RSS
- 代码块复制
- 深色模式

### 第三阶段：平台化

- 多作者
- 权限管理
- 用户管理
- 站内信
- 消息通知
- 投稿人激励和优质投稿人升级为作者
- 数据分析
- 邮件订阅，后续可选
- Newsletter 推送，后续可选
- 主题配置
- 导入导出
- 重定向管理
- 全文搜索优化
- 相关文章推荐

### 第四阶段：生态化

- 会员体系
- 付费订阅
- 插件机制
- 主题市场
- 多语言内容
- Headless API
- Webhook 集成
- 高级内容审核工作流

## 14. 推荐落地方案

如果目标是快速上线并保持后续扩展能力，建议采用：

```text
Vue 3 + Vite + TypeScript + Vue Router + Pinia
Go + Gin
PostgreSQL 全文搜索 + GORM 或 sqlc / pgx
Redis
Asynq 或 Go worker
S3 / OSS 对象存储
Docker + Nginx
```

该方案具备以下优势：

- 开发效率高
- 架构边界清晰，前后端职责明确
- 部署方式成熟
- 数据结构清晰
- 使用 PostgreSQL 承担全文搜索，减少早期运维复杂度
- 后续可平滑扩展评论、订阅和多作者能力

## 15. 后续可扩展方向

- 主题市场
- 专栏体系
- 会员订阅
- 内容付费
- AI 摘要
- AI 标签推荐
- 自动生成文章目录
- 多语言内容
- Web Push 通知
- 评论反垃圾策略
- 内容审核工作流

## 16. 成熟系统补充项优先级

结合 WordPress、Ghost、Medium、Hashnode、Hugo、Jekyll 等成熟系统经验，当前方案建议按以下优先级补充：

| 优先级 | 补充项 | 原因 |
| --- | --- | --- |
| P0 | 文章预览、草稿、定时发布、基础 SEO、Sitemap、RSS | 博客上线的基础能力，直接影响发布流程和搜索收录 |
| P0 | 数据库备份、对象存储、上传校验、XSS 清理、限流 | 保障内容资产和系统安全 |
| P1 | 文章版本历史、重定向管理、导航管理、媒体库、审计日志 | 提升长期运营可维护性，避免内容误操作和链接失效 |
| P1 | 搜索、评论审核、用户投稿、统计分析 | 提升用户体验、内容供给和运营能力 |
| P2 | Newsletter、会员订阅、多作者协作、主题配置 | 支持平台化和商业化 |
| P2 | 导入导出、Headless API、Webhook、插件机制 | 支持生态扩展和系统迁移 |

### 16.1 推荐实现顺序

1. 先完成文章、分类、标签、后台登录、Markdown 编辑、SEO、RSS 和 Sitemap。
2. 再完成媒体库、评论、搜索、阅读量、导航管理和基础统计。
3. 然后补充用户投稿、投稿审核、文章版本历史、预览链接、重定向管理、审计日志和备份恢复。
4. 最后考虑 Newsletter、会员体系、多作者协作、主题市场和插件机制。

### 16.2 暂不建议早期实现的能力

以下能力复杂度较高，早期不建议作为 MVP 范围：

- 完整插件市场
- 复杂推荐算法
- 付费内容和订单系统
- 多租户 SaaS
- 可视化页面搭建器
- 大规模工作流审批引擎

这些能力需要稳定的内容模型、权限模型和运营数据作为基础，适合在系统已有真实使用反馈后再设计。
