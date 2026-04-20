# GoBlog（StarDreamerCyberNook）

`Gin + GORM + Redis + Elasticsearch`をベースとしたブログ/コミュニティバックエンドプロジェクトで、ユーザー、記事、コメント、メッセージ、フォロー、チャット、サイト設定などのモジュールを含んでいます。

## 🌐 多言語ドキュメント
**言語を選択：**:[🇨🇳 中文](README.zh-CN.md) • [🇺🇸 English](../README.md) • [🇯🇵 日本語](README.ja.md) • [🇰🇷 한국어](README.ko.md)

## � 対象ユーザー

- 自立と主体性を重視する技術愛好家：第三者プラットフォームのルールに縛られず、完全にコントロールできる個人ブログ/コミュニティを持ちたい。カスタマイズ性とデータ主権を重視
- 小規模サーバーを持つサイト運営者：VPS/クラウド/自宅サーバーなどの基礎的なリソースがあり、安定したテキスト/画像コンテンツサービスを構築したい
- コミュニティ運営者：特定分野の交流コミュニティ、公式フォーラム、軽量コンテンツプラットフォームを構築し、ユーザーの交流、コンテンツ審査、SEO最適化を行いたい
- プライバシーとデータ安全性を重視するユーザー：重要なコンテンツデータを商用プラットフォームに預けたくなく、ローカル配備と自主バックアップを実現したい

## �📋 機能概要

- **ブログとコミュニティの一体化**：記事、コメント、コレクション、メッセージ、フォロー、プライベートチャットを同じバックエンドに集中
- **完全な検索機能**：`Elasticsearch`に基づく記事の全文検索、ハイライト、多次元ソート
- **高並列処理対応**：閲覧、いいね、コレクション、コメント数は最初にRedisに書き込まれ、後で定時タスクでデータベースにバッチ同期
- **AIビジネス統合**：サイトAIアシスタント、記事、コメント、ニックネームなどのコンテンツ審査をサポート
- **完全な運用機能**：サイト設定、SEO、バナー、フレンドリンク、プロモーションスロット、ログ管理をサポート

> 📖 **機能ドキュメント**：[ブログ機能ドキュメント](功能文档.md)

> ⚠️ **注意**：プロジェクト本体はデータベースの読み書き分離をサポートしていますが、**データベース間のデータ同期をネイティブにサポートしていません**。リポジトリの`otherSoftware`はCanalベースの簡単な購読同期ソリューションを提供しています（設定と使用手順は本文で提供されています）。

## 🏗️ プロジェクト構造

```
go-blog
├─ api/                 # コントローラーレイヤー
├─ router/              # ルート登録
├─ models/              # データモデルとESマッピング
├─ service/             # ビジネスサービス（定時タスク、ESサービス、Redisサービスを含む）
├─ middleware/          # ミドルウェア
├─ core/                # 設定/ログ/DB/Redis/ES/AI初期化
├─ conf/                # 設定構造体
├─ flags/               # コマンドラインパラメータ（移行、インデックス作成、ユーザー作成）
├─ init/                # ローカル依存サービスdocker-composeと基本設定
├─ otherSoftware/       # データ同期ソリューション（Canal + canal-go）
├─ setting.yaml         # メイン設定ファイル
└─ main.go              # エントリーポイント
```

## 🔧 環境要件

| コンポーネント | バージョン要件 | 備考 |
|----------|------------|------|
| Go | 1.26.0 | プロジェクトの`go.mod`バージョンに従う |
| MySQL | 5.7 | プロジェクト内でdocker-composeを提供 |
| Redis | 7 | プロジェクト内でdocker-composeを提供、デュアルインスタンス |
| Elasticsearch | 7.17.x | プロジェクトは`olivere/elastic/v7`を使用 |
| AIサービス | オプション | デフォルト設定例：`http://localhost:1234/v1`（現在はローカルモデルのみサポート） |

## 🚀 クイックスタート

### 1️⃣ 依存サービスを起動

プロジェクトルートディレクトリ`go-blog`でそれぞれ実行：

```bash
# MySQLを起動
docker compose -f init/SQL/docker-compose.yml up -d

# Redisを起動
docker compose -f init/Redis/docker-compose.yml up -d

# Elasticsearchを起動
docker compose -f init/ES/docker-compose.yml up -d
```

設定ファイルを書いた後、リリースプログラムを実行：

Windows
```bash
# プロジェクトを起動
.\main_windows_amd64.exe
```
Linux
```bash
# プロジェクトを起動
./main_linux_amd64
```
macOS
```bash
# プロジェクトを起動
./main_macos_amd64
```



> 💡 **ヒント**：MySQL 5.7以上のバージョンとCanalの互換性があまり良くないため、`init/SQL/docker-compose.yml`でバージョン`5.7.360.0`を設定しています。
### 2️⃣ プロジェクトをコンパイルする

Windowsでコンパイル:

```powershell
# Windows AMD64
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_windows_amd64.exe .\main.go

# Linux AMD64
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_linux_amd64 .\main.go

# macOS AMD64
$env:GOOS="darwin"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -ldflags="-s -w" -trimpath -o main_macos_amd64 .\main.go
```

### 3️⃣ メイン設定`setting.yaml`を変更

以下の重要なフィールドを優先的に確認することをお勧めします：

- `system.ip`、`system.port`、`system.env`、`system.gin_mode`
- `db`（最初が書き込みデータベース、後続が読み取りデータベース）
- `redisStatic`、`redisDynamic`
- `es.url`、`es.username`、`es.password`
- `jwt.accessTokenSecret`、`jwt.refreshTokenSecret`
- `email`（メール確認コードログイン/登録用）
- `ai.enable`（AIを使用しない場合は`false`に設定）

<details>
<summary>📖 完全な設定例を表示</summary>

```yaml
system:
  ip: 0.0.0.0
  port: 8080
  env: dev
  gin_mode: debug # Gin実行モード：debugまたはrelease

log:
  app: GoBlog
  dir: log
  log_level: debug

ai:
  enable: true
  model: local
  host: http://localhost:1234/v1
  ApiKey: 123456
  nickName: ニックネーム
  avatar: "https://example.com/avatar.jpg" # TODO：後でアバター源を適応

email:
  domain: smtp.qq.com
  port: 587 # QQメールの一般的なポートは587または465
  sendEmail: xxxxxx@qq.com
  authCode: "xxxxxxx" # メール承認コード
  sendNickname: ニックネーム

upload:
  size: 20 # アップロードファイルサイズ制限、単位MB
  whiteList: # アップロードファイルホワイトリスト
    ".jpg": ~
    ".jpeg": ~
    ".png": ~
    ".gif": ~
    ".webp": ~
    ".bmp": ~
    ".tiff": ~
    ".svg": ~
  uploadDir: images # 画像アップロードディレクトリ

jwt:
  accessExpire: 30 # アクセストークン有効期限（分）、推奨30分
  refreshExpire: 172 # リフレッシュトークン有効期限（時間）、推奨1週間（172時間）
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
  # 複数データベース読み書き分離：最初が書き込みデータベース、残りが読み取りデータベース
  - user: root
    password: root
    host: 127.0.0.1
    port: 3306
    db_name: db
    debug: false # デバッグを有効にするか（完全なログを印刷）
    sql_name: mysql
  # - user: root # この形式で複数のデータベースを追加できます
  #   password: root
  #   host: 127.0.0.1
  #   port: 3306
  #   db: db
  #   debug: false # デバッグを有効にするか（完全なログを印刷）
  #   source: mysql

site:
  siteInfo:
    title: "星夢ネットワークスペース" # サイトタイトル
    logo: "/static/images/logo.png" # サイトロゴパス
    beian: "京ICP备XXXXXXXX号" # 登録番号
    mode: 1 # 実行モード（1: ブログモード, 2: コミュニティモードなど、コード列挙に対応する必要がある）

  project:
    title: "StarDreamer" # プロジェクト名
    icon: "/static/images/favicon.ico" # プロジェクトアイコン
    webPath: "https://www.example.com" # プロジェクトアクセスパス

  seo:
    keywords: "技術ブログ, Go言語, 人工知能, 共有"
    description: "技術共有と人工知能探索に特化した個人サイト。"

  about:
    siteDate: "2023-01-01"
    qq: "123456789"
    wechat: "StarDreamer_Official"
    biliBili: "https://space.bilibili.com/your_uid"
    gitHub: "https://github.com/your_username"

  indexRight:
    list:
      - title: "人気記事"
        enable: true
      - title: "最新コメント"
        enable: true
      - title: "フレンドリンク"
        enable: true
      - title: "タグクラウド"
        enable: true

  article:
    # 説明：Goフィールド名はDisableExaminationですが、yamlタグはenableExaminationです
    # enableExaminationとして読み取る場合：trueは「レビューを有効にする」（つまり無効にしない）を意味します
    # ビジネスセマンティクス：true = レビューが必要、false = レビュー不要
    enableExamination: true

  login:
    QQLogin: true # TODO：まだ実装されていません
    usernamePassword: true
    emailLogin: true # TODO：常時オンを検討できます（基本的なログイン方法）
    captcha: false # キャプチャを有効にするか
```

</details>

### 4️⃣ データベース構造を初期化

```bash
go run main.go -db
```

### 5️⃣ ESインデックスを初期化

```bash
go run main.go -es
```

### 6️⃣ サービスを起動

```bash
go run main.go -f setting.yaml
```

🌐 起動後のデフォルトリスニング：`http://127.0.0.1:8080`

## 📝 コマンドラインパラメータ

| パラメータ | 説明 |
|----------|------|
| `-f` | 設定ファイルパス（デフォルト`setting.yaml`） |
| `-db` | GORM自動移行を実行 |
| `-es` | ESインデックスを作成/再構築 |
| `-v` | バージョンを表示 |
| `-t user -s create` | コマンドラインでユーザーを作成 |

**例：**
```bash
go run main.go -t user -s create
```

## ⚙️ 重要な実行説明

- **初期化順序**：設定 → ログ → IPライブラリ → DB → Redis → ES → AI → 定時タスク → ルート
- **ESは強力な依存関係**：ES初期化失敗はサービス起動中断を引き起こします
- **インターフェース仕様**：現在のインターフェースは`/api`プレフィックスを統一的に使用、例：`/api/user/login`
- **静的リソース**：静的リソースディレクトリは`/web`としてマッピングされ、ローカルの`static/`に対応
- **定時タスク**：現在の定時タスク`SyncArticle/SyncComment`は10分レベルのバッチ同期で、Redisのカウント増分を更新するために使用

## 🔄 データベース同期ソリューション（otherSoftware）

### 📍 背景

プロジェクトはネイティブに読み書き分離のみを行い、クロスデータベース自動同期は行いません。  
`otherSoftware`は簡単なソリューションを提供します：

- `otherSoftware/canal`：Canalサーバー、MySQLバイナリログを購読
- `otherSoftware/canal-go/canal-go-1.1.2/samples/main.go`：Goクライアントが変更を消費し、ダウンストリーム処理を行います（例ではESに同期）

### ✅ 前提条件

1. MySQLバイナリログが有効（プロジェクト`init/SQL/master/my.cnf`はすでに`log-bin`、`server-id`を含む）
2. Canal接続情報とMySQLの実際のアカウントパスワードが一致
3. `canal.instance.master.address`は購読するMySQLを指す（通常はマスター`127.0.0.1:3306`）
4. `setting.yaml`、MySQL、Canalの3か所でパスワードが一致することをお勧め

### 🔧 Canalを設定

ファイルを編集：

- `otherSoftware/canal/canal.deployer-1.1.8/conf/canal.properties`
- `otherSoftware/canal/canal.deployer-1.1.8/conf/example/instance.properties`

**重点を置く：**

- `canal.destinations=example`
- `canal.port=11111`
- `canal.instance.master.address=127.0.0.1:3306`
- `canal.instance.dbUsername=...`
- `canal.instance.dbPassword=...`
- `canal.instance.filter.regex=.*\..*`（ターゲットテーブルのみをリッスンするように変更可能）

### 🚀 Canalを起動

リポジトリの説明に従い、`startup.bat`を直接使用して起動（Windows環境）：

- **Windows**：`otherSoftware/canal/canal.deployer-1.1.8/bin/startup.bat`をダブルクリック

その後、ログを確認：

- `otherSoftware/canal/canal.deployer-1.1.8/logs/example/example.log`

MySQL接続成功の関連ログが見えれば、Canal側が正常であることを示します。

### 🎯 canal-goコンシューマを起動

`otherSoftware/canal-go/canal-go-1.1.2`ディレクトリで実行：

```bash
go run samples/main.go
```

この例のデフォルト：

- **Canalに接続**：`127.0.0.1:11111`
- **宛先**：`example`
- **購読ルール**：`.*\.article_models`
- **処理ロジック**：バイナリログの変更を読み取り、変換後に`article_index`（ES）に書き込む

ESではなく「別のデータベース」に同期したい場合は、`samples/ES_pkg`の処理ロジックを独自のDB書き込みロジックに置き換えてください。

### 📋 使用フローの推奨

1. 最初にMySQL/Redis/ESとGoBlogメインサービスを起動
2. Canalサーバーを起動（バイナリログ購読ストリームを生成）
3. canal-goコンシューマを起動（変換とDB/ESへの書き込みを実行）
4. マスターデータベースで`insert/update/delete`を実行し、ターゲットエンドが同期されていることを確認

## 🌐 NGINX 水平スケーリング（任意）

最近のアーキテクチャ調整により、サービス層は実質的にステートレス化され、業務データは `MySQL / Redis / Elasticsearch` に集約されています。  
この前提により、`NGINX` のリバースプロキシとロードバランシングを使って水平拡張を行えます。

### ✅ 適したケース

- 単一インスタンスの CPU または接続数が上限に近く、並行処理性能を段階的に引き上げたい
- 複数ホストまたは複数コンテナを運用しており、外部公開入口を 1 つに統一したい
- サービス停止なしで段階的な増設やノード入れ替えを行いたい

### 🔧 実装ポイント

1. 複数の GoBlog インスタンスを起動（同一バージョン・同一設定を推奨、ポートのみ分離）
2. NGINX の `upstream` にバックエンドノードを登録
3. `proxy_pass` で入口トラフィックを `upstream` に転送
4. 必要に応じてヘルスチェック、タイムアウト再試行、keepalive などを有効化

### 📈 期待できる効果

- リクエスト負荷を複数ノードへ分散し、全体スループットを向上
- 単一点への負荷集中を抑え、ピーク時の応答安定性を改善
- ノード単位のローリング更新が可能になり、リリースリスクを低減

## ❓ よくある質問

| 問題 | 解決策 |
|------|--------|
| **MySQL接続失敗** | `setting.yaml`とCanalのアカウントパスワード、ポートが一致しているか確認 |
| **ES起動失敗** | `http://127.0.0.1:9200`にアクセスでき、アカウントパスワードが一致しているか確認 |
| **Redis警告除去ポリシー** | プロジェクトは動的Redisの`maxmemory-policy`をチェックします |
| **インターフェース404** | 現在のインターフェースに`/api`プレフィックスが付いていることに注意し、`/api/user/login`のようなパスにアクセス |
| **Canalデータなし** | バイナリログが有効になっているか、`instance.properties`のマスターアドレス/アカウント、購読ルールが一致しているかを優先的に確認 |

## 💡 開発者向けヒント

- コードにはTODOが含まれています（ES v8へのアップグレード、ルートプレフィックス切り替え、定時タスク頻度など）
- 本番使用の場合、以下を追加することをお勧め：認証、レート制限、モニタリング、環境固有の設定管理、エラー回復とべき等処理

## 📞 お問い合わせ

デプロイメントまたは使用中に問題が発生した場合は、以下の方法でフィードバックをお願いします：

- **🐛 GitHub Issues**： [プロジェクトリポジトリ](https://github.com/Bury-Lee/go-blog)でIssueを提出
- **📧 メール連絡**： [18151161@qq.com](mailto:18151161@qq.com)
- **👥 ディスカッショングループ**： [星夢の交流グループ](https://qun.qq.com/universal-share/share?ac=1&authKey=2MOKPRKsyf8SGY12y3L%2By8yC53zfKakQDg5qiZvgz46DHm%2Bil90q6MuER5XVKo4g&busi_data=eyJncm91cENvZGUiOiIxMDk4NDgzNzk0IiwidG9rZW4iOiJMdTVWVWFQK3pMYXdteDdrVzF5MzE1Nm12SDlHLy9PYm1zZXJBUm5peGxKcGptdHoxcXhacWtsSlNNTDN6S3hVIiwidWluIjoiMTgxNTExNjEifQ%3D%3D&data=71mrINsJgoFhsfYAIO6n6qMWIh9Fi73oWgVrPeRDFjKIwlhBnVaCGFKx5Hr73xvNrEsKaAIk-gvPCV2nkslvHQ&svctype=4&tempid=h5_group_info)

著者は貴重な技術問題の解決を非常に喜んでおり、プロジェクトへの貢献のためにPRの提出も歓迎します！
