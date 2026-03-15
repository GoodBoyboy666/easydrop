# database

`database` 是一个用于 Wire 注入的数据库初始化包，统一提供 `*gorm.DB` 实例。

## 支持的数据库类型

- `mysql`
- `postgres`
- `sqlite`

## 核心接口

- `NewDB(cfg *Config) (*gorm.DB, error)`

## 配置说明

- `Driver`：数据库类型，必填。
- `Host`：主机地址（`mysql/postgres` 必填）。
- `Port`：端口（`mysql/postgres` 必填且 > 0）。
- `Username`：用户名（`mysql/postgres` 必填）。
- `Password`：密码（可选）。
- `DatabaseName`：数据库名（`mysql/postgres` 必填）。
- `SSLMode`：`postgres` 的 sslmode（可选，默认 `disable`）。
- `TimeZone`：`postgres` 时区（可选）。
- `Charset`：`mysql` 字符集（可选，默认 `utf8mb4`）。
- `ParseTime`：`mysql` parseTime 参数（可选）。
- `Loc`：`mysql` loc 参数（可选，默认 `Local`）。
- `SQLitePath`：`sqlite` 数据库路径（`sqlite` 必填）。
- `MaxOpenConns`：最大连接数，可选。
- `MaxIdleConns`：最大空闲连接数，可选。
- `ConnMaxLifetime`：连接最大生命周期，可选。
- `ConnMaxIdleTime`：连接最大空闲时间，可选。

