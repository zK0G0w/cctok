package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Usage token 用量
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// Record 解析后的单条 assistant 记录
type Record struct {
	MessageID string
	Model     string
	Project   string
	SessionID string
	Source    string // "Claude Code" 或 "Codex"
	Usage     Usage
	Timestamp time.Time
}

type rawRecord struct {
	Type      string     `json:"type"`
	Message   rawMessage `json:"message"`
	Timestamp string     `json:"timestamp"`
	SessionID string     `json:"sessionId"`
	Cwd       string     `json:"cwd"`
}

type rawMessage struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Usage Usage  `json:"usage"`
}

// DiscoverFiles 遍历 Claude projects 目录，返回所有 .jsonl 文件路径
func DiscoverFiles(claudeDir string) ([]string, error) {
	projectsDir := filepath.Join(claudeDir, "projects")
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Claude Code 数据目录不存在: %s", projectsDir)
	}

	var files []string
	err := filepath.WalkDir(projectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ParseFile 解析单个 JSONL 文件，返回 assistant 类型的 Record 列表
func ParseFile(path string, projectName string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []Record
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw rawRecord
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		if raw.Type != "assistant" || raw.Message.ID == "" {
			continue
		}

		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			ts = time.Time{}
		}

		records = append(records, Record{
			MessageID: raw.Message.ID,
			Model:     raw.Message.Model,
			Usage:     raw.Message.Usage,
			Timestamp: ts,
			SessionID: raw.SessionID,
			Project:   projectName,
		})
	}

	return records, nil
}

// ParseAll 发现并解析所有 JSONL 文件（Claude Code + Codex）
func ParseAll(claudeDir string) ([]Record, error) {
	files, err := DiscoverFiles(claudeDir)
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(claudeDir, "projects")
	projectNameCache := make(map[string]string)

	// 第一遍：收集每个项目目录的可读名称（从有 cwd 的文件中获取）
	for _, f := range files {
		dirName := extractProjectDir(f, projectsDir)
		if _, ok := projectNameCache[dirName]; ok {
			continue
		}
		_, firstCwd, _ := parseFileWithCwd(f)
		if firstCwd != "" {
			projectNameCache[dirName] = ExtractProjectName(firstCwd)
		}
	}

	// 第二遍：解析所有文件并赋予项目名
	var all []Record
	for _, f := range files {
		dirName := extractProjectDir(f, projectsDir)
		projectName := projectNameCache[dirName]
		if projectName == "" {
			projectName = dirName
		}

		records, _, err := parseFileWithCwd(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: 跳过文件 %s: %v\n", f, err)
			continue
		}

		for i := range records {
			records[i].Project = projectName
		}
		all = append(all, records...)
	}

	// 合并 Codex 数据
	codexRecords, err := ParseAllCodex()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: 读取 Codex 数据失败: %v\n", err)
	} else {
		all = append(all, codexRecords...)
	}

	return all, nil
}

// extractProjectDir 从文件路径中提取项目目录名
func extractProjectDir(filePath, projectsDir string) string {
	rel, err := filepath.Rel(projectsDir, filePath)
	if err != nil {
		return "unknown"
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	if len(parts) == 0 {
		return "unknown"
	}
	return parts[0]
}

// parseFileWithCwd 解析 JSONL 文件，返回记录列表和第一条记录的 cwd
func parseFileWithCwd(path string) ([]Record, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	var records []Record
	var firstCwd string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw rawRecord
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		if raw.Type != "assistant" || raw.Message.ID == "" {
			continue
		}

		if firstCwd == "" && raw.Cwd != "" {
			firstCwd = raw.Cwd
		}

		ts, err := time.Parse(time.RFC3339Nano, raw.Timestamp)
		if err != nil {
			ts = time.Time{}
		}

		records = append(records, Record{
			MessageID: raw.Message.ID,
			Model:     raw.Message.Model,
			Usage:     raw.Message.Usage,
			Timestamp: ts,
			SessionID: raw.SessionID,
			Source:    "Claude Code",
		})
	}

	return records, firstCwd, nil
}

// ExtractProjectName 从 cwd 路径中提取最后两段作为项目名（保留兼容）
func ExtractProjectName(cwd string) string {
	if cwd == "" {
		return "unknown"
	}
	cleaned := filepath.Clean(cwd)
	parts := strings.Split(cleaned, string(filepath.Separator))
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "unknown"
}
