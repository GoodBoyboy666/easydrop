# email

`email` 是一个用于 Wire 注入的邮件发送包，基于 `xhit/go-simple-mail/v2` 实现。

## 核心能力

- 通过 `NewClient(cfg *Config)` 创建邮件客户端。
- 通过 `ProviderSet` 提供 Wire 注入入口。
- 支持 HTML 邮件发送：`SendHTML(ctx, to, subject, htmlBody)`。
- 支持三种 TLS 模式：`notls`、`ssl`、`starttls`。

## 配置项

- `Host`：SMTP 主机地址（必填）。
- `Port`：SMTP 端口（必填，>0）。
- `Username`：SMTP 用户名（必填）。
- `Password`：SMTP 密码（可选）。
- `FromEmail`：发件人邮箱（必填）。
- `TLSMode`：TLS 模式（可选，默认 `starttls`）。
- `ConnectTimeout`：连接超时（可选，默认 10 秒）。
- `SendTimeout`：发送超时（可选，默认 10 秒）。

## 注意事项

- 本包会自动去除空收件人并去重。
- 发送内容为空主题或空 HTML 正文会直接返回错误。

