# GoBlog (StarDreamerCyberNook)

A blog/community backend project based on `Gin + GORM + Redis + Elasticsearch`, featuring modules for users, articles, comments, messages, follows, chat, and site configuration.

## 🌐 Multilingual Documentation

**Languages:** [🇨🇳 中文](docs/README.zh-CN.md) • [🇺🇸 English](README.md) • [🇯🇵 日本語](docs/README.ja.md) • [🇰🇷 한국어](docs/README.ko.md)

## 👤 Who Is It For?

- **Tech enthusiasts who want full control**: Build a fully self-hosted personal blog or community without relying on third-party platform rules; prioritize customization and data ownership
- **Site owners with a small server**: Have basic server resources such as a VPS, cloud instance, or home server and want a stable text-and-image content service
- **Community operators**: Need a vertical community, official forum, or lightweight content platform with user interaction, content moderation, and SEO optimization
- **Privacy and data security minded users**: Prefer local deployment and self-managed backups instead of hosting core content data on commercial platforms


## 📋 Features Overview

- **Blog & Community Integration**: Articles, comments, collections, messages, follows, and private chat all in one backend
- **Complete Search Capabilities**: Full-text article search, highlighting, and multi-dimensional sorting based on `Elasticsearch`
- **High Concurrency Friendly**: Views, likes, collections, and comment counts are written to Redis first, then synchronized to the database in batches by scheduled tasks
- **AI Integration**: Site AI assistant and content moderation for articles, comments, and nicknames
- **Multi-model AI Support**: Now supports multiple AIs (OpenAI interface compatible), added AI article summary and AI article rating features, and supports outputting AI responses in debug mode
- **Complete Operations Features**: Site configuration, SEO, banners, friend links, promotion slots, and log management

> 📖 **Feature Documentation**: [Blog Feature Documentation](docs/功能文档.md)

> ⚠️ **Note**: The project supports database read-write separation but **does not natively support data synchronization between databases**. The repository provides a simple subscription-based synchronization solution using Canal (configuration and usage steps are provided in this document).

## 🏗️ Project Structure

```
go-blog
├─ api/                 # Controller layer
├─ router/              # Route registration
├─ models/              # Data models and ES Mapping
├─ service/             # Business services (including scheduled tasks, ES services, Redis services)
├─ middleware/          # Middleware
├─ core/                # Configuration/log/DB/Redis/ES/AI initialization
├─ conf/                # Configuration structures
├─ flags/               # Command line arguments (migration, index creation, user creation)
├─ init/                # Local dependency services docker-compose and basic configuration
├─ otherSoftware/       # Data synchronization solution (Canal + canal-go)
├─ setting.yaml         # Main configuration file
└─ main.go              # Entry point
```

## 🔧 Environment Requirements

| Component | Version Requirement | Notes |
|----------|-------------------|-------|
| Go | 1.26.0 | According to project `go.mod` version |
| MySQL | 5.7 | Docker-compose provided in project |
| Redis | 7 | Docker-compose provided in project, dual instances |
| Elasticsearch | 7.17.x | Project uses `olivere/elastic/v7` |
| AI Service | Optional | Default configuration example: `http://localhost:1234/v1` (currently only supports local models) |

## 🚀 Quick Start

### 1️⃣ Start Dependency Services

Execute in the project root directory `go-blog`:

```bash
# Start MySQL
docker compose -f init/SQL/docker-compose.yml up -d

# Start Redis
docker compose -f init/Redis/docker-compose.yml up -d

# Start Elasticsearch
docker compose -f init/ES/docker-compose.yml up -d
```

After writing the configuration file, execute the release program:

Windows
```bash
# Start project
.\main_windows_amd64.exe
```
Linux
```bash
# Start project
./main_linux_amd64
```
macOS
```bash
# Start project
./main_macos_amd64
```



> 💡 **Tip**: Due to compatibility issues between MySQL versions above 5.7 and Canal, version `5.7.360.0` is configured in `init/SQL/docker-compose.yml`.

### 2️⃣ Compile Project

Compile on Windows:
```bash
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

### 3️⃣ Modify Main Configuration `setting.yaml`

Key fields to confirm first:

- `system.ip`, `system.port`, `system.env`, `system.gin_mode`
- `db` (first is write database, subsequent are read databases)
- `redisStatic`, `redisDynamic`
- `es.url`, `es.username`, `es.password`
- `jwt.accessTokenSecret`, `jwt.refreshTokenSecret`
- `email` (for email verification code login/registration)
- `ai.enable` (set to `false` if not using AI)

<details>
<summary>📖 Click to view complete configuration example</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  gin_mode: debug # Gin run mode: debug or release
  cron: true             # Whether to enable scheduled tasks
  scheduled_cleanup: true    # Whether to enable scheduled cleanup of visit records

log:
  app: GoBlog
  dir: log
  log_level: debug

ai:
  enable: true
  model: local
  host: http://localhost:1234/v1
  ApiKey: 123456
  nickName: Nickname
  avatar: "https://example.com/avatar.jpg" # TODO: Adapt avatar source later

email:
  domain: smtp.qq.com
  port: 587 # QQ email common ports are 587 or 465
  sendEmail: xxxxxx@qq.com
  authCode: "xxxxxxx" # Email authorization code
  sendNickname: Nickname

upload:
  size: 20 # Upload file size limit, unit MB
  whiteList: # Upload file whitelist
    ".jpg": ~
    ".jpeg": ~
    ".png": ~
    ".gif": ~
    ".webp": ~
    ".bmp": ~
    ".tiff": ~
    ".svg": ~
  uploadDir: images # Image upload directory

jwt:
  accessExpire: 30 # Access token expiration time (minutes), recommended 30 minutes
  refreshExpire: 172 # Refresh token expiration time (hours), recommended one week (172 hours)
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
  # Multiple database read-write separation: first is write database, rest are read databases
  - user: root
    password: root
    host: 127.0.0.1
    port: 3306
    db_name: db
    debug: false # Whether to enable debug (print complete logs)
    sql_name: mysql
  # - user: root # Can continue adding multiple databases in this format
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db: db
  #   debug: false # Whether to enable debug (print complete logs)
  #   source: mysql

site:
  siteInfo:
    title: "StarDreamer Cyberspace" # Site title
    logo: "/static/images/logo.png" # Site Logo path
    beian: "京ICP备XXXXXXXX号" # Filing number
    mode: 1 # Run mode (1: blog mode, 2: community mode, etc., needs to correspond to code enum)

  project:
    title: "StarDreamer" # Project name
    icon: "/static/images/favicon.ico" # Project icon
    webPath: "https://www.example.com" # Project access path

  seo:
    keywords: "tech blog, Go language, AI, sharing"
    description: "A personal site focused on technology sharing and AI exploration."

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "Popular Articles"
        enable: true
      - title: "Latest Comments"
        enable: true
      - title: "Friend Links"
        enable: true
      - title: "Tag Cloud"
        enable: true

  article:
    # Note: Go field name is DisableExamination, but yaml tag is enableExamination
    # If read as enableExamination: true means "enable review" (i.e., not disabled)
    # Business semantics: true = needs review, false = no review needed
    enableExamination: true

  login:
    QQLogin: true # TODO: Not yet implemented
    usernamePassword: true
    emailLogin: true # TODO: Can consider always on (basic login method)
    captcha: false # Whether to enable captcha
```

</details>

### 4️⃣ Initialize Database Structure

```bash
go run main.go -db
```

### 5️⃣ Initialize ES Index

```bash
go run main.go -es
```

### 6️⃣ Start Service

```bash
go run main.go -f setting.yaml
```

🌐 Default listening after startup: `http://127.0.0.1:8080`

## 📝 Command Line Parameters

| Parameter | Description |
|-----------|-------------|
| `-f` | Configuration file path (default `setting.yaml`) |
| `-db` | Execute GORM auto migration |
| `-es` | Create/rebuild ES index |
| `-v` | View version |
| `-t user -s create` | Create user via command line |

**Example:**
```bash
go run main.go -t user -s create
```

## ⚙️ Key Runtime Instructions

- **Initialization order**: Configuration → Log → IP library → DB → Redis → ES → AI → Scheduled tasks → Routes
- **ES is a strong dependency**: ES initialization failure will cause service startup interruption
- **Interface specification**: Current interfaces use `/api` prefix, e.g., `/api/user/login`
- **Static resources**: Static resource directory mapped as `/web`, corresponding to local `static/`
- **Scheduled tasks**: Current scheduled tasks `SyncArticle/SyncComment` are 10-minute level batch synchronization for refreshing count increments in Redis

## 🔄 Database Synchronization Solution (otherSoftware)

### 📍 Background

The project natively only does read-write separation, not cross-database automatic synchronization.  
`otherSoftware` provides a simple solution:

- `otherSoftware/canal`: Canal Server, subscribes to MySQL binlog
- `otherSoftware/canal-go/canal-go-1.1.2/samples/main.go`: Go client consumes changes and does downstream processing (example syncs to ES)

### ✅ Prerequisites

1. MySQL binlog enabled (project `init/SQL/master/my.cnf` already includes `log-bin`, `server-id`)
2. Canal connection info matches your actual MySQL account password
3. `canal.instance.master.address` points to the MySQL to subscribe (usually master `127.0.0.1:3306`)
4. Recommend unified password to avoid inconsistency between `setting.yaml`, MySQL, and Canal

### 🔧 Configure Canal

Edit files:

- `otherSoftware/canal/canal.deployer-1.1.8/conf/canal.properties`
- `otherSoftware/canal/canal.deployer-1.1.8/conf/example/instance.properties`

**Key focus:**

- `canal.destinations=example`
- `canal.port=11111`
- `canal.instance.master.address=127.0.0.1:3306`
- `canal.instance.dbUsername=...`
- `canal.instance.dbPassword=...`
- `canal.instance.filter.regex=.*\..*` (can change to only listen to target tables)

### 🚀 Start Canal

According to repository instructions, start directly with `startup.bat` (Windows environment):

- **Windows**: Double-click `otherSoftware/canal/canal.deployer-1.1.8/bin/startup.bat`

Then check logs:

- `otherSoftware/canal/canal.deployer-1.1.8/logs/example/example.log`

Seeing MySQL connection success logs indicates Canal is working properly.

### 🎯 Start canal-go Consumer

Execute in `otherSoftware/canal-go/canal-go-1.1.2` directory:

```bash
go run samples/main.go
```

This example defaults:

- **Connect to Canal**: `127.0.0.1:11111`
- **destination**: `example`
- **Subscription rule**: `.*\.article_models`
- **Processing logic**: Read binlog changes, convert and write to `article_index` (ES)

If you want to sync to "another database" instead of ES, replace the processing logic in `samples/ES_pkg` with your own DB write logic.

### 📋 Recommended Usage Process

1. Start MySQL/Redis/ES and GoBlog main service first
2. Start Canal Server (produce binlog subscription stream)
3. Start canal-go consumer (execute conversion and write to DB/ES)
4. Execute `insert/update/delete` in master database to verify if target end is synchronized

## 🌐 NGINX Horizontal Scaling (Optional)

After recent architecture adjustments, the service layer is now effectively stateless, and business data is centralized in `MySQL / Redis / Elasticsearch`.  
Based on this design, you can use `NGINX` reverse proxy and load balancing to scale out horizontally.

### ✅ Suitable Scenarios

- A single instance is approaching CPU or connection limits, and you need smoother concurrency scaling
- Multiple hosts or container instances are available, and you want one unified public entry point
- You need gradual capacity expansion or node replacement without service interruption

### 🔧 Implementation Notes

1. Start multiple GoBlog instances (same version and config recommended, different ports)
2. Register backend nodes in an NGINX `upstream`
3. Forward inbound traffic to the `upstream` via `proxy_pass`
4. Enable health checks, timeout retries, and keepalive policies as needed

### 📈 Expected Benefits

- Distributes request pressure across instances and increases overall throughput
- Reduces single-node hotspots and improves stability during peak traffic
- Supports rolling updates by node to lower release risk

## ❓ Common Issues

| Issue | Solution |
|-------|----------|
| **MySQL connection failed** | Check if account password and port are consistent between `setting.yaml` and Canal |
| **ES startup failed** | Confirm `http://127.0.0.1:9200` is accessible and account password matches |
| **Redis warning eviction policy** | Project will check dynamic Redis `maxmemory-policy` |
| **Interface 404** | Note current interfaces have `/api` prefix, should access paths like `/api/user/login` |
| **Canal no data** | Priority check if binlog is enabled, master address/account in `instance.properties`, and subscription rules match |

## 💡 Developer Only Tips

- Code contains some TODOs (e.g., ES upgrade to v8, route prefix switching, scheduled task frequency)
- For production use, recommend adding: authentication, rate limiting, monitoring, environment-specific configuration management, error recovery and idempotent processing

## 📞 Contact Us

If you encounter any problems during deployment or use, please provide feedback through the following methods:

- **🐛 GitHub Issues**: Submit Issue in [project repository](https://github.com/Bury-Lee/go-blog)
- **📧 Email Contact**: [18151161@qq.com](mailto:18151161@qq.com)
- **👥 Discussion Group**: [StarDreamer Discussion Group](https://qun.qq.com/universal-share/share?ac=1&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ&svctype=4&tempid=h5_group_info)

The author is very willing to solve valuable technical problems and welcomes PR submissions to contribute to the project!
