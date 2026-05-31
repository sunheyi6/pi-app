package piagent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Client 管理 pi RPC 子进程的通信
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	events   chan RPCEvent
	mu       sync.Mutex
	done     chan struct{}
	running  bool
	respChs  map[string]chan *RPCResponse
	reqCount int
}

// NewClient 创建新的 pi RPC 客户端，启动子进程
func NewClient(cwd string, sessionPath string) (*Client, error) {
	args := []string{"--mode", "rpc"}
	if sessionPath != "" {
		args = append(args, "--session", sessionPath)
	}

	piExec, err := resolvePiExecutable()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(piExec, args...)
	cmd.Dir = cwd
	cmd.Env = withPnpmPath(os.Environ())

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdin 管道失败: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdout 管道失败: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stderr 管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 pi 进程失败: %w", err)
	}

	client := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		events:  make(chan RPCEvent, 100),
		done:    make(chan struct{}),
		running: true,
		respChs: make(map[string]chan *RPCResponse),
	}

	go client.readEvents()
	go client.readStderr()

	return client, nil
}

func resolvePiExecutable() (string, error) {
	// Prefer PATH lookup first.
	for _, name := range []string{"pi", "pi.cmd", "pi.CMD"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	// Fallback for Windows + pnpm global install location.
	homeDir, err := os.UserHomeDir()
	if err == nil {
		for _, p := range piExecutableCandidates(homeDir) {
			if _, statErr := os.Stat(p); statErr == nil {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("未找到 pi 可执行文件，请确认已安装: npm install -g @earendil-works/pi-coding-agent")
}

func withPnpmPath(env []string) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return env
	}
	pnpmDir := filepath.Join(homeDir, "AppData", "Local", "pnpm")
	if _, statErr := os.Stat(pnpmDir); statErr != nil {
		return env
	}

	pathVal := os.Getenv("PATH")
	newPath := pnpmDir
	if pathVal != "" {
		newPath = pnpmDir + string(os.PathListSeparator) + pathVal
	}

	replaced := false
	for i, kv := range env {
		if len(kv) >= 5 && (kv[:5] == "PATH=" || kv[:5] == "Path=") {
			env[i] = "PATH=" + newPath
			replaced = true
			break
		}
	}
	if !replaced {
		env = append(env, "PATH="+newPath)
	}
	return env
}

// readEvents 从 stdout 读取 JSONL 事件
func (c *Client) readEvents() {
	defer close(c.events)
	scanner := bufio.NewScanner(c.stdout)
	// 增大 buffer 以处理大型 JSON 行
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		// 尝试解析为响应
		var resp RPCResponse
		if err := json.Unmarshal(line, &resp); err == nil && resp.Type == "response" {
			c.mu.Lock()
			if ch, ok := c.respChs[resp.ID]; ok {
				ch <- &resp
			}
			c.mu.Unlock()
			continue
		}

		// 尝试解析为事件
		var event RPCEvent
		if err := json.Unmarshal(line, &event); err == nil {
			select {
			case c.events <- event:
			default:
				log.Printf("[piagent] 事件通道已满，丢弃事件: %s", event.Type)
			}
			continue
		}

		// 未知格式，记录日志
		log.Printf("[piagent] 无法解析输出: %s", string(line[:min(len(line), 200)]))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[piagent] stdout 读取错误: %v", err)
	}
}

// readStderr 读取 stderr（用于调试）
func (c *Client) readStderr() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		log.Printf("[pi stderr] %s", scanner.Text())
	}
}

// SendCommand 发送命令并等待响应
func (c *Client) SendCommand(cmd RPCCommand) (json.RawMessage, error) {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil, fmt.Errorf("client 已关闭")
	}

	c.reqCount++
	reqID := fmt.Sprintf("req-%d", c.reqCount)
	cmd.ID = reqID

	respCh := make(chan *RPCResponse, 1)
	c.respChs[reqID] = respCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.respChs, reqID)
		c.mu.Unlock()
	}()

	// 序列化并发送命令
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("序列化命令失败: %w", err)
	}
	data = append(data, '\n')

	c.mu.Lock()
	_, err = c.stdin.Write(data)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("写入命令失败: %w", err)
	}

	// 等待响应
	resp := <-respCh
	if resp == nil {
		return nil, fmt.Errorf("未收到响应")
	}

	if !resp.Success {
		return nil, fmt.Errorf("命令失败: %s", resp.Error)
	}

	return resp.Data, nil
}

// SendCommandAsync 发送命令（不等待响应）
func (c *Client) SendCommandAsync(cmd RPCCommand) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("client 已关闭")
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("序列化命令失败: %w", err)
	}
	data = append(data, '\n')

	_, err = c.stdin.Write(data)
	return err
}

// Events 返回事件通道
func (c *Client) Events() <-chan RPCEvent {
	return c.events
}

// Stop 停止 pi 子进程
func (c *Client) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}
	c.running = false

	// 关闭 stdin 通知进程退出
	_ = c.stdin.Close()

	// 杀死进程（如果还活着）
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	close(c.done)
}

// IsRunning 检查是否在运行
func (c *Client) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
