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

// codexRawRecord Codex JSONL 单行记录
type codexRawRecord struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type codexSessionMeta struct {
	ID  string `json:"id"`
	Cwd string `json:"cwd"`
}

type codexTurnContext struct {
	TurnID string `json:"turn_id"`
	Model  string `json:"model"`
}

type codexEventMsg struct {
	Type string          `json:"type"`
	Info *codexTokenInfo `json:"info"`
}

type codexTokenInfo struct {
	TotalTokenUsage *codexTokenUsage `json:"total_token_usage"`
	LastTokenUsage  *codexTokenUsage `json:"last_token_usage"`
}

type codexTokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

// DiscoverCodexFiles 遍历 Codex sessions 目录，返回所有 .jsonl 文件路径
func DiscoverCodexFiles() ([]string, error) {
	homeDir, _ := os.UserHomeDir()
	sessionsDir := filepath.Join(homeDir, ".codex", "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var files []string
	err := filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
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

// ParseCodexFile 解析单个 Codex session JSONL 文件
// 返回每个 turn 的增量 token 用量记录
func ParseCodexFile(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		sessionID string
		cwd       string
		model     string
		lastTotal *codexTokenUsage
		records   []Record
		lastTS    time.Time
	)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var raw codexRawRecord
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, raw.Timestamp)

		switch raw.Type {
		case "session_meta":
			var meta codexSessionMeta
			if json.Unmarshal(raw.Payload, &meta) == nil {
				sessionID = meta.ID
				cwd = meta.Cwd
			}

		case "turn_context":
			var tc codexTurnContext
			if json.Unmarshal(raw.Payload, &tc) == nil {
				if tc.Model != "" {
					model = tc.Model
				}
			}

		case "event_msg":
			var evt codexEventMsg
			if json.Unmarshal(raw.Payload, &evt) == nil && evt.Type == "token_count" {
				if evt.Info != nil && evt.Info.TotalTokenUsage != nil {
					lastTotal = evt.Info.TotalTokenUsage
					lastTS = ts
				}
			}
		}
	}

	// 用 total_token_usage 作为整个会话的总用量
	if lastTotal != nil {
		usage := Usage{
			InputTokens:              lastTotal.InputTokens,
			OutputTokens:             lastTotal.OutputTokens + lastTotal.ReasoningOutputTokens,
			CacheCreationInputTokens: 0,
			CacheReadInputTokens:     lastTotal.CachedInputTokens,
		}

		project := ExtractProjectName(cwd)
		if model == "" {
			model = "codex-unknown"
		}

		records = append(records, Record{
			MessageID: fmt.Sprintf("codex-%s", sessionID),
			Model:     model,
			Usage:     usage,
			Timestamp: lastTS,
			SessionID: sessionID,
			Project:   project,
			Source:    "Codex",
		})
	}

	return records, nil
}

// ParseAllCodex 发现并解析所有 Codex session 文件
func ParseAllCodex() ([]Record, error) {
	files, err := DiscoverCodexFiles()
	if err != nil {
		return nil, err
	}

	var all []Record
	for _, f := range files {
		records, err := ParseCodexFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: 跳过 Codex 文件 %s: %v\n", f, err)
			continue
		}
		all = append(all, records...)
	}
	return all, nil
}
