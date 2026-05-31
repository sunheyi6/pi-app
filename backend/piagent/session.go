package piagent

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// SessionManager 管理会话文件
type SessionManager struct {
	mu              sync.RWMutex
	cache           []SessionInfo
	cacheExpiry     time.Time
	cacheTTL        time.Duration
	msgCache        map[string][]json.RawMessage // sessionPath → messages
	msgCacheMax     int
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		cacheTTL:    3 * time.Second,
		msgCache:    make(map[string][]json.RawMessage),
		msgCacheMax: 10,
	}
}

// SessionHeaderInfo 会话头部信息
type SessionHeaderInfo struct {
	SessionID    string
	DisplayName  string
	Cwd          string
	MessageCount int
}

// InvalidateCache 使缓存失效（会话变更后调用）
func (sm *SessionManager) InvalidateCache() {
	sm.mu.Lock()
	sm.cacheExpiry = time.Time{}
	sm.mu.Unlock()
}

// UpsertSessionInCache 更新或插入单个会话到缓存（增量更新，避免全量扫描）
func (sm *SessionManager) UpsertSessionInCache(info SessionInfo) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, s := range sm.cache {
		if s.FilePath == info.FilePath {
			sm.cache[i] = info
			// 重新排序
			sort.Slice(sm.cache, func(i, j int) bool {
				return sm.cache[i].LastModified > sm.cache[j].LastModified
			})
			return
		}
	}
	// 新会话，插入并排序
	sm.cache = append(sm.cache, info)
	sort.Slice(sm.cache, func(i, j int) bool {
		return sm.cache[i].LastModified > sm.cache[j].LastModified
	})
}

// ListSessions 列出会话（带内存缓存，避免频繁扫描磁盘）
func (sm *SessionManager) ListSessions(cwd string) ([]SessionInfo, error) {
	sm.mu.RLock()
	if time.Now().Before(sm.cacheExpiry) && len(sm.cache) > 0 {
		result := sm.cache
		sm.mu.RUnlock()
		return result, nil
	}
	sm.mu.RUnlock()

	// 重建缓存
	sessions, err := sm.scanSessions(cwd)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.cache = sessions
	sm.cacheExpiry = time.Now().Add(sm.cacheTTL)
	sm.mu.Unlock()

	return sessions, nil
}

// scanSessions 实际扫描磁盘获取会话列表
func (sm *SessionManager) scanSessions(cwd string) ([]SessionInfo, error) {
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

// readSessionHeader 读取 JSONL 会话文件的头部信息（只读前 64KB，避免加载大文件）
func (sm *SessionManager) readSessionHeader(filePath string) (*SessionHeaderInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 只读取前 64KB 获取头部信息，避免加载整个大文件
	buf := make([]byte, 64*1024)
	n, _ := f.Read(buf)
	if n == 0 {
		return &SessionHeaderInfo{}, nil
	}
	data := string(buf[:n])

	info := &SessionHeaderInfo{}
	lines := strings.Split(data, "\n")
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
			if sid, ok := entry["id"].(string); ok {
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

	// 如果文件中途被截断，从文件大小估算剩余消息数
	if n == len(buf) {
		if fi, err := os.Stat(filePath); err == nil {
			ratio := float64(fi.Size()) / float64(n)
			messageCount = int(float64(messageCount) * ratio)
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

	// 更新缓存中的 displayName
	sm.mu.Lock()
	for i, s := range sm.cache {
		if s.FilePath == filePath {
			sm.cache[i].DisplayName = name
			break
		}
	}
	sm.mu.Unlock()

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

// PreloadSessions 并发预加载多个会话的 JSONL 消息到内存缓存
func (sm *SessionManager) PreloadSessions(sessionPaths []string) {
	var wg sync.WaitGroup
	for _, p := range sessionPaths {
		// 已缓存则跳过
		sm.mu.RLock()
		_, ok := sm.msgCache[p]
		sm.mu.RUnlock()
		if ok {
			continue
		}
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			msgs, err := sm.readSessionMessages(path)
			if err != nil {
				return
			}
			sm.mu.Lock()
			// 超过上限则清掉一半旧条目
			if len(sm.msgCache) >= sm.msgCacheMax {
				count := 0
				for k := range sm.msgCache {
					delete(sm.msgCache, k)
					count++
					if count >= sm.msgCacheMax/2 {
						break
					}
				}
			}
			sm.msgCache[path] = msgs
			sm.mu.Unlock()
		}(p)
	}
	wg.Wait()
}

// GetCachedMessages 获取缓存的消息（用于即时切换）
func (sm *SessionManager) GetCachedMessages(sessionPath string) ([]json.RawMessage, bool) {
	sm.mu.RLock()
	msgs, ok := sm.msgCache[sessionPath]
	sm.mu.RUnlock()
	return msgs, ok
}

// InvalidateMessageCache 清除指定会话的消息缓存
func (sm *SessionManager) InvalidateMessageCache(sessionPath string) {
	sm.mu.Lock()
	delete(sm.msgCache, sessionPath)
	sm.mu.Unlock()
}

// LoadMessagesFromFile 公开方法：直接读取 JSONL 文件并缓存
func (sm *SessionManager) LoadMessagesFromFile(filePath string) ([]json.RawMessage, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否已有缓存
	if cached, ok := sm.msgCache[filePath]; ok {
		return cached, nil
	}

	return sm.readSessionMessagesLocked(filePath)
}

// readSessionMessagesLocked 读取并缓存（需持有锁）
func (sm *SessionManager) readSessionMessagesLocked(filePath string) ([]json.RawMessage, error) {
	msgs, err := sm.readSessionMessages(filePath)
	if err != nil {
		return nil, err
	}
	// 缓存
	if len(sm.msgCache) >= sm.msgCacheMax {
		count := 0
		for k := range sm.msgCache {
			delete(sm.msgCache, k)
			count++
			if count >= sm.msgCacheMax/2 {
				break
			}
		}
	}
	sm.msgCache[filePath] = msgs
	return msgs, nil
}

// readSessionMessages 直接读取 JSONL 文件中的消息（不走 pi RPC）
func (sm *SessionManager) readSessionMessages(filePath string) ([]json.RawMessage, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var msgs []json.RawMessage
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 校验是否为合法 JSON
		if !json.Valid([]byte(line)) {
			continue
		}
		// 跳过 session 头部行
		var entry map[string]any
		json.Unmarshal([]byte(line), &entry)
		if entryType, _ := entry["type"].(string); entryType == "session" {
			continue
		}
		msgs = append(msgs, json.RawMessage(line))
	}
	return msgs, nil
}

// generateID 生成唯一 ID
func generateID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 8)
	r.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
