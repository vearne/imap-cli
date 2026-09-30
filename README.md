# imap-cli

一个面向"人 + 大模型"的 IMAP 命令行工具，基于 [emersion/go-imap v2](https://github.com/emersion/go-imap)。
典型用法：让 LLM 通过 `list` 找出目标邮件（拿 UID），确认后用 `delete` 批量删除旧邮件。

## 安装

```sh
go install github.com/vearne/imap-cli@latest
```

## 配置（环境变量）

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `IMAP_HOST` | 是 | — | 如 `imap.gmail.com` |
| `IMAP_PORT` | 否 | `993` | |
| `IMAP_USERNAME` | 是 | — | |
| `IMAP_PASSWORD` | 是 | — | Gmail 用应用专用密码；QQ/163 用授权码（不是网页登录密码） |
| `IMAP_TLS` | 否 | `tls` | `tls` / `starttls` / `none` |

使用前需在邮箱设置中开启 IMAP 服务。

## 命令

```sh
imap-cli folders [--counts]                     # 列出文件夹
imap-cli list [--from S] [--subject S] [--since DATE] [--before DATE]
               [--unread] [--limit N] [--folder F] [--json]
                                                # 搜索并列出邮件（新→旧）
imap-cli read <uid> [--raw] [--folder F] [--json]
                                                # 读单封邮件
imap-cli mark <uid-set> read|unread|flagged|unflagged [--folder F]
imap-cli move <uid-set> <dest-folder> [--folder F]
imap-cli delete <uid-set> [--dry-run] [--folder F]
```

`<uid-set>` 语法：`5`、`1:500`、`1:500,800,1000:1200`。

所有命令支持 `--json` 输出结构化 JSON，方便程序/大模型消费。

## 典型工作流（批量删除旧邮件）

```sh
# 1. 找出 2024 年之前的订阅邮件
imap-cli list --before 2024-01-01 --from newsletter@example.com --json
# 2. 核对将删除的范围
imap-cli delete 1:500 --dry-run
# 3. 执行删除（不可恢复！）
imap-cli delete 1:500
```

## ⚠️ 注意

- **delete 是真删除**（IMAP `\Deleted` + `EXPUNGE`），不可恢复。先 `--dry-run`。
- **Gmail**：直接 delete 会从所有视图消失；想走回收站请 `move` 到 `[Gmail]/Trash`。
- UID 跨会话稳定，但邮箱被重建（UIDVALIDITY 变化）后需重新 `list`。

## 开发

```sh
go test ./...     # 集成测试基于 go-imap 自带的 imapmemserver，无需真实邮箱
go build -o imap-cli .
```
