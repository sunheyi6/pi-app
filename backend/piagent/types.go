package piagent

import "encoding/json"

// RPCCommand 发送给 pi 的命令
type RPCCommand struct {
	ID                 string          `json:"id,omitempty"`
	Type               string          `json:"type"`
	Message            string          `json:"message,omitempty"`
	Images             json.RawMessage `json:"images,omitempty"`
	StreamingBehavior  string          `json:"streamingBehavior,omitempty"`
	Provider           string          `json:"provider,omitempty"`
	ModelId            string          `json:"modelId,omitempty"`
	Level              string          `json:"level,omitempty"`
	Mode               string          `json:"mode,omitempty"`
	EntryId            string          `json:"entryId,omitempty"`
	SessionPath        string          `json:"sessionPath,omitempty"`
	OutputPath         string          `json:"outputPath,omitempty"`
	Name               string          `json:"name,omitempty"`
	Enabled            *bool           `json:"enabled,omitempty"`
	CustomInstructions string          `json:"customInstructions,omitempty"`
	Value              string          `json:"value,omitempty"`
	Confirmed          *bool           `json:"confirmed,omitempty"`
	Cancelled          bool            `json:"cancelled,omitempty"`
}

// RPCResponse pi 的响应
type RPCResponse struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Command string          `json:"command,omitempty"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// RPCEvent pi 发出的事件
type RPCEvent struct {
	ID                    string                 `json:"id,omitempty"`
	Type                  string                 `json:"type"`
	Method                string                 `json:"method,omitempty"`
	Title                 string                 `json:"title,omitempty"`
	Message               json.RawMessage        `json:"message,omitempty"`
	Options               []string               `json:"options,omitempty"`
	Timeout               int                    `json:"timeout,omitempty"`
	Placeholder           string                 `json:"placeholder,omitempty"`
	Prefill               string                 `json:"prefill,omitempty"`
	NotifyType            string                 `json:"notifyType,omitempty"`
	Messages              json.RawMessage        `json:"messages,omitempty"`
	AssistantMessageEvent *AssistantMessageEvent `json:"assistantMessageEvent,omitempty"`
	ToolCallId            string                 `json:"toolCallId,omitempty"`
	ToolName              string                 `json:"toolName,omitempty"`
	Args                  json.RawMessage        `json:"args,omitempty"`
	Result                *ToolResult            `json:"result,omitempty"`
	IsError               bool                   `json:"isError,omitempty"`
	PartialResult         *ToolResult            `json:"partialResult,omitempty"`
	Steering              []string               `json:"steering,omitempty"`
	FollowUp              []string               `json:"followUp,omitempty"`
	Reason                string                 `json:"reason,omitempty"`
	Attempt               int                    `json:"attempt,omitempty"`
	MaxAttempts           int                    `json:"maxAttempts,omitempty"`
	DelayMs               int                    `json:"delayMs,omitempty"`
	ErrorMessage          string                 `json:"errorMessage,omitempty"`
	Aborted               bool                   `json:"aborted,omitempty"`
	WillRetry             bool                   `json:"willRetry,omitempty"`
	Success               bool                   `json:"success,omitempty"`
	FinalError            string                 `json:"finalError,omitempty"`
	ExtensionPath         string                 `json:"extensionPath,omitempty"`
	Event                 string                 `json:"event,omitempty"`
}

// AssistantMessageEvent streaming delta 事件
type AssistantMessageEvent struct {
	Type         string          `json:"type"`
	ContentIndex int             `json:"contentIndex,omitempty"`
	Delta        string          `json:"delta,omitempty"`
	Content      string          `json:"content,omitempty"`
	Partial      json.RawMessage `json:"partial,omitempty"`
	ToolCall     *ToolCall       `json:"toolCall,omitempty"`
	Reason       string          `json:"reason,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	Details json.RawMessage     `json:"details"`
}

// ToolResultContent 工具结果内容
type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ImageData 图片数据
type ImageData struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	FilePath     string `json:"filePath"`
	SessionId    string `json:"sessionId"`
	DisplayName  string `json:"displayName"`
	MessageCount int    `json:"messageCount"`
	LastModified string `json:"lastModified"`
}

// AppInfo 应用信息
type AppInfo struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Cwd     string `json:"cwd"`
	HomeDir string `json:"homeDir"`
}
