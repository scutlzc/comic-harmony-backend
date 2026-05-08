# ComicHarmony Backend

Go 后端服务，为 ComicHarmony 鸿蒙漫画阅读器提供 RESTful API。

## 技术栈

- **Go 1.22** + [chi](https://github.com/go-chi/chi) 路由
- **PostgreSQL 16** + [pgx/v5](https://github.com/jackc/pgx)
- **JWT** 认证（Bearer Token）
- **WebP** 图片处理（ffmpeg + libwebp）

## API 端点（22 个）

| 模块 | 端点 |
|------|------|
| 健康检查 | `GET /health` |
| 用户 | `POST /auth/register`, `POST /auth/login` |
| 漫画 | `GET /api/v1/comics`, `GET /api/v1/comics/{id}`, `POST /api/v1/comics`, `DELETE /api/v1/comics/{id}` |
| 章节 | `GET /api/v1/comics/{id}/chapters`, `GET /api/v1/chapters/{id}` |
| 上传 | `POST /api/v1/upload` (CBZ/CBR/7Z/EPUB/PDF/MOBI → WebP) |
| 数据源 | `GET/POST/DELETE /api/v1/sources`, 浏览/搜索 |
| 收藏 | `GET/POST/DELETE /api/v1/favorites`, `GET .../check/{id}` |
| 历史 | `GET/POST/DELETE /api/v1/history` |

## 快速开始

```bash
# 1. 编译
make build

# 2. 配置数据库
export DATABASE_URL="postgres://.../comic_harmony"
make migrate

# 3. 运行
make run

# 4. Docker Compose（推荐）
make docker-up
```

## 上传支持格式

| 格式 | 处理方式 |
|------|---------|
| CBZ/ZIP | 直接解压 |
| CBR/RAR | unrar 解压 |
| CB7/7Z | 7z 解压 |
| EPUB | 解压提取图片 |
| PDF | pdfimages 提取 |
| MOBI | Calibre 转 CBZ |

## 项目结构

```
cmd/server/main.go          # 入口
internal/
├── config/                  # 配置
├── database/               # PostgreSQL 连接
├── datasource/             # 数据源层 (Komga/WebDAV/CloudDrive)
│   ├── core/               # 接口 + 管理器
│   ├── komga/
│   ├── webdav/
│   └── clouddrive/
├── handler/                # API 处理器
├── middleware/             # CORS / 日志
├── model/                 # 数据模型
├── repository/            # 数据访问层
├── response/              # 统一响应
├── service/               # 业务逻辑
└── upload/                # 上传解析管道
migrations/                # SQL 迁移
```
