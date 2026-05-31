# 项目规范与记录

## 语言偏好
- **默认使用中文进行回复和思考**，包括思考过程、工具调用说明、代码注释和对话。
- 仅在代码、命令、文件名等必须使用英文的场景使用英文。

## "记一下" 指令
表示用户要求将当前讨论的结论、规范、要点或经验记录下来。
你必须使用 `edit` 或 `write` 工具将内容追加/写入本文件。

## 对话风格
- 回答简洁直接，不加无意义的客套话或表情符号。
- 用户提问时，**先回答问题本身**，再进行代码修改或执行命令。
- 回应反馈时，**先明确表示同意或不同意**，再说明做了什么变更。
- 做大范围改动前，必须**先完整阅读相关文件**，不能仅凭搜索片段就动手修改。

## Git 操作
- 可能存在多个 pi session 同时修改此仓库。Git 操作必须严格限定在本次 session 的修改范围内。
- 提交时使用显式路径：`git add <file1> <file2>`，**严禁** `git add -A` / `git add .`。
- **禁止**运行以下破坏性命令：`git reset --hard`、`git checkout .`、`git clean -fd`、`git stash`。
- 提交前运行 `git status` 确认只暂存了本次 session 自己的文件。

---

## 记录格式

每条新记录在文件末尾追加，格式如下：

```
### YYYY-MM-DD - 记录主题
- 要点 1
- 要点 2
```

---

### 2026-05-31 - Shell 环境约定
- pi 的 bash 工具实际运行在 **Git Bash（MSYS2）** 环境，`&&`、`||`、重定向等 bash 语法均可用。
- `start /b` 等 cmd 内置命令在 Git Bash 中不可用。
- Git Bash 路径：`D:\soft\Git\bin\bash.exe`。

### 2026-05-31 - Wails 项目启动与重启
- **单实例规则**：重启时先关闭旧进程，再启动新实例。
- **启动脚本**：`scripts/restart-wails.ps1`（重构版，含 `-StartOnly` 开关；HTTP 校验；按进程名+端口双重关闭旧进程）。
- **启动项目**：`powershell -ExecutionPolicy Bypass -File .\scripts\restart-wails.ps1 -StartOnly`
- **重启项目**：`powershell -ExecutionPolicy Bypass -File .\scripts\restart-wails.ps1`
- **运行态校验以端口 HTTP 200 为准**（`curl` 127.0.0.1:5173 与 localhost:34115），比进程检查更可靠。
- **完成标准**：启动后必须做运行态校验，校验通过后才能反馈完成，不能只以"命令已执行"作为完成。
- 启动失败时检查日志：`wails-dev.log` 和 `wails-dev.err.log`。

### 2026-05-29 - 消息输入队列
- 需求：AI 回答期间用户输入不丢失，自动排队，回答结束后依次发送。
- 修改文件：
  - `frontend/src/stores/chatStore.ts`：新增 `inputQueue` 状态、`enqueueInput`/`dequeueInput`/`clearInputQueue` 方法。
  - `frontend/src/components/InputBox.vue`：AI 繁忙时（`isStreaming`/`isAgentRunning`），发送不直接 emit，而是 `store.enqueueInput()`；发送按钮始终显示（不再被中止按钮替代）；输入框上方渲染队列列表。
  - `frontend/src/composables/usePiAgent.ts`：`agent_end` 和 `turn_end` 事件中调用 `processInputQueue()`，自动从队列取出并发送下一条。
  - `frontend/src/App.vue`：中止时调用 `store.clearInputQueue()` 清空队列。

### 2026-05-28 - 联网搜索配置
- 本机已安装 pi 包：`git:github.com/sunheyi6/sun`，包含两个扩展：
  - `message-queue.ts`：消息队列
  - `opencli-browser.ts`：浏览器集成（web_navigate、web_extract、web_search 等 11 个工具）
- 已修改 `opencli-browser.ts` 的 `web_search` 工具：**默认搜索引擎从百度改为谷歌优先**。
  - 优先使用 Google：直接通过 URL `https://www.google.com/search?q=...&hl=zh-CN` 搜索
  - 自动降级：如果 Google 被拦截（CAPTCHA / consent / 无结果），自动回退到百度搜索
- 项目本地 skill：`.pi/skills/web-search/SKILL.md` — 联网搜索指导规范（谷歌优先策略）
