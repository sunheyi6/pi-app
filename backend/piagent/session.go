package piagent

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionManager 管理会话文件
type SessionManager struct{}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

// SessionHeaderInfo 会话头部信息
type SessionHeaderInfo struct {
	SessionID    string
	DisplayName  string
	Cwd          string
	MessageCount int
}

// ListSessions 列出会话（递归扫描子目录以支持 pi 的项目分组存储）
func (sm *SessionManager) ListSessions(cwd string) ([]SessionInfo, error) {
	homeDir, _ := os.UserHomeDir()

	// pi 会话存储目录
	sessionDir := filepath.Join(homeDir, ".pi", "agent", "sessions")

	var sessions []SessionInfo

	// 递归遍历所有子目录，查找 .jsonl 文件
	err := filepath.WalkDir(sessionDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过无法访问的目录
		}
		if d.IsDir() {
			return nil // 继续递归
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}

		// 读取会话头部信息
		info, err := sm.readSessionHeader(path)
		if err != nil {
			return nil
		}

		modTime := ""
		if fi, err := d.Info(); err == nil {
			modTime = fi.ModTime().Format("2006-01-02 15:04")
		}

		displayName := info.DisplayName
		if displayName == "" {
			displayName = strings.TrimSuffix(d.Name(), ".jsonl")
		}

		sessions = append(sessions, SessionInfo{
			FilePath:     path,
			SessionId:    info.SessionID,
			DisplayName:  displayName,
			MessageCount: info.MessageCount,
			LastModified: modTime,
		})
		return nil
	})

	if err != nil {
		// WalkDir 失败（目录不存在等），返回空列表
		return []SessionInfo{}, nil
	}

	// 按修改时间倒序排序
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastModified > sessions[j].LastModified
	})

	return sessions, nil
}

// readSessionHeader 读取 JSONL 会话文件的头部信息
func (sm *SessionManager) readSessionHeader(filePath string) (*SessionHeaderInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	info := &SessionHeaderInfo{}
	lines := strings.Split(string(data), "\n")
	messageCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		entryType, _ := entry["type"].(string)
		if entryType == "session" {
			if sid, ok := entry["sessionId"].(string); ok {
				info.SessionID = sid
			}
			if name, ok := entry["displayName"].(string); ok {
				info.DisplayName = name
			}
			if cwd, ok := entry["cwd"].(string); ok {
				info.Cwd = cwd
			}
		} else {
			messageCount++
		}
	}

	info.MessageCount = messageCount
	return info, nil
}

// UpdateDisplayName 更新 JSONL 文件中会话的 displayName
func (sm *SessionManager) UpdateDisplayName(filePath string, name string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取会话文件失败: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entryType, _ := entry["type"].(string); entryType == "session" {
			entry["displayName"] = name
			newLine, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("序列化 session 行失败: %w", err)
			}
			lines[i] = string(newLine)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到 session 头部行")
	}

	// 写回文件
	newData := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(newData), 0o644); err != nil {
		return fmt.Errorf("写入会话文件失败: %w", err)
	}

	return nil
}

// getSessionFileSize 获取会话文件大小
func (sm *SessionManager) getSessionFileSize(filePath string) string {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "0 KB"
	}
	size := fi.Size()
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

// generateID 生成唯一 ID
func generateID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 8)
	r.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
