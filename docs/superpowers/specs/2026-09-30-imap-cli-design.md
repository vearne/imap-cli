# imap-cli 设计文档

日期：2026-09-30
状态：已确认（用户已批准设计方向）

## 1. 目标

为 `emersion/go-imap/v2` 封装一个命令行工具 `imap-cli`，让大模型（agent）能通过 shell
调用它来处理 IMAP 邮箱中的邮件。**首要场景：批量删除旧邮件**；同时覆盖日常的
读取（folders/list/read）与整理（mark/move/delete）。

非目标（本版不做）：
- 发送邮件（SMTP 超出 IMAP 范围）
- 下载附件正文（`read` 仅列出附件元信息）
- MCP server、常驻 daemon、交互式 TUI
- 配置文件（凭据只从环境变量读取）

## 2. 技术栈

| 依赖 | 用途 |
|---|---|
| `github.com/emersion/go-imap/v2` (v2.0.0-beta.8) | IMAP 客户端协议实现 |
| `github.com/emersion/go-message` | MIME 解析（正文、附件） |
| `github.com/spf13/cobra` | CLI 框架（子命令、help、shell 补全） |
| `golang.org/x/text` | GBK 等非 UTF-8 邮件编码转码 |

## 3. 配置（仅环境变量）

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `IMAP_HOST` | 是 | — | 如 `imap.gmail.com` |
| `IMAP_PORT` | 否 | `993` | |
| `IMAP_USERNAME` | 是 | — | |
| `IMAP_PASSWORD` | 是 | — | Gmail 用应用专用密码；QQ/163 用授权码 |
| `IMAP_TLS` | 否 | `tls` | `tls` / `starttls` / `none` |

缺失必填变量时直接报错退出，并提示是哪个变量缺失。

## 4. 命令集

全局参数：`--json`（结构化输出，供 LLM 消费）、`--folder`（目标文件夹，默认 `INBOX`）。

```
imap-cli folders [--counts]
    列出所有文件夹。默认仅名称与分隔符；--counts 时对每个文件夹发 STATUS
    查询（UNSEEN/MESSAGES），较慢但信息全。

imap-cli list [--from S] [--subject S] [--since DATE] [--before DATE]
              [--unread] [--limit N]
    搜索并列出邮件。DATE 格式 YYYY-MM-DD，映射到 IMAP SENTSINCE/SENTBEFORE
    （按 Date 头）。输出字段：uid, from, subject, date, seen, flagged, size。
    按日期倒序（客户端排序），--limit 默认 50。

imap-cli read <uid> [--raw]
    读取单封邮件：关键头（From/To/Cc/Date/Subject）+ 正文 + 附件清单（文件名+大小）。
    正文优先取 text/plain；无 text/plain 时输出 text/html 原文并标注 [html]。
    --raw 输出原始 RFC822 报文。

imap-cli mark <uid-set> read|unread|flagged|unflagged
    修改标记，映射到 \Seen / \Flagged 的增删。

imap-cli move <uid-set> <dest-folder>
    移动邮件。优先用 MOVE（RFC 6851）；服务器不支持时降级为 COPY + \Deleted + EXPUNGE。

imap-cli delete <uid-set> [--dry-run]
    删除邮件：\Deleted + EXPUNGE（标准 IMAP 语义，不可恢复）。
    --dry-run 只输出将删除的数量与 UID 范围，不做任何修改。
```

### UID 集合语法

所有针对邮件的命令接受 UID 集合：`5`、`1:20000`、`1:500,800,1000:1200`。
内部用 `imap.UIDSet` 一次提交，不逐封循环。

## 5. 关键设计决策

1. **无状态**：每次调用 建连 → 登录 → 操作 → Logout。简单、可组合，LLM 友好。
2. **UID 定位**：跨会话稳定，LLM 可以先 `list` 拿 UID 再 `delete`。若操作中检测到
   UIDVALIDITY 变化，报错终止并提示重新 list。
3. **输出双模式**：默认人读（对齐文本），`--json` 输出结构化 JSON。错误信息走 stderr，
   退出码非 0。
4. **删除安全**：`delete` 必须显式给出 UID 集合（来自先前 `list` 的结果），提供
   `--dry-run` 预演。审查动作发生在 LLM/用户层面（list → 确认 → delete）。
5. **编码**：mUTF-7 文件夹名与 UTF-8 的转换交给 go-imap v2 的内置处理；正文/头的
   非 UTF-8 charset（GBK 等）用 `x/text` 转码。

## 6. 已知注意事项（写入 README）

- **删除不可恢复**：EXPUNGE 后邮件消失。
- **Gmail 特殊语义**：直接 EXPUNGE 会让邮件从所有视图消失；想走"回收站"应
  `move` 到 `[Gmail]/Trash`。
- **QQ/163**：密码栏填授权码，不是网页登录密码；需在邮箱设置里开启 IMAP。

## 7. 错误处理

- 连接失败 / TLS 失败 / 认证失败：单行清晰错误 + 常见原因提示
  （授权码 vs 登录密码、IMAP 服务未开启、应用专用密码）。
- 认证失败时明确提示"已尝试 USER=<IMAP_USERNAME>"，避免排查错账号。
- 所有命令错误退出码为 1；--dry-run 永远退出 0。

## 8. 测试策略

- **集成测试**：使用 go-imap v2 自带的 `imapmemserver`（内存 IMAP 服务器），
  在 `go test` 中起真实协议栈覆盖全部命令，不 mock 协议层。
- **单元测试**：UID 集合参数解析、日期过滤映射、输出格式化（文本/JSON）、
  MIME 正文抽取（含 GBK 样本）。

## 9. 项目结构（预估）

```
imap-cli/
├── main.go              // cobra 根命令
├── internal/
│   ├── config/          // 环境变量读取与校验
│   ├── client/          // imapclient 封装：连接、登录、UID 操作
│   └── output/          // 文本 / JSON 双格式输出
├── cmd/                 // folders, list, read, mark, move, delete 子命令
└── README.md
```
