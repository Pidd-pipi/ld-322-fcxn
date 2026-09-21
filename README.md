# 绿境智控｜温室环境监测面板

> **项目类型：全栈Web应用** · 面向现代农业温室的 IoT 环境监测、报警、设备控制与分析报告系统。

## Docker Compose 一键启动（推荐）

> 首次启动前请先复制环境变量文件；项目即使位于中文目录名下也能正常运行。

```bash
cd "农业与生活服务主题项目提示词/ld-322"
cp .env.example .env
docker compose up -d
```

检查全部服务状态：

```bash
docker compose ps
curl http://localhost:19622/healthz
```

访问：

- 前端面板：http://localhost:18622
- 后端健康检查：http://localhost:19622/healthz
- API 基址：http://localhost:19622/api/v1
- 演示账户：`admin` / `admin123`（前端会自动获取演示令牌）

停止服务并保留数据：

```bash
docker compose down
```

## 主要功能

- **多温室总览**：预置两个温室，卡片展示温度、湿度、光照、CO₂、土壤湿度的最新数值；30 秒自动刷新。
- **传感器采集与模拟**：通过 API 写入传感器读数；总览页可一键生成一轮演示采样。
- **趋势与历史**：按温室和日/周/月范围查看 ECharts 折线趋势，支持图表缩放、平移及 CSV 导出。
- **阈值报警**：每个传感器具备上下限；超限时持久化报警并通过 WebSocket 推送，支持标记为已处理。
- **远程控制**：可开关循环风机、遮阳帘、灌溉泵、补光灯；每次操作保留设备操作记录，支持创建定时任务 API。
- **环境报告**：自动计算平均值、最高/最低值与报警统计，支持日/周/月报告以及 PDF 导出。

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 前端 | React 18、TypeScript、Vite、Ant Design、ECharts |
| 后端 | **Go 1.22 + Gin + GORM** |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 实时通信 | gorilla/websocket |
| 认证 | JWT（golang-jwt/jwt/v5） |
| 部署 | Docker Compose、Nginx |

## 本地开发（备选）

需要先自行启动 MySQL 8 和 Redis，并按 `.env` 配置连接信息。

```bash
# 终端 1：后端
cd backend
go mod tidy
go run ./cmd/server

# 终端 2：前端
cd frontend
npm install
npm run dev
```

本地 Vite 服务会把 `/api` 和 `/ws` 代理到 `http://localhost:19622`。

## API 清单

所有业务响应遵循：`{ "code": 0, "message": "ok", "data": ... }`。写操作需携带 `Authorization: Bearer <token>`；先用登录接口换取令牌。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/auth/login` | 演示登录 |
| GET / POST | `/api/v1/greenhouses` | 温室列表 / 新增温室 |
| GET | `/api/v1/greenhouses/:id` | 温室详情（传感器、阈值、设备） |
| POST | `/api/v1/readings` | 写入传感器读数并评估报警 |
| GET | `/api/v1/readings/latest?greenhouse_id=1` | 最新读数 |
| GET | `/api/v1/readings/history?greenhouse_id=1` | 历史读数；可带 `start`、`end`、`types` |
| POST | `/api/v1/greenhouses/:id/simulate` | 生成模拟读数 |
| PUT | `/api/v1/sensors/:id/threshold` | 更新传感器上下限 |
| GET / PATCH | `/api/v1/alerts`、`/api/v1/alerts/:id/handle` | 报警查询 / 处理 |
| GET / PATCH | `/api/v1/devices`、`/api/v1/devices/:id/toggle` | 设备查询 / 开关 |
| POST | `/api/v1/schedules` | 创建设备定时任务 |
| GET | `/api/v1/reports/environment?greenhouse_id=1&range=day` | 环境分析报告 |
| GET | `/ws` | WebSocket 读数/报警/设备状态推送 |

完整的接口轮廓位于 [`backend/api/openapi.yaml`](backend/api/openapi.yaml)。

## 项目结构

```text
ld-322/
├── docker-compose.yml            # 前端、后端、MySQL、Redis 编排
├── .env.example                  # Docker 环境变量示例
├── frontend/
│   ├── src/api/                  # API 客户端
│   ├── src/components/charts/    # 独立 ECharts 组件
│   ├── src/components/dashboard/ # 监测、设备、报警组件
│   ├── src/pages/                # 总览、历史、报告页面
│   ├── Dockerfile
│   └── nginx.conf                # SPA、API、WebSocket 反向代理
└── backend/
    ├── cmd/server/main.go        # 仅负责装配与启动
    ├── internal/
    │   ├── config/ constants/ errors/ logger/
    │   ├── model/ repository/ service/ handler/ middleware/
    │   ├── dto/ router/ websocket/
    ├── database/migrations/      # 迁移边界文档
    ├── database/seeds/           # 演示数据说明
    ├── api/openapi.yaml
    └── Dockerfile
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `COMPOSE_PROJECT_NAME` | `cygreenenv` | Compose 项目/容器前缀，避免中文目录名问题 |
| `DB_NAME` | `greenhouse` | MySQL 数据库名 |
| `DB_USER` / `DB_PASSWORD` | `greenhouse` / `greenhouse_pwd` | 应用数据库账号 |
| `DB_ROOT_PASSWORD` | `root_pwd` | MySQL root 密码 |
| `JWT_SECRET` | 示例随机值 | JWT 签名密钥，生产环境必须替换 |
| `FRONTEND_PORT` | `18622` | 前端宿主机端口 |
| `BACKEND_PORT` | `19622` | 后端宿主机端口 |
| `DB_PORT` | `33062` | MySQL 调试端口，可按需修改 |

## Docker 部署说明与常见问题

- Compose 顶层 `name: cygreenenv` 与 `.env` 中的 `COMPOSE_PROJECT_NAME` 保证在中文目录下仍有合法项目名；启动命令无需 `-p`。
- `db_data` 与 `redis_data` 是命名数据卷，不会绑定到含中文的宿主机路径；`docker compose down` 不会删除数据。若要清理演示数据，请执行 `docker compose down -v`。
- Nginx 将 `/api/` 转发给 `backend:8080`，`/ws` 转发为 WebSocket；浏览器请求不硬编码 `localhost`。
- 若端口冲突，编辑 `.env` 中的 `FRONTEND_PORT`、`BACKEND_PORT` 或 `DB_PORT` 后重新执行 `docker compose up -d`。
- 若后端尚未健康，前端会等待后端健康检查；可用 `docker compose logs backend` 排查 MySQL 连接错误。

## License

MIT License。
