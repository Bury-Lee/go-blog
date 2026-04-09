详见视频
Canal极简入门：一小时让你快速上手Canal数据同步神技~

启动canal通过bin的bat文件
要双击,不能在命令行打开




# Canal 数据库密码配置操作指南

## 概述

本文档指导您如何在 Canal 配置文件中正确设置数据库连接密码，包括目标 MySQL 数据库密码和 TSDB 元数据存储密码。

---

## 一、配置文件位置

| 配置文件 | 路径 | 用途 |
|---------|------|------|
| 全局配置 | `conf/canal.properties` | Canal Server 全局参数 |
| 实例配置 | `conf/example/instance.properties` | 单个 Canal 实例的数据库连接信息 |

> 注：`example` 是默认实例名，如果您配置了多个实例（`canal.destinations = example,test1`），则每个实例都有独立的配置文件目录。

---

## 二、配置目标 MySQL 数据库密码

### 2.1 找到实例配置文件

```bash
# 默认实例路径
conf/example/instance.properties
```

### 2.2 修改数据库连接信息

找到以下配置项并修改：

```properties
# MySQL 主库地址
canal.instance.master.address=127.0.0.1:3306

# 数据库用户名（建议单独创建 canal 用户并授权）
canal.instance.dbUsername=canal

# 【关键】数据库密码
canal.instance.dbPassword=您的实际密码
```

### 2.3 MySQL 用户授权示例

在目标 MySQL 数据库中执行：

```sql
-- 创建 canal 用户
CREATE USER 'canal'@'%' IDENTIFIED BY '您的实际密码';

-- 授权（最小权限原则）
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';

-- 刷新权限
FLUSH PRIVILEGES;
```

---

## 三、配置 TSDB 元数据存储密码

TSDB 用于存储表结构变更历史，解决 binlog 中字段 ID 与名称的映射问题。

### 3.1 方案 A：使用默认 H2 数据库（推荐测试环境）

无需额外配置，使用全局配置文件中的默认设置：

```properties
# canal.properties
canal.instance.tsdb.enable = true
canal.instance.tsdb.dbUsername = canal
canal.instance.tsdb.dbPassword = canal
```

### 3.2 方案 B：使用 MySQL 存储 TSDB（推荐生产环境）

编辑 `instance.properties`，启用 MySQL 存储：

```properties
# 启用 TSDB 并指向 MySQL
canal.instance.tsdb.enable=true
canal.instance.tsdb.url=jdbc:mysql://127.0.0.1:3306/canal_tsdb
canal.instance.tsdb.dbUsername=canal_tsdb
canal.instance.tsdb.dbPassword=您的TSDB密码
```

同时修改 `canal.properties` 中的 Spring 配置：

```properties
# 切换为 MySQL TSDB 配置
canal.instance.tsdb.spring.xml = classpath:spring/tsdb/mysql-tsdb.xml
```

创建 TSDB 数据库：

```sql
-- 创建 TSDB 专用数据库
CREATE DATABASE canal_tsdb CHARACTER SET utf8mb4;

-- 创建用户
CREATE USER 'canal_tsdb'@'127.0.0.1' IDENTIFIED BY '您的TSDB密码';
GRANT ALL PRIVILEGES ON canal_tsdb.* TO 'canal_tsdb'@'127.0.0.1';
FLUSH PRIVILEGES;
```

---

## 四、密码安全配置（生产环境推荐）

### 4.1 使用 Druid 加密密码

编辑 `instance.properties`：

```properties
# 启用 Druid 解密
canal.instance.enableDruid=true

# 配置 RSA 公钥（用于解密）
canal.instance.pwdPublicKey=MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB...

# 将 dbPassword 替换为加密后的密文
canal.instance.dbPassword=加密后的密文
```

### 4.2 生成加密密码步骤

```bash
# 使用 Canal 提供的工具生成密文
java -cp lib/* com.alibaba.druid.filter.config.ConfigTools 您的明文密码

# 输出示例：
# privateKey: xxx...
# publicKey: MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJB...
# password: 加密后的密文
```

---

## 五、配置验证步骤

### 5.1 检查配置文件语法

```bash
# 确保没有多余的空格或特殊字符
grep -E "dbPassword|dbUsername" conf/example/instance.properties
```

### 5.2 启动 Canal 并验证连接

```bash
# 启动 Canal
sh bin/startup.sh

# 查看日志确认连接成功
tail -f logs/example/example.log
```

成功标志：
```
[INFO] c.a.o.c.i.d.r.mysql.MysqlConnection - connect MysqlConnection...
[INFO] c.a.o.c.i.d.r.mysql.MysqlConnection - handshake initialization...
```

### 5.3 常见连接错误排查

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| `Access denied for user` | 用户名或密码错误 | 检查 `dbUsername` 和 `dbPassword` |
| `Unknown database` | TSDB 数据库不存在 | 创建 `canal_tsdb` 数据库 |
| `Communications link failure` | 网络或端口问题 | 检查 `master.address` 和防火墙 |
| `RSA public key not available` | MySQL 8.0 认证问题 | 修改用户认证插件为 `mysql_native_password` |

---

## 六、配置文件模板（快速复制）

### instance.properties 最小可用配置

```properties
# MySQL 连接信息
canal.instance.master.address=192.168.1.100:3306
canal.instance.dbUsername=canal_sync
canal.instance.dbPassword=StrongP@ssw0rd2024
canal.instance.connectionCharset=UTF-8

# TSDB 配置（使用 MySQL）
canal.instance.tsdb.enable=true
canal.instance.tsdb.url=jdbc:mysql://192.168.1.100:3306/canal_tsdb
canal.instance.tsdb.dbUsername=canal_tsdb
canal.instance.tsdb.dbPassword=TsdbP@ssw0rd2024

# 其他必要配置
canal.instance.gtidon=false
canal.instance.filter.regex=.*\\..*
```

---

## 七、相关配置项速查表

| 配置项 | 所在文件 | 默认值 | 说明 |
|--------|---------|--------|------|
| `canal.instance.dbPassword` | instance.properties | `canal` | **目标 MySQL 密码（必改）** |
| `canal.instance.tsdb.dbPassword` | canal.properties | `canal` | H2 元数据库密码 |
| `canal.instance.tsdb.dbPassword` | instance.properties | 注释 | MySQL 元数据库密码 |
| `canal.admin.passwd` | canal.properties | 空 | Admin 控制台密码 |
| `canal.passwd` | canal.properties | 注释 | Canal 客户端连接密码 |

---

**文档版本：** 2026-03-19  
**适用 Canal 版本：** 1.1.x 及以上