# Blog

一个面向个人/团队内容发布的现代博客系统。前台使用 Vue 3 + Vite，后端使用 Go + Gin，支持文章、专题、评论、投稿、媒体库、站内信、后台管理、SEO 页面和 Docker 部署。

[演示连接](https://blog.jecyai.com/)

## 功能概览

- 前台内容：主页、归档、专题、搜索、作者页、文章详情、阅读进度、文章反馈、收藏和评论。
- Markdown 写作：`md-editor-v3` 编辑器、实时预览、代码高亮、图片上传、查看页主题/代码主题切换。
- 投稿工作流：登录用户投稿，后台审核、退回、发布、下架和重新上架。
- 后台管理：概览、文章、投稿、分类标签、专题、评论、用户、站内信、媒体库、导航、重定向、导入导出、统计、审计日志、系统设置。
- 用户中心：账号设置、邮箱验证、登录设备、我的评论、收藏、私密文章、投稿、站内信。
- 媒体存储：本地 `/uploads` 或 MinIO；上传对象路径按 `YYYY/MM/DD` 分目录。
- 部署能力：SQLite 本地开发，PostgreSQL/Redis/MinIO 生产部署，Docker Compose 编排。

## 技术栈

- Web: Vue 3, Vite, TypeScript, Pinia, Vue Router, Element Plus, md-editor-v3
- API: Go 1.24, Gin, PostgreSQL/SQLite, Redis, MinIO SDK
- Infra: Docker Compose, PostgreSQL 16, Redis 7, MinIO

## 项目结构

```text
api/                     Go API 服务
  cmd/server/            服务入口、数据库初始化
  internal/config/       环境变量配置
  internal/database/     PostgreSQL/SQLite 迁移
  internal/modules/      业务模块
web/                     Vue 前端
  src/app/               应用壳、导航和布局
  src/components/        通用组件
  src/pages/             前台、账号和后台页面
  src/shared/            API、Markdown、主题等共享逻辑
deploy/                  Docker Compose 和运维脚本
docs/                    原型和设计参考
```

## 本地开发

### 1. 启动基础设施

只需要 PostgreSQL/Redis 时：

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis
```

如果要本地联调 MinIO：

```bash
docker compose -f deploy/docker-compose.yml up -d minio minio-init
```

MinIO 控制台默认地址：`http://localhost:9001`。

### 2. 配置 API

复制示例配置：

```bash
cp api/.env.example api/.env
```

本地默认可以直接使用 SQLite：

```env
DB_TYPE=sqlite
SQLITE_PATH=data/blog.sqlite
MEDIA_STORAGE=local
UPLOAD_DIR=uploads
```

如果本地使用 PostgreSQL：

```env
DB_TYPE=postgres
DATABASE_URL=postgres://blog:blog@localhost:15432/blog?sslmode=disable
REDIS_ADDR=localhost:16379
# 如有密码：REDIS_PASSWORD=your-password
# 本机共享 Redis 示例：REDIS_ADDR=10.1.152.5:6379 / REDIS_PASSWORD=123456
```

如果本地使用 MinIO：

```env
MEDIA_STORAGE=minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=blog-minio
MINIO_SECRET_KEY=change-this-minio-password
MINIO_BUCKET=blog-media
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000/blog-media
```

API 启动时会自动加载 `api/.env`；系统环境变量优先级高于文件配置。

### 3. 启动 API

```bash
cd api
go install github.com/air-verse/air@v1.61.7
air
```

没有安装 Air 时：

```bash
cd api
go run ./cmd/server
```

健康检查：`http://localhost:8080/api/health`

### 4. 启动 Web

```bash
cd web
npm install
npm run dev
```

默认前端地址：`http://localhost:5173`

Web 默认通过 `VITE_API_BASE_URL=/api` 访问 API；需要改后端地址时复制 `web/.env.example` 为 `web/.env` 并调整。

## 默认账号

- Admin: `admin@example.com` / `password`
- User: `linyi@example.com` / `password`

首次迁移/初始化会写入示例文章、用户、评论、投稿、专题、导航和系统设置数据。

## 环境变量

常用 API 环境变量：

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `APP_ENV` | 运行环境，生产可设为 `production` | `development` |
| `API_HTTP_ADDR` | API 监听地址 | `:8080` |
| `WEB_ORIGIN` | 前端源，用于 CORS/CSRF | `http://localhost:5173` |
| `PUBLIC_URL` | 站点公网地址，用于邮件和 SEO | 同 `WEB_ORIGIN` |
| `DB_TYPE` | `sqlite` 或 `postgres` | `sqlite` |
| `DATABASE_URL` | PostgreSQL 连接串 | `postgres://blog:blog@localhost:5432/blog?sslmode=disable` |
| `SQLITE_PATH` | SQLite 文件路径 | `data/blog.sqlite` |
| `REDIS_ADDR` | Redis 地址（限流与登录锁定） | `localhost:6379` |
| `REDIS_PASSWORD` | Redis 密码，可为空 | 空 |
| `MEDIA_STORAGE` | `local` 或 `minio` | `local` |
| `UPLOAD_DIR` | 本地上传目录 | `uploads` |
| `MINIO_ENDPOINT` | MinIO 内部访问地址 | 空 |
| `MINIO_ACCESS_KEY` | MinIO Access Key | 空 |
| `MINIO_SECRET_KEY` | MinIO Secret Key | 空 |
| `MINIO_BUCKET` | 媒体 Bucket | `blog-media` |
| `MINIO_USE_SSL` | MinIO 内部连接是否使用 HTTPS | `false` |
| `MINIO_PUBLIC_URL` | 图片公网访问前缀 | 空 |
| `SMTP_HOST` | SMTP 主机 | 空 |
| `SMTP_PORT` | SMTP 端口 | `587` |
| `SMTP_USERNAME` | SMTP 用户名 | 空 |
| `SMTP_PASSWORD` | SMTP 密码 | 空 |
| `SMTP_FROM` | 发件人 | 空 |

## 生产部署

Docker Compose 文件位于 `deploy/docker-compose.yml`，包含：

- `postgres`: PostgreSQL 数据库
- `redis`: 限流与登录失败锁定（不可用时回退内存）
- `minio`: 媒体对象存储
- `minio-init`: 创建 bucket 并设置匿名下载
- `api`: Go API
- `web`: 前端静态站点

建议在 `deploy/.env` 中配置生产变量：

```env
APP_ENV=production
WEB_ORIGIN=https://example.com
PUBLIC_URL=https://example.com

DB_TYPE=postgres
DATABASE_URL=postgres://blog:blog@postgres:5432/blog?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

MEDIA_STORAGE=minio
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-strong-secret
MINIO_BUCKET=blog-media
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=https://example.com/blog-media

SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
```

启动：

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

生产注意事项：

- 修改默认 PostgreSQL 密码、MinIO 密钥和公开域名。
- `MINIO_PUBLIC_URL` 会写入媒体 URL，生产环境应配置为公网可访问的域名或 CDN 地址。
- 如果前端和 API 不同域，确认 `WEB_ORIGIN`、`PUBLIC_URL`、反向代理和 Cookie 策略一致。
- 本地上传模式会使用 `api_uploads` 卷；生产推荐 MinIO。

## 备份与恢复

PostgreSQL 备份：

```bash
deploy/scripts/backup-postgres.sh
```

恢复：

```bash
deploy/scripts/restore-postgres.sh backups/blog-YYYYmmdd-HHMMSS.sql
```

## Turnstile 本地开发

- 如果使用真实 Cloudflare Turnstile site key，本地调试需要在 Cloudflare dashboard 中把 `localhost` 和 `127.0.0.1` 加入允许域名。
- 如果只测试流程，可使用 Cloudflare 测试 keys：
  - Site Key: `1x00000000000000000000AA`
  - Secret Key: `1x0000000000000000000000000000000AA`

## 验证

后端：

```bash
cd api
go test ./...
```

前端类型检查：

```bash
cd web
npm run typecheck
```

前端生产构建：

```bash
cd web
npm run build
```

## 友链

- [Linux.do](https://linux.do/)
- [NodeLoc](https://nodeloc.com/)
