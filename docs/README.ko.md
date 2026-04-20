# GoBlog（StarDreamerCyberNook）

`Gin + GORM + Redis + Elasticsearch`를 기반으로 하는 블로그/커뮤니티 백엔드 프로젝트로, 사용자, 게시물, 댓글, 메시지, 팔로우, 채팅, 사이트 설정 등의 모듈을 포함합니다.

## 🌐 다국어 문서

**언어 선택:**:[🇨🇳 中文](README.zh-CN.md) • [🇺🇸 English](../README.md) • [🇯🇵 日本語](README.ja.md) • [🇰🇷 한국어](README.ko.md)

## 👤 추천 대상

- 독립성과 자율성을 중시하는 기술 애호가: 서드파티 콘텐츠 플랫폼의 규칙에 의존하지 않고, 완전히 통제 가능한 개인 블로그나 커뮤니티를 운영하며 커스터마이징과 데이터 주권을 중요하게 생각하는 경우
- 소형 서버를 보유한 사이트 운영자: VPS, 클라우드 서버, 홈 서버 등 기본적인 서버 자원이 있고 안정적인 텍스트·이미지 콘텐츠 서비스를 구축하려는 경우
- 커뮤니티 운영자: 특정 분야 커뮤니티, 공식 포럼, 경량 콘텐츠 플랫폼을 구축하고 사용자 상호작용, 콘텐츠 검토, SEO 최적화를 지원하려는 경우
- 개인정보와 데이터 보안을 중시하는 사용자: 핵심 콘텐츠 데이터를 상용 플랫폼에 맡기지 않고 로컬 배포와 자체 백업을 원할 경우

## 📋 기능 개요

- **블로그와 커뮤니티 일체화**: 게시물, 댓글, 컬렉션, 메시지, 팔로우, 비공개 채팅을 동일한 백엔드에 집중
- **완전한 검색 기능**: `Elasticsearch` 기반 게시물 전문 검색, 하이라이트, 다차원 정렬
- **고성능 병렬 처리 지원**: 조회수, 좋아요, 컬렉션, 댓글 수는 먼저 Redis에 기록되고, 이후 정기 작업으로 데이터베이스에 일괄 동기화
- **AI 비즈니스 통합**: 사이트 AI 어시스턴트, 게시물, 댓글, 닉네임 등 콘텐츠 검토 지원
- **완전한 운영 기능**: 사이트 설정, SEO, 배너, 친구 링크, 프로모션 슬롯, 로그 관리 지원

> 📖 **기능 문서**: [블로그 기능 문서](功能文档.md)

> ⚠️ **주의**: 프로젝트 본체는 데이터베이스 읽기/쓰기 분리를 지원하지만, **데이터베이스 간 데이터 동기화를 네이티브로 지원하지 않습니다**. 저장소의 `otherSoftware`는 Canal 기반의 간단한 구독 동기화 솔루션을 제공합니다 (설정 및 사용 단계는 본 문서에 제공되어 있습니다).

## 🏗️ 프로젝트 구조

```
go-blog
├─ api/                 # 컨트롤러 레이어
├─ router/              # 라우트 등록
├─ models/              # 데이터 모델 및 ES 매핑
├─ service/             # 비즈니스 서비스 (정기 작업, ES 서비스, Redis 서비스 포함)
├─ middleware/          # 미들웨어
├─ core/                # 설정/로그/DB/Redis/ES/AI 초기화
├─ conf/                # 설정 구조체
├─ flags/               # 명령줄 매개변수 (마이그레이션, 인덱스 생성, 사용자 생성)
├─ init/                # 로컬 종속 서비스 docker-compose 및 기본 설정
├─ otherSoftware/       # 데이터 동기화 솔루션 (Canal + canal-go)
├─ setting.yaml         # 메인 설정 파일
└─ main.go              # 진입점
```

## 🔧 환경 요구사항

| 구성요소 | 버전 요구사항 | 비고 |
|----------|-------------|------|
| Go | 1.26.0 | 프로젝트 `go.mod` 버전에 따름 |
| MySQL | 5.7 | 프로젝트 내 docker-compose 제공 |
| Redis | 7 | 프로젝트 내 docker-compose 제공, 이중 인스턴스 |
| Elasticsearch | 7.17.x | 프로젝트는 `olivere/elastic/v7` 사용 |
| AI 서비스 | 선택사항 | 기본 설정 예시: `http://localhost:1234/v1` (현재 로컬 모델만 지원) |

## 🚀 빠른 시작

### 1️⃣ 종속 서비스 시작

프로젝트 루트 디렉토리 `go-blog`에서 각각 실행:

```bash
# MySQL 시작
docker compose -f init/SQL/docker-compose.yml up -d

# Redis 시작
docker compose -f init/Redis/docker-compose.yml up -d

# Elasticsearch 시작
docker compose -f init/ES/docker-compose.yml up -d
```

설정 파일을 작성한 후 릴리스 프로그램 실행:

Windows
```bash
# 프로젝트 시작
.\main_windows_amd64.exe
```
Linux
```bash
# 프로젝트 시작
./main_linux_amd64
```
macOS
```bash
# 프로젝트 시작
./main_macos_amd64
```



> 💡 **팁**: MySQL 5.7 이상 버전과 Canal 간의 호환성이 그다지 좋지 않으므로, `init/SQL/docker-compose.yml`에서 버전 `5.7.360.0`을 구성했습니다.

### 2️⃣ 프로젝트 컴파일

Windows에서 컴파일:

```powershell
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

### 3️⃣ 메인 설정 `setting.yaml` 수정

다음의 중요한 필드를 우선적으로 확인하는 것이 좋습니다:

- `system.ip`, `system.port`, `system.env`, `system.gin_mode`
- `db` (첫 번째가 쓰기 데이터베이스, 이후가 읽기 데이터베이스)
- `redisStatic`, `redisDynamic`
- `es.url`, `es.username`, `es.password`
- `jwt.accessTokenSecret`, `jwt.refreshTokenSecret`
- `email` (이메일 확인 코드 로그인/등록용)
- `ai.enable` (AI를 사용하지 않으면 `false`로 설정)

<details>
<summary>📖 완전한 설정 예시 보기</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  gin_mode: debug # Gin 실행 모드: debug 또는 release

log:
  app: GoBlog
  dir: log
  log_level: debug

ai:
  enable: true
  model: local
  host: http://localhost:1234/v1
  ApiKey: 123456
  nickName: 닉네임
  avatar: "https://example.com/avatar.jpg" # TODO: 나중에 아바타 소스 적응

email:
  domain: smtp.qq.com
  port: 587 # QQ 이메일 일반적인 포트는 587 또는 465
  sendEmail: xxxxxx@qq.com
  authCode: "xxxxxxx" # 이메일 승인 코드
  sendNickname: 닉네임

upload:
  size: 20 # 업로드 파일 크기 제한, 단위 MB
  whiteList: # 업로드 파일 화이트리스트
    ".jpg": ~
    ".jpeg": ~
    ".png": ~
    ".gif": ~
    ".webp": ~
    ".bmp": ~
    ".tiff": ~
    ".svg": ~
  uploadDir: images # 이미지 업로드 디렉토리

jwt:
  accessExpire: 30 # 액세스 토큰 만료 시간 (분), 권장 30분
  refreshExpire: 172 # 리프레시 토큰 만료 시간 (시간), 권장 1주일 (172시간)
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
  # 다중 데이터베이스 읽기/쓰기 분리: 첫 번째가 쓰기 데이터베이스, 나머지가 읽기 데이터베이스
  - user: root
    password: root
    host: 127.0.0.1
    port: 3306
    db_name: db
    debug: false # 디버그 활성화 여부 (완전한 로그 출력)
    sql_name: mysql
  # - user: root # 이 형식으로 계속해서 여러 데이터베이스 추가 가능
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db: db
  #   debug: false # 디버그 활성화 여부 (완전한 로그 출력)
  #   source: mysql

site:
  siteInfo:
    title: "성망 네트워크 공간" # 사이트 제목
    logo: "/static/images/logo.png" # 사이트 로고 경로
    beian: "京ICP备XXXXXXXX号" # 등록 번호
    mode: 1 # 실행 모드 (1: 블로그 모드, 2: 커뮤니티 모드 등, 코드 열거에 대응해야 함)

  project:
    title: "StarDreamer" # 프로젝트 이름
    icon: "/static/images/favicon.ico" # 프로젝트 아이콘
    webPath: "https://www.example.com" # 프로젝트 접속 경로

  seo:
    keywords: "기술 블로그, Go 언어, 인공지능, 공유"
    description: "기술 공유와 인공지능 탐구에 집중하는 개인 사이트."

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "인기 게시물"
        enable: true
      - title: "최신 댓글"
        enable: true
      - title: "친구 링크"
        enable: true
      - title: "태그 클라우드"
        enable: true

  article:
    # 설명: Go 필드명은 DisableExamination이지만, yaml 태그는 enableExamination입니다
    # enableExamination으로 읽기: true는 "검토 활성화" (즉, 비활성화하지 않음)를 의미
    # 비즈니스 의미: true = 검토 필요, false = 검토 불필요
    enableExamination: true

  login:
    QQLogin: true # TODO: 아직 구현되지 않음
    usernamePassword: true
    emailLogin: true # TODO: 항상 켜기 고려 가능 (기본 로그인 방식)
    captcha: false # 캡차 활성화 여부
```

</details>

### 4️⃣ 데이터베이스 구조 초기화

```bash
go run main.go -db
```

### 5️⃣ ES 인덱스 초기화

```bash
go run main.go -es
```

### 6️⃣ 서비스 시작

```bash
go run main.go -f setting.yaml
```

🌐 시작 후 기본 리스닝: `http://127.0.0.1:8080`

## 📝 명령줄 매개변수

| 매개변수 | 설명 |
|----------|------|
| `-f` | 설정 파일 경로 (기본값 `setting.yaml`) |
| `-db` | GORM 자동 마이그레이션 실행 |
| `-es` | ES 인덱스 생성/재구축 |
| `-v` | 버전 보기 |
| `-t user -s create` | 명령줄로 사용자 생성 |

**예시:**
```bash
go run main.go -t user -s create
```

## ⚙️ 주요 실행 설명

- **초기화 순서**: 설정 → 로그 → IP 라이브러리 → DB → Redis → ES → AI → 정기 작업 → 라우트
- **ES는 강력한 종속성**: ES 초기화 실패는 서비스 시작 중단을 일으킴
- **인터페이스 사양**: 현재 인터페이스는 `/api` 접두사를 통일적으로 사용, 예: `/api/user/login`
- **정적 리소스**: 정적 리소스 디렉토리는 `/web`으로 매핑되며, 로컬 `static/`에 해당
- **정기 작업**: 현재 정기 작업 `SyncArticle/SyncComment`는 10분 수준의 일괄 동기화로, Redis의 카운트 증분을 새로고침하는 데 사용

## 🔄 데이터베이스 동기화 솔루션 (otherSoftware)

### 📍 배경

프로젝트는 네이티브로 읽기/쓰기 분리만 수행하며, 교차 데이터베이스 자동 동기화는 수행하지 않습니다.  
`otherSoftware`는 간단한 솔루션을 제공합니다:

- `otherSoftware/canal`: Canal 서버, MySQL 바이너리 로그 구독
- `otherSoftware/canal-go/canal-go-1.1.2/samples/main.go`: Go 클라이언트가 변경사항을 소비하고 다운스트림 처리 수행 (예시에서는 ES에 동기화)

### ✅ 전제 조건

1. MySQL 바이너리 로그 활성화 (프로젝트 `init/SQL/master/my.cnf`는 이미 `log-bin`, `server-id`를 포함)
2. Canal 연결 정보와 MySQL 실제 계정 비밀번호가 일치
3. `canal.instance.master.address`는 구독할 MySQL을 가리킴 (보통 마스터 `127.0.0.1:3306`)
4. `setting.yaml`, MySQL, Canal 세 곳의 비밀번호가 일치하도록 권장

### 🔧 Canal 설정

파일 편집:

- `otherSoftware/canal/canal.deployer-1.1.8/conf/canal.properties`
- `otherSoftware/canal/canal.deployer-1.1.8/conf/example/instance.properties`

**주요 초점:**

- `canal.destinations=example`
- `canal.port=11111`
- `canal.instance.master.address=127.0.0.1:3306`
- `canal.instance.dbUsername=...`
- `canal.instance.dbPassword=...`
- `canal.instance.filter.regex=.*\..*` (대상 테이블만 수신하도록 변경 가능)

### 🚀 Canal 시작

저장소 설명에 따라 `startup.bat`를直接使用하여 시작 (Windows 환경):

- **Windows**: `otherSoftware/canal/canal.deployer-1.1.8/bin/startup.bat` 더블클릭

그런 다음 로그 확인:

- `otherSoftware/canal/canal.deployer-1.1.8/logs/example/example.log`

MySQL 연결 성공 로그가 보이면 Canal 측이 정상임을 나타냅니다.

### 🎯 canal-go 소비자 시작

`otherSoftware/canal-go/canal-go-1.1.2` 디렉토리에서 실행:

```bash
go run samples/main.go
```

이 예시의 기본값:

- **Canal 연결**: `127.0.0.1:11111`
- **대상**: `example`
- **구독 규칙**: `.*\.article_models`
- **처리 로직**: 바이너리 로그 변경사항을 읽고, 변환 후 `article_index`(ES)에 쓰기

ES가 아닌 "다른 데이터베이스"에 동기화하려면 `samples/ES_pkg`의 처리 로직을 자체 DB 쓰기 로직으로 교체하십시오.

### 📋 권장 사용 프로세스

1. 먼저 MySQL/Redis/ES와 GoBlog 메인 서비스를 시작
2. Canal 서버 시작 (바이너리 로그 구독 스트림 생성)
3. canal-go 소비자 시작 (변환 및 DB/ES 쓰기 실행)
4. 마스터 데이터베이스에서 `insert/update/delete`를 실행하여 대상 엔드가 동기화되는지 확인

## 🌐 NGINX 수평 확장 (선택)

최근 아키텍처 조정을 통해 서비스 레이어는 사실상 무상태(stateless) 구조가 되었고, 비즈니스 데이터는 `MySQL / Redis / Elasticsearch` 에 통합 저장됩니다.  
이 구조를 기반으로 `NGINX` 리버스 프록시와 로드 밸런싱을 사용해 수평 확장을 구성할 수 있습니다.

### ✅ 적용 시나리오

- 단일 인스턴스의 CPU 또는 연결 수가 한계에 가까워져 동시 처리 성능을 점진적으로 높여야 하는 경우
- 여러 호스트/컨테이너 인스턴스를 운영 중이며 외부 진입점을 하나로 통합하려는 경우
- 서비스 중단 없이 점진적 증설 또는 백엔드 노드 교체를 진행하려는 경우

### 🔧 구현 포인트

1. GoBlog 인스턴스를 여러 개 기동 (버전/설정은 동일 권장, 포트만 분리)
2. NGINX `upstream` 에 백엔드 노드를 등록
3. `proxy_pass` 로 유입 트래픽을 `upstream` 으로 전달
4. 필요에 따라 헬스체크, 타임아웃 재시도, keepalive 정책을 활성화

### 📈 기대 효과

- 요청 부하를 다중 인스턴스로 분산하여 전체 처리량 향상
- 단일 노드 과부하를 줄여 피크 시간대 응답 안정성 개선
- 노드 단위 롤링 배포가 가능해져 변경 리스크 완화

## ❓ 자주 묻는 질문

| 문제 | 해결책 |
|------|--------|
| **MySQL 연결 실패** | `setting.yaml`와 Canal의 계정 비밀번호, 포트가 일치하는지 확인 |
| **ES 시작 실패** | `http://127.0.0.1:9200`에 접근할 수 있고 계정 비밀번호가 일치하는지 확인 |
| **Redis 경고 제거 정책** | 프로젝트는 동적 Redis의 `maxmemory-policy`를 확인합니다 |
| **인터페이스 404** | 현재 인터페이스에 `/api` 접두사가 있음을 주의하고, `/api/user/login` 같은 경로에 접근 |
| **Canal 데이터 없음** | 바이너리 로그가 활성화되어 있는지, `instance.properties`의 마스터 주소/계정, 구독 규칙이 일치하는지를 우선적으로 확인 |

## 💡 개발자용 팁

- 코드에 TODO가 포함되어 있습니다 (예: ES v8로 업그레이드, 라우트 접두사 전환, 정기 작업 빈도)
- 프로덕션 사용 시 다음을 추가하는 것이 좋습니다: 인증, 속도 제한, 모니터링, 환경별 설정 관리, 오류 복구 및 멱등 처리

## 📞 문의하기

배포 또는 사용 중 문제가 발생한 경우, 다음 방법으로 피드백을 제공해 주십시오:

- **🐛 GitHub Issues**: [프로젝트 저장소](https://github.com/Bury-Lee/go-blog)에 Issue 제출
- **📧 이메일 연락**: [18151161@qq.com](mailto:18151161@qq.com)
- **👥 토론 그룹**: [성몽의 교류 그룹](https://qun.qq.com/universal-share/share?ac=1&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ&svctype=4&tempid=h5_group_info)

작성자는 귀중한 기술 문제 해결을 매우 기꺼이 하며, 프로젝트 기여를 위한 PR 제출도 환영합니다!
