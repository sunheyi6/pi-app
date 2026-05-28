# Pi Desktop

基于 [pi coding agent](https://github.com/earendil-works/pi) 的桌面 AI 编程助手，使用 Wails (Go + Vue) 构建。

## 前置条件

1. **Node.js** >= 18
2. **Go** >= 1.22
3. **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
4. **pi coding agent**: `npm install -g @earendil-works/pi-coding-agent`

## 开发

```bash
cd frontend
npm install

cd ..
wails dev
```

Windows（默认：可见终端前台启动，且仅保留一个 Wails 实例）：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\restart-wails.ps1
```

## 构建

```bash
wails build
```

## 项目结构

```
pi-app/
├── main.go                    # Wails 入口
├── app.go                     # 应用逻辑，Go 后端 API
├── go.mod
├── wails.json                 # Wails 配置
├── backend/
│   └── piagent/
│       ├── client.go          # pi RPC 客户端（子进程管理 + JSONL 通信）
│       ├── types.go           # Go 类型定义（RPC 协议）
│       └── session.go         # 会话发现与管理
└── frontend/
    └── src/
        ├── App.vue            # 主布局（Sidebar + ChatArea）
        ├── composables/
        │   └── usePiAgent.ts  # pi agent 通信 hook
        ├── stores/
        │   └── chatStore.ts   # Pinia 状态管理
        ├── components/
        │   ├── Sidebar.vue    # 左侧会话列表
        │   ├── ChatArea.vue   # 右侧聊天区域
        │   ├── ChatMessage.vue # 聊天消息气泡
        │   ├── InputBox.vue   # 底部输入框
        │   ├── ThinkingBlock.vue # AI 思考块
        │   ├── ToolCallCard.vue  # 工具调用卡片
        │   └── ModelSelector.vue # 模型/思考级别选择
        └── types/
            └── index.ts       # TypeScript 类型
```

## 架构

```
┌────────────────────────────────────────┐
│           Wails Desktop App             │
│                                        │
│  ┌──────────┐     ┌────────────────┐   │
│  │ Vue 前端  │◄───►│   Go 后端      │   │
│  │ (聊天UI) │绑定 │ (RPC Client)   │   │
│  └──────────┘     └───────┬────────┘   │
│                           │            │
│                  stdin/stdout JSONL    │
│                           │            │
│                 ┌─────────▼────────┐   │
│                 │  pi --mode rpc   │   │
│                 │  (子进程)        │   │
│                 └──────────────────┘   │
└────────────────────────────────────────┘
```

## 功能

- [x] 多会话管理（左侧会话列表）
- [x] 聊天式对话（用户/AI 气泡）
- [x] 流式文本输出
- [x] AI 思考过程展示
- [x] 工具调用详情（读文件、执行命令等）
- [x] 模型切换
- [x] 思考级别调整
- [x] 引导消息 / 跟进消息
- [x] 中止当前操作
- [x] Token 用量显示
- [x] 会话分叉 / 克隆
- [x] Markdown 渲染
