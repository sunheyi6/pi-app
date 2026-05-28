package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pi-desktop/backend/piagent"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	client    *piagent.Client
	sessions  *piagent.SessionManager
	mu        sync.Mutex
	curCwd    string
	events    map[string]func(data any) // 前端事件回调
}

func NewApp() *App {
	return &App{
		events: make(map[string]func(data any)),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 获取用户主目录作为默认工作目录
	homeDir, _ := os.UserHomeDir()
	a.curCwd = homeDir

	a.sessions = piagent.NewSessionManager()
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.client != nil {
		a.client.Stop()
	}
}

// StartAgent 启动 pi agent RPC 子进程
func (a *App) StartAgent(cwd string, sessionPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 停止旧的 agent
	if a.client != nil {
		a.client.Stop()
	}

	if cwd != "" {
		a.curCwd = cwd
	}

	client, err := piagent.NewClient(a.curCwd, sessionPath)
	if err != nil {
		return fmt.Errorf("启动 agent 失败: %w", err)
	}
	a.client = client

	// 启动事件监听
	go a.listenEvents(client)

	return nil
}

// listenEvents 监听 pi 事件并转发到前端
func (a *App) listenEvents(client *piagent.Client) {
	for event := range client.Events() {
		// 将事件转发到前端
		eventData, _ := json.Marshal(event)
		runtime.EventsEmit(a.ctx, "pi-event", string(eventData))
	}
}

// StopAgent 停止 agent
func (a *App) StopAgent() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.client != nil {
		a.client.Stop()
		a.client = nil
	}
}

// SendPrompt 发送用户提示
func (a *App) SendPrompt(message string, images []piagent.ImageData) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	// 转换为 JSON 字符串
	imgJSON := "[]"
	if len(images) > 0 {
		b, _ := json.Marshal(images)
		imgJSON = string(b)
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:    "prompt",
		Message: message,
		Images:  json.RawMessage(imgJSON),
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// EnsureSessionNamed 在 agent 完成后设置会话标题（此时文件已存在）
func (a *App) EnsureSessionNamed(firstMessage string) string {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		log.Printf("[EnsureSessionNamed] client 为空")
		return ""
	}

	// 获取当前状态
	stateResp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_state",
	})
	if err != nil {
		log.Printf("[EnsureSessionNamed] get_state 失败: %v", err)
		return ""
	}

	log.Printf("[EnsureSessionNamed] get_state 原始响应: %s", string(stateResp))

	// 使用 map 解析以兼容多种可能的字段名
	var rawMap map[string]any
	if err := json.Unmarshal(stateResp, &rawMap); err != nil {
		log.Printf("[EnsureSessionNamed] 解析 state 失败: %v", err)
		return ""
	}

	// 尝试多种可能的字段名获取 sessionName
	sessionName := getStringFromMap(rawMap, "sessionName", "session_name", "name", "title")
	sessionFile := getStringFromMap(rawMap, "sessionFile", "session_file", "file", "path", "sessionPath")

	log.Printf("[EnsureSessionNamed] 解析结果 sessionName=%q sessionFile=%q", sessionName, sessionFile)

	// 已有标题则跳过（但如果是默认文件名格式，允许覆盖）
	if sessionName != "" && !looksLikeDefaultName(sessionName) {
		log.Printf("[EnsureSessionNamed] 已有标题 '%s'，跳过", sessionName)
		return sessionName
	}

	// 生成标题
	name := generateSessionName(firstMessage)
	if name == "" {
		log.Printf("[EnsureSessionNamed] 生成标题为空")
		return ""
	}

	// 如果 sessionFile 为空，尝试从会话列表中找到最新的匹配文件
	targetFile := sessionFile
	if targetFile == "" {
		sessions, _ := a.sessions.ListSessions(a.curCwd)
		if len(sessions) > 0 {
			// 取最新会话的文件（通常是当前会话）
			targetFile = sessions[0].FilePath
			log.Printf("[EnsureSessionNamed] sessionFile 为空，回退到最新会话文件: %s", targetFile)
		}
	}

	if targetFile == "" {
		log.Printf("[EnsureSessionNamed] 无法确定会话文件，跳过")
		return ""
	}

	log.Printf("[EnsureSessionNamed] 设置标题 '%s' -> %s", name, targetFile)

	// 方式 1：通过 RPC set_session_name（保持与 pi 内存状态一致）
	if _, err := client.SendCommand(piagent.RPCCommand{
		Type: "set_session_name",
		Name: name,
	}); err != nil {
		log.Printf("[EnsureSessionNamed] set_session_name RPC 失败: %v", err)
	}

	// 方式 2：直接修改 JSONL 文件（确保前端侧边栏可读取）
	if err := a.sessions.UpdateDisplayName(targetFile, name); err != nil {
		log.Printf("[EnsureSessionNamed] 修改文件失败: %v", err)
		return ""
	}

	log.Printf("[EnsureSessionNamed] 成功: %s", name)
	return name
}

// getStringFromMap 从 map 中按多个候选键获取字符串值
func getStringFromMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch val := v.(type) {
			case string:
				return val
			case []byte:
				return string(val)
			default:
				return fmt.Sprintf("%v", val)
			}
		}
	}
	return ""
}

// looksLikeDefaultName 判断名称是否像系统自动生成的默认名（允许覆盖）
func looksLikeDefaultName(name string) bool {
	// 空值、纯时间戳、纯 UUID、session- 前缀等视为默认名称
	if name == "" {
		return true
	}
	// 如果以 session- 开头，可能是默认文件名
	if strings.HasPrefix(strings.ToLower(name), "session-") {
		return true
	}
	// 如果名称全是数字和横线（类似时间戳）
	if strings.Trim(name, "0123456789-") == "" {
		return true
	}
	return false
}

// generateSessionName 从消息内容生成会话标题
func generateSessionName(message string) string {
	s := strings.TrimSpace(message)
	// 只取第一行
	if idx := strings.IndexAny(s, "\n\r"); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// 按字符截取（最多 30 个字符）
	runes := []rune(s)
	maxLen := 30
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

// SendSteer 发送引导消息
func (a *App) SendSteer(message string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:    "steer",
		Message: message,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// SendFollowUp 发送跟进消息
func (a *App) SendFollowUp(message string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:    "follow_up",
		Message: message,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// Abort 中止当前操作
func (a *App) Abort() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "abort",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// SetModel 切换模型
func (a *App) SetModel(provider string, modelId string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:     "set_model",
		Provider: provider,
		ModelId:  modelId,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// SetThinkingLevel 设置思考级别
func (a *App) SetThinkingLevel(level string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:  "set_thinking_level",
		Level: level,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GetState 获取当前状态
func (a *App) GetState() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_state",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GetMessages 获取所有消息
func (a *App) GetMessages() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_messages",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// NewSession 创建新会话
func (a *App) NewSession() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "new_session",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GetSessions 获取会话列表
func (a *App) GetSessions() string {
	sessions, err := a.sessions.ListSessions(a.curCwd)
	if err != nil {
		log.Printf("获取会话列表失败: %v", err)
		return "[]"
	}
	b, _ := json.Marshal(sessions)
	return string(b)
}

// GetForks 获取可复用的分支消息
func (a *App) GetForks() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_fork_messages",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// Fork 从指定分支创建新会话
func (a *App) Fork(entryId string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:    "fork",
		EntryId: entryId,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// ExportHTML 导出会话为 HTML
func (a *App) ExportHTML(outputPath string) (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type:       "export_html",
		OutputPath: outputPath,
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GetAvailableModels 获取可用模型列表
func (a *App) GetAvailableModels() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_available_models",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// GetCommands 获取可用命令
func (a *App) GetCommands() (string, error) {
	a.mu.Lock()
	client := a.client
	a.mu.Unlock()

	if client == nil {
		return "{}", fmt.Errorf("agent 未启动")
	}

	resp, err := client.SendCommand(piagent.RPCCommand{
		Type: "get_commands",
	})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// SelectDirectory 打开目录选择对话框
func (a *App) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择工作目录",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// GetAppInfo 获取应用信息
func (a *App) GetAppInfo() string {
	info := map[string]string{
		"os":      goruntime.GOOS,
		"arch":    goruntime.GOARCH,
		"cwd":     a.curCwd,
		"homeDir": "",
	}
	homeDir, _ := os.UserHomeDir()
	info["homeDir"] = homeDir

	b, _ := json.Marshal(info)
	return string(b)
}

// 确保路径存在
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
