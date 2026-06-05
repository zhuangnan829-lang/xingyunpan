# 星云盘 V2

星云盘 V2 是一个基于 Go + Gin + GORM + Vue 3 + Vite 的私有网盘系统，提供文件管理、分享访问、回收站、版本管理、协作、后台管理、队列任务、监控指标和 Swagger API 文档等能力。项目采用前后端分离架构，后端负责 API、认证、存储与任务调度，前端提供用户网盘与管理员控制台。

## 功能特性

- 用户认证：注册、登录、JWT 访问令牌、刷新令牌、邮箱验证码、找回密码。
- 文件管理：上传、下载、重命名、删除、目录管理、秒传检测、缩略图、自定义属性。
- 分片上传：初始化上传、分片记录、进度查询、完成合并、取消任务。
- 文件分享：创建分享、访问公开分享、访问码验证、下载次数与访问次数统计。
- 回收站：移入回收站、恢复、永久删除、清空回收站。
- 文件版本：历史版本查看、版本下载、版本恢复、版本删除。
- 协作权限：添加协作者、权限检查、协作者列表、权限更新与移除。
- 搜索能力：文件搜索、搜索建议、全文搜索配置与重建索引入口。
- 管理后台：站点设置、外观设置、用户与用户组、存储策略、节点、文件、物理文件、邮件、验证码、队列任务等。
- 运维能力：健康检查、Prometheus 指标、Swagger 文档、pprof 开发调试、后台 worker。

## 技术栈

| 模块 | 技术 |
| --- | --- |
| 后端 | Go 1.25、Gin、GORM、Viper、Zap |
| 数据库 | MySQL 8.0 |
| 缓存/队列辅助 | Redis 6+ |
| 存储 | 本地存储，配置中预留 MinIO / OSS |
| 认证 | JWT、bcrypt |
| API 文档 | swaggo / Swagger |
| 监控 | Prometheus metrics |
| 前端 | Vue 3、Vite 5、TypeScript、Pinia、Vue Router |
| UI | Element Plus、Arco Design Vue |
| 测试 | Go test、Vitest |

## 目录结构

```text
.
├── cmd/
│   ├── server/                 # 后端 HTTP 服务入口
│   └── worker/                 # 统一队列 worker 入口
├── configs/                    # 配置模板、Nginx、systemd、监控配置
├── docs/                       # Swagger / Postman 等文档资源
├── frontend/                   # Vue 3 + Vite 前端
│   ├── src/api/                # 前端 API 封装
│   ├── src/router/             # 路由配置
│   ├── src/stores/             # Pinia 状态
│   └── src/views/              # 页面视图
├── internal/
│   ├── config/                 # 配置、数据库、Redis 初始化
│   ├── controller/             # HTTP 控制器
│   ├── middleware/             # 鉴权、日志、限流、指标、文件校验
│   ├── model/                  # GORM 模型
│   ├── queue/                  # 后台队列运行时
│   ├── repository/             # 数据访问层
│   ├── service/                # 业务服务层
│   └── worker/                 # 清理、采集等 worker 逻辑
├── pkg/                        # 公共包：缓存、加密、哈希、JWT、日志、Redis、存储
├── scripts/                    # 启动、部署、迁移、诊断、压测、备份脚本
├── tools/search/               # Meilisearch / Tika 等搜索相关工具
├── docker-compose.yml          # MySQL、Redis、MinIO 本地依赖
└── Makefile                    # 常用后端命令
```

## 快速开始

### 环境要求

- Go 1.25 或与 `go.mod` 兼容的 Go 版本
- Node.js 18+ 和 npm
- MySQL 8.0+
- Redis 6+
- Docker / Docker Compose，可选，用于快速启动 MySQL、Redis、MinIO

### 1. 启动依赖服务

```bash
docker compose up -d
```

如果你的环境使用旧版 Compose 命令，也可以执行：

```bash
docker-compose up -d
```

默认会启动：

- MySQL: `localhost:3306`
- Redis: `localhost:6379`
- MinIO API: `localhost:9000`
- MinIO Console: `localhost:9001`

### 2. 准备配置

```bash
cp configs/config.yaml.example configs/config.yaml
```

根据你的本地环境修改 `configs/config.yaml`。如果使用仓库自带的 `docker-compose.yml`，数据库连接通常可以配置为：

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: password
  database: xingyunpan

redis:
  host: localhost
  port: 6379
```

生产环境务必修改 `jwt.secret`、数据库密码、邮箱 SMTP 配置和公开访问地址。

### 3. 启动后端

后端启动时会加载配置、初始化数据库和 Redis，并自动迁移主要数据表。

```bash
go mod download
go run cmd/server/main.go
```

后端默认地址：

- API: `http://localhost:8080`
- 健康检查: `http://localhost:8080/health`
- Swagger: `http://localhost:8080/swagger/index.html`
- Prometheus 指标: `http://localhost:8080/metrics`
- pprof: `http://localhost:6060/debug/pprof/`，仅非 release 模式启动

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认地址为 `http://localhost:3000`，开发环境会把 `/api` 请求代理到 `http://localhost:8080`。如需修改代理目标，可设置 `VITE_PROXY_TARGET`。

### 5. 启动后台 worker，可选

如果需要处理队列任务、清理任务或异步文件任务，可以单独启动 worker：

```bash
go run cmd/worker/main.go
```

当前 worker 使用统一队列运行器，依赖数据库、Redis 和本地存储配置。

## Windows 本地启动

仓库提供了若干 Windows 脚本，适合本地开发调试：

```bat
scripts\start-all-services-windows.bat
```

该脚本会检查 Go、Node.js、MySQL、Redis，并尝试启动后端和前端。脚本内容目前也存在部分历史编码文本，但核心命令仍可作为参考。

## 常用命令

```bash
# 启动后端
go run cmd/server/main.go

# 启动 worker
go run cmd/worker/main.go

# 构建后端
go build -o bin/server ./cmd/server
go build -o bin/worker ./cmd/worker

# 后端测试
go test ./...

# 前端开发
cd frontend && npm run dev

# 前端构建
cd frontend && npm run build

# 前端测试
cd frontend && npm run test
```

Makefile 也提供了部分后端快捷命令：

```bash
make dev
make build
make test
make docker-up
make docker-down
```

## API 概览

完整接口以代码路由和 Swagger 为准：

- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI YAML: `docs/api-swagger.yaml`
- OpenAPI JSON: `docs/api-swagger.json`

主要接口分组：

| 分组 | 路径前缀 | 说明 |
| --- | --- | --- |
| 健康与监控 | `/health`、`/ping`、`/metrics` | 服务健康、连通性、Prometheus 指标 |
| 用户认证 | `/api/v1/user/*` | 注册、登录、邮箱验证码、密码重置、用户信息 |
| 文件 | `/api/v1/file/*` | 上传、列表、下载、重命名、删除、秒传、缩略图 |
| 分片上传 | `/api/v1/files/multipart/*` | 大文件分片上传流程 |
| 文件夹 | `/api/v1/folder/*` | 创建、重命名、移动、删除 |
| 分享 | `/api/v1/shares/*` | 创建分享、访问分享、访问码验证、删除分享 |
| 搜索 | `/api/v1/search/*` | 文件搜索与建议 |
| 回收站 | `/api/v1/recycle/*` | 回收站列表、恢复、永久删除、清空 |
| 版本与协作 | `/api/v1/files/*`、`/api/v1/collaborations/*` | 文件版本、协作者、权限 |
| 管理后台 | `/api/v1/admin/*` | 站点、用户、文件、存储、节点、邮件、队列等管理接口 |

## 配置说明

核心配置位于 `configs/config.yaml`，可参考 `configs/config.yaml.example`。

```yaml
server:
  port: 8080
  mode: debug
  base_url: http://localhost:8080

database:
  host: localhost
  port: 3306
  username: root
  password: password
  database: xingyunpan

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

storage:
  type: local
  base_path: ./storage

jwt:
  secret: change-this-secret-in-production
  expire_hours: 24
  refresh_expire_hours: 168

upload:
  max_file_size: 5368709120
  chunk_size: 5242880
```

说明：

- `server.base_url` 会影响分享链接、邮件链接等外部访问地址。
- `storage.base_path` 是本地文件存储目录。
- `upload.max_file_size` 默认约 5GB。
- `email.enabled` 为 `false` 时，邮箱验证码相关能力需要按实际业务确认使用方式。
- `storage.type` 当前 worker 明确支持 `local`，MinIO / OSS 配置项已预留。

## 前端页面

当前前端路由包括：

- 登录、注册、找回密码
- 我的文件、分享访问、我的分享、回收站、协作、个人资料
- 管理后台首页
- 管理后台：站点设置、文件系统、存储策略、节点、用户组、用户、文件、物理文件
- 部分分类视图和管理视图使用占位页面，后续可继续接入实际业务数据

## 开发约定

后端整体按照 Controller、Service、Repository 分层组织：

```text
HTTP 请求
  -> middleware
  -> controller
  -> service
  -> repository
  -> database / redis / storage
```

建议保持以下约定：

- 控制器只处理参数、鉴权上下文和响应格式。
- 业务规则放在 `internal/service`。
- 数据库访问放在 `internal/repository`。
- 公共能力放在 `pkg`。
- 新增接口后同步更新 Swagger 或相关 API 文档。
- 修改前端 API 时同步检查 `frontend/src/types` 和对应页面状态。

## 测试与验证

后端：

```bash
go test ./...
```

前端：

```bash
cd frontend
npm run test
npm run build
```

项目中还有多组验证脚本，例如：

- `scripts/verify.bat`
- `scripts/verify-production-setup.bat`
- `scripts/verify-monitoring.bat`
- `scripts/validate-api-docs.bat`
- `scripts/benchmark.bat`

Linux 环境可优先查看同名 `.sh` 脚本。

## 部署相关

仓库提供了以下部署与运维资源：

- `docker-compose.yml`：本地依赖服务。
- `docker-compose.monitoring.yml`：监控相关服务。
- `configs/nginx-ssl.conf`：Nginx SSL 参考配置。
- `configs/systemd/`：server、worker、备份、清理、监控等 systemd 示例。
- `configs/prometheus.yml`、`configs/alerts.yml`、`configs/alertmanager.yml`：Prometheus 与告警配置。
- `scripts/deploy.sh`、`scripts/deploy.bat`：部署脚本入口。
- `scripts/backup.*`、`scripts/restore.*`：备份与恢复脚本。

生产部署前建议检查：

- `configs/config.prod.yaml` 是否符合目标环境。
- JWT 密钥和数据库密码是否已替换。
- `server.mode` 是否设置为 `release`。
- 文件存储目录、日志目录、备份目录是否有权限。
- 前端构建产物和 Nginx 反向代理是否指向正确后端地址。
- `/health`、`/metrics`、Swagger 或 API 探活是否正常。

## 常见问题

### 后端启动失败，提示数据库连接错误

检查 MySQL 是否已启动，以及 `configs/config.yaml` 中的 `database.host`、`database.port`、用户名、密码和数据库名是否正确。使用 Docker Compose 时，默认 root 密码是 `password`。

### Redis 连接失败

后端部分功能会降级或异常。请确认 Redis 已启动，端口为 `6379`，配置中的密码和 db 与实际环境一致。

### 前端请求接口失败

确认后端已运行在 `http://localhost:8080`。开发环境默认由 Vite 将 `/api` 代理到后端，如后端地址不同，请设置 `VITE_PROXY_TARGET`。

### Swagger 页面打不开

确认后端已启动，并访问 `http://localhost:8080/swagger/index.html`。如果文档不是最新，可运行相关 Swagger 生成脚本后再启动服务。

### 端口被占用

修改 `configs/config.yaml` 中的 `server.port`，或修改 `frontend/vite.config.ts` 中的开发服务器端口。

## 当前注意事项

- 仓库中部分历史脚本和源码字符串存在编码损坏现象，README 已整理为 UTF-8 中文。
- `configs/config.yaml` 可能包含本地真实配置，提交前请确认不要泄露敏感信息。
- `frontend/node_modules`、`frontend/dist`、`server.exe`、`worker.exe` 等构建或依赖产物不建议纳入版本库。

## License

MIT License
