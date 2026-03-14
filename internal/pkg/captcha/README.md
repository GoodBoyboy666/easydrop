# captcha

`captcha` 是一个无状态验证码工具包，调用时传入 `Config` 即可完成校验。

## 支持类型

- `geetest_v4`
- `hcaptcha`
- `recaptcha`
- `turnstile`

## 核心接口

- `Verify(ctx, cfg, payload)`：统一验证码校验入口。

## 设计说明

- 包本身无全局状态，配置通过每次调用传入。

