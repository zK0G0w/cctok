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

// modelSynthetic 是 Claude Code 内部生成的占位记录，usage 全为 0，需过滤
const modelSynthetic = "<synthetic>"

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

// ParseAll 发现并解析所有 JSONL 文件（Claude Code + Codex）
func ParseAll(claudeDir string) ([]Record, error) {
	files, err := DiscoverFiles(claudeDir)
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(claudeDir, "projects")
	projectNameCache := make(map[string]string)

	type fileResult struct {
		records []Record
		cwd     string
	}
	resultCache := make(map[string]fileResult)

	// 单次遍历：解析文件并收集项目名（同一项目目录下取第一个有效 cwd 作为项目名来源）
	for _, f := range files {
		records, cwd, err := parseFileWithCwd(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: 跳过文件 %s: %v\n", f, err)
			continue
		}
		resultCache[f] = fileResult{records: records, cwd: cwd}

		dirName := extractProjectDir(f, projectsDir)
		if _, ok := projectNameCache[dirName]; !ok && cwd != "" {
			projectNameCache[dirName] = ExtractProjectName(cwd)
		}
	}

	// 利用缓存结果分配项目名（子代理文件可能无 cwd，回退到目录名）
	var all []Record
	for _, f := range files {
		res, ok := resultCache[f]
		if !ok {
			continue
		}
		dirName := extractProjectDir(f, projectsDir)
		projectName := projectNameCache[dirName]
		if projectName == "" {
			projectName = dirName
		}
		for i := range res.records {
			res.records[i].Project = projectName
		}
		all = append(all, res.records...)
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
	// 部分行包含完整 message content，需要较大 buffer
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

		if raw.Message.Model == modelSynthetic {
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

// ExtractProjectName 从 cwd 路径中提取最后两段作为项目名
// 例：/Users/dev/workspace/backend/payment → "backend/payment"
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
