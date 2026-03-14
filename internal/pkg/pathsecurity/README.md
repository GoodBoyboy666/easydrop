# pathsecurity

`pathsecurity` 是一个无状态路径安全工具包，用于防范路径穿越攻击。

## 能力

- `SecureJoin(baseDir, userPath)`：安全拼接用户路径并返回绝对路径。
- `IsPathUnderBase(baseDir, targetPath)`：判断目标路径是否位于基础目录内。

## 安全规则

- 拒绝空路径与包含 `NUL` 字符的路径。
- 拒绝绝对路径与盘符注入。
- 拒绝路径穿越（如 `../`）。
- 检测目录链路中存在符号链接时直接拒绝。
- 允许目标文件不存在，但其父目录必须存在且通过符号链接检查。

