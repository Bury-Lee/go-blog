# GoBlog（StarDreamerCyberNook）

一个基于 `Gin + GORM + Redis + Elasticsearch` 的博客/社区后端项目，包含用户、文章、评论、消息、关注、聊天、站点配置等模块。

## 🌐 多语言文档

**其他语言：**:[🇨🇳 中文](README.zh-CN.md) • [🇺🇸 English](../README.md) • [🇯🇵 日本語](README.ja.md) • [🇰🇷 한국어](README.ko.md)

## 👤 适合人群:

- 追求独立与自主的技术爱好者：希望拥有完全可控的个人博客或社区平台，不依赖第三方内容平台的规则限制，追求个性化定制和数据主权
- 拥有小型服务器的站长：具备基础服务器资源（如VPS、云主机或家用服务器），希望搭建稳定可靠的图文内容服务
- 社区运营者：需要搭建垂直领域的交流社区、官方论坛或轻量化内容平台，支持用户互动、内容审核和SEO优化
- 注重隐私与数据安全的用户：不愿将核心内容数据托管于商业平台，希望实现本地化部署和自主备份

## 📋 功能速览

- **博客与社区一体化**：文章、评论、收藏、消息、关注、私聊集中在同一套后端
- **搜索能力完整**：基于 `Elasticsearch` 提供文章全文搜索、高亮和多维排序
- **高并发友好**：浏览、点赞、收藏、评论计数先写 Redis，再由定时任务批量同步数据库
- **AI 已接入业务**：支持站点 AI 助手，以及文章、评论、昵称等内容审核
- **运营能力齐全**：支持站点配置、SEO、轮播图、友情链接、推广位和日志管理

> 📖 **功能文档**：[博客功能文档](功能文档.md)

> ⚠️ **注意**：项目本体支持数据库读写分离，但**不原生支持数据库之间的数据同步**。仓库中的 `otherSoftware` 提供了基于 Canal 的简易订阅同步方案（本文已给出配置与使用步骤）。

## 🏗️ 项目结构

```
go-blog
├─ api/                 # 控制器层
├─ router/              # 路由注册
├─ models/              # 数据模型与 ES Mapping
├─ service/             # 业务服务（含定时任务、ES服务、Redis服务）
├─ middleware/          # 中间件
├─ core/                # 配置/日志/DB/Redis/ES/AI 初始化
├─ conf/                # 配置结构体
├─ flags/               # 命令行参数（迁移、建索引、建用户）
├─ init/                # 本地依赖服务 docker-compose 与基础配置
├─ otherSoftware/       # 数据同步方案（Canal + canal-go）
├─ setting.yaml         # 主配置文件
└─ main.go              # 入口
```

## 🔧 环境要求

| 组件            | 版本要求   | 备注                                            |
| ------------- | ------ | --------------------------------------------- |
| Go            | 1.26.0 | 按项目 `go.mod` 版本                               |
| MySQL         | 5.7    | 项目内提供 docker-compose                          |
| Redis         | 7      | 项目内提供 docker-compose，双实例                      |
| Elasticsearch | 7.17.x | 项目使用 `olivere/elastic/v7`                     |
| AI 服务         | 可选     | 默认配置示例：`http://localhost:1234/v1` (目前仅支持本地模型) |

## 🚀 快速启动

### 1️⃣ 启动依赖服务

在项目根目录 `go-blog` 下分别执行：

```bash
# 启动 MySQL
docker compose -f init/SQL/docker-compose.yml up -d

# 启动 Redis
docker compose -f init/Redis/docker-compose.yml up -d

# 启动 Elasticsearch
docker compose -f init/ES/docker-compose.yml up -d
```

在写完配置文件后执行发行版程序:

Windows

```bash
# 启动项目
.\main_windows_amd64.exe
```

Linux

```bash
# 启动项目
./main_linux_amd64
```

macOS

```bash
# 启动项目
./main_macos_amd64
```

> 💡 **提示**：由于 MySQL 5.7 以上版本和 Canal 之间的兼容性不是特别好，故在 `init/SQL/docker-compose.yml` 中配置了 `5.7.360.0` 版本。

### 2️⃣ 编译项目

在Windows上编译:
```bash
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

### 3️⃣ 修改主配置 `setting.yaml`

建议优先确认以下关键字段：

- `system.ip`、`system.port`、`system.env`、`system.gin_mode`
- `db`（第一个为写库，后续为读库）
- `redisStatic`、`redisDynamic`
- `es.url`、`es.username`、`es.password`
- `jwt.accessTokenSecret`、`jwt.refreshTokenSecret`
- `email`（用于邮箱验证码登录/注册）
- `ai.enable`（不使用 AI 可设为 `false`）

<details>
<summary>📖 点击查看完整配置示例</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  gin_mode: debug # Gin 运行模式：debug 或 release

log:
  app: GoBlog
  dir: log
  log_level: debug

ai:
  enable: true
  model: local
  host: http://localhost:1234/v1
  ApiKey: 123456
  nickName: 昵称
  avatar: "https://example.com/avatar.jpg" # TODO：后续适配头像来源

email:
  domain: smtp.qq.com
  port: 587 # QQ 邮箱常用端口为 587 或 465
  sendEmail: xxxxxx@qq.com
  authCode: "xxxxxxx" # 邮箱授权码
  sendNickname: 昵称

upload:
  size: 20 # 上传文件大小限制，单位 MB
  whiteList: # 上传文件白名单
    ".jpg": ~
    ".jpeg": ~
    ".png": ~
    ".gif": ~
    ".webp": ~
    ".bmp": ~
    ".tiff": ~
    ".svg": ~
  uploadDir: images # 图片上传目录

jwt:
  accessExpire: 30 # 访问令牌过期时间（分钟），推荐 30 分钟
  refreshExpire: 172 # 刷新令牌过期时间（小时），推荐一周（172 小时）
  accessTokenSecret: xxxxx
  refreshTokenSecret: xxxxxxx
  issuer: "StarDreamer"

redisStatic:
  addr: 127.0.0.1:6379
  password: redis
  db: 1

redisDynamic:
  addr: 127.0.0.1:6378
  password: redis
  db: 2

es:
  url: http://127.0.0.1:9200
  username: elastic
  password: es

db:
  # 多数据库读写分离：第一个为写库，其余为读库
  - user: root
    password: root
    host: 127.0.0.1
    port: 3306
    db_name: db
    debug: false # 是否启用调试（打印完整日志）
    sql_name: mysql
  # - user: root # 可按此格式继续添加多个数据库
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db: db
  #   debug: false # 是否启用调试（打印完整日志）
  #   source: mysql

site:
  siteInfo:
    title: "星梦网络空间" # 站点标题
    logo: "/static/images/logo.png" # 站点 Logo 路径
    beian: "京ICP备XXXXXXXX号" # 备案号
    mode: 1 # 运行模式（1: 博客模式, 2: 社区模式等，需对应代码枚举）

  project:
    title: "StarDreamer" # 项目名称
    icon: "/static/images/favicon.ico" # 项目图标
    webPath: "https://www.example.com" # 项目访问路径

  seo:
    keywords: "技术博客, Go语言, 人工智能, 分享"
    description: "一个专注于技术分享与人工智能探索的个人站点。"

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "热门文章"
        enable: true
      - title: "最新评论"
        enable: true
      - title: "友情链接"
        enable: true
      - title: "标签云"
        enable: true

  article:
    # 说明：Go 字段名是 DisableExamination，但 yaml tag 是 enableExamination
    # 若以 enableExamination 读取：true 表示"启用审核"（即不禁用）
    # 业务语义：true = 需要审核，false = 无需审核
    enableExamination: true

  login:
    QQLogin: true # TODO：尚未实现
    usernamePassword: true
    emailLogin: true # TODO：可考虑固定开启（基础登录方式）
    captcha: false # 是否启用验证码
```

</details>

### 4️⃣ 初始化数据库结构

```bash
go run main.go -db
```

### 5️⃣ 初始化 ES 索引

```bash
go run main.go -es
```

### 6️⃣ 启动服务

```bash
go run main.go -f setting.yaml
```

🌐 启动后默认监听：`http://127.0.0.1:8080`

## 📝 命令行参数

| 参数                  | 说明                        |
| ------------------- | ------------------------- |
| `-f`                | 配置文件路径（默认 `setting.yaml`） |
| `-db`               | 执行 GORM 自动迁移              |
| `-es`               | 创建/重建 ES 索引               |
| `-v`                | 查看版本                      |
| `-t user -s create` | 命令行创建用户                   |

**示例：**

```bash
go run main.go -t user -s create
```

## ⚙️ 关键运行说明

- **初始化顺序**：配置 → 日志 → IP库 → DB → Redis → ES → AI → 定时任务 → 路由
- **ES 为强依赖**：ES 初始化失败会导致服务启动中断
- **接口规范**：当前接口统一使用 `/api` 前缀，例如 `/api/user/login`
- **静态资源**：静态资源目录映射为 `/web`，对应本地 `static/`
- **定时任务**：当前定时任务 `SyncArticle/SyncComment` 为 10 分钟级批量同步，用于刷新 Redis 中的计数增量

## 🔄 数据库同步方案（otherSoftware）

### 📍 背景

项目原生只做读写分离，不做跨库自动同步。\
`otherSoftware` 提供了一个简易方案：

- `otherSoftware/canal`：Canal Server，订阅 MySQL binlog
- `otherSoftware/canal-go/canal-go-1.1.2/samples/main.go`：Go 客户端消费变更并做下游处理（示例里同步到 ES）

### ✅ 前置条件

1. MySQL 开启 binlog（项目 `init/SQL/master/my.cnf` 已包含 `log-bin`、`server-id`）
2. Canal 连接信息和你的 MySQL 实际账号密码一致
3. `canal.instance.master.address` 指向要订阅的 MySQL（通常主库 `127.0.0.1:3306`）
4. 建议统一密码，避免 `setting.yaml`、MySQL、Canal 三处不一致

### 🔧 配置 Canal

编辑文件：

- `otherSoftware/canal/canal.deployer-1.1.8/conf/canal.properties`
- `otherSoftware/canal/canal.deployer-1.1.8/conf/example/instance.properties`

**重点关注：**

- `canal.destinations=example`
- `canal.port=11111`
- `canal.instance.master.address=127.0.0.1:3306`
- `canal.instance.dbUsername=...`
- `canal.instance.dbPassword=...`
- `canal.instance.filter.regex=.*\..*`（可改成只监听目标库表）

### 🚀 启动 Canal

按仓库说明，使用 `startup.bat` 直接启动（Windows环境下）：

- **Windows**：双击 `otherSoftware/canal/canal.deployer-1.1.8/bin/startup.bat`

然后查看日志：

- `otherSoftware/canal/canal.deployer-1.1.8/logs/example/example.log`

看到 MySQL 连接成功相关日志即表示 Canal 端正常。

### 🎯 启动 canal-go 消费端

在 `otherSoftware/canal-go/canal-go-1.1.2` 目录执行：

```bash
go run samples/main.go
```

该示例默认：

- **连接 Canal**：`127.0.0.1:11111`
- **destination**：`example`
- **订阅规则**：`.*\.article_models`
- **处理逻辑**：读取 binlog 变更，转换后写入 `article_index`（ES）

如果你要同步到"另一个数据库"而不是 ES，可替换 `samples/ES_pkg` 中的处理逻辑为自己的 DB 写入逻辑。

### 📋 使用流程建议

1. 先启动 MySQL/Redis/ES 与 GoBlog 主服务
2. 启动 Canal Server（生产 binlog 订阅流）
3. 启动 canal-go 消费端（执行转换与落库/落 ES）
4. 在主库执行 `insert/update/delete` 验证目标端是否同步

## 🌐 NGINX 水平扩展（可选）

经过近期架构调整，项目服务层已具备无状态特性，业务数据统一落在 `MySQL / Redis / Elasticsearch`。  
在此基础上，可通过 `NGINX` 反向代理与负载均衡能力实现横向扩展。

### ✅ 适用场景

- 单实例 CPU 或连接数已接近瓶颈，需要平滑提升并发处理能力
- 已有多台主机或多个容器实例，希望对外统一一个访问入口
- 需要在不中断服务的情况下，逐步扩容或替换后端节点

### 🔧 实施要点

1. 启动多个 GoBlog 实例（建议统一版本与配置，仅端口不同）
2. 在 NGINX `upstream` 中注册多个后端节点
3. 通过 `proxy_pass` 将入口流量转发至 `upstream`
4. 按需启用健康检查、超时重试、连接保持等策略

### 📈 效果预期

- 多实例分担请求压力，提升整体吞吐量
- 降低单点负载，改善高峰期响应稳定性
- 支持按节点滚动发布，降低变更风险

## ❓ 常见问题

| 问题               | 解决方案                                                     |
| ---------------- | -------------------------------------------------------- |
| **MySQL 连接失败**   | 检查 `setting.yaml` 与 Canal 的账号密码、端口是否一致                   |
| **ES 启动失败**      | 确认 `http://127.0.0.1:9200` 可访问，账号密码匹配                    |
| **Redis 警告淘汰策略** | 项目会检查动态 Redis 的 `maxmemory-policy`                       |
| **接口 404**       | 注意当前接口带 `/api` 前缀，应访问 `/api/user/login` 这类路径             |
| **Canal 无数据**    | 优先检查 binlog 是否开启、`instance.properties` 的主库地址/账号、订阅规则是否匹配 |

## 💡 仅开发者提示

- 代码中含部分 TODO（如 ES 升级 v8、路由前缀切换、定时任务频率）
- 若用于生产，建议补充：鉴权、限流、监控、配置分环境管理、错误恢复与幂等处理

## 📞 联系我们

如果在部署或使用过程中遇到任何问题，欢迎通过以下方式反馈：

- **🐛 GitHub Issues**: 在 [项目仓库](https://github.com/Bury-Lee/go-blog) 提交 Issue
- **📧 邮箱联系**: <18151161@qq.com>
- **👥 交流群**: [星梦的交流群](https://qun.qq.com/universal-share/share?ac=1\&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g\&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D\&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ\&svctype=4\&tempid=h5_group_info)

作者非常乐意解决有价值的技术问题，也欢迎提交 PR 参与项目贡献！
