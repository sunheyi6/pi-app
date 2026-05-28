# 项目规范与记录

## 语言偏好
- **默认使用中文进行回复和思考**,包括思考过程、工具调用说明、代码注释和对话。
- 仅在代码、命令、文件名等必须使用英文的场景使用英文。

## "记一下" 指令,表示用户要求将当前讨论的结论、规范、要点或经验记录下来。
你必须使用 `edit` 或 `write` 工具将内容追加/写入本文件。

---

## 记录格式

每条新记录在文件末尾追加,格式如下:

```
### YYYY-MM-DD - 记录主题
- 要点 1
- 要点 2
```

---

### 2026-05-28 - Wails 启动与重启约定
- 启动命令:在项目根目录执行 `wails dev`。
- 若需要启动但已存在 `wails` 进程,必须先关闭当前 `wails`,再启动新的实例,始终只保留一个 `wails`。
- 新增重启脚本:`scripts/restart-wails.ps1`,执行方式:`powershell -ExecutionPolicy Bypass -File .\scripts\restart-wails.ps1`。
- 约定:后续让助手"重启项目"时,按上述单实例规则执行。

### 2026-05-28 - 默认启动方式
- 默认使用"可见终端前台启动"方式启动项目(在新终端窗口执行 `wails dev`),确保桌面 App 窗口可见。
- 若需要重启,仍遵循单实例规则:先关闭已有 `wails`,再以前台可见方式启动一个新实例。

### 2026-05-28 - 启动完成标准
- 当用户要求"启动项目/重启项目"时，完成标准是"启动成功"，不能只以"命令已执行"作为完成。
- 启动后必须做运行态校验（至少包含：`wails` 进程存在、前端 dev server 可用/监听），校验通过后才能反馈完成。

### 2026-05-28 - 实际可用的启动命令（cmd 环境）
- 当前环境 bash 工具实际运行在 Windows cmd 下（非 Git Bash / WSL），命令须用 cmd 语法，连接命令用 `&` 而非 `&&`。
- 关闭旧进程：先用 `tasklist | findstr /i wails` 检查，若有则 `taskkill /f /im wails.exe` 杀进程。
- 启动命令：`cd /d D:\product\pi-app & start /b wails dev > wails-dev.log 2>&1`
  - `start /b` 在当前窗口后台启动，不阻塞。
  - 日志重定向到 `wails-dev.log`，方便排查问题。
- 等待编译（约 15-20 秒），可循环 `ping -n N 127.0.0.1 >nul` 做延迟。
- 运行态校验：
  - 进程检查：`tasklist | findstr /i wails`
  - 前端端口检查：`curl -s -o nul -w "%{http_code}" http://127.0.0.1:5173` 期望返回 200
  - Wails 端口检查：`curl -s -o nul -w "%{http_code}" http://localhost:34115` 期望返回 200
  - 日志也可用 `type wails-dev.log` 查看启动详情。

### 2026-05-28 - 联网搜索配置
- 本机已安装 pi 包：`git:github.com/sunheyi6/sun`，包含两个扩展：
  - `message-queue.ts`：消息队列
  - `opencli-browser.ts`：浏览器集成（web_navigate、web_extract、web_search 等 11 个工具）
- 已修改 `opencli-browser.ts` 的 `web_search` 工具：**默认搜索引擎从百度改为谷歌优先**。
  - 优先使用 Google：直接通过 URL `https://www.google.com/search?q=...&hl=zh-CN` 搜索
  - 自动降级：如果 Google 被拦截（CAPTCHA / consent / 无结果），自动回退到百度搜索
- 项目本地 skill：`.pi/skills/web-search/SKILL.md` — 联网搜索指导规范（谷歌优先策略）
