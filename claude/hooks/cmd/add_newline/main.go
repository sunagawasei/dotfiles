// add_newline は Claude Code のポストフックとして動作し、
// ファイル末尾に改行がない場合に自動的に追加します。
package main

import (
	"encoding/json"
	"io"
	"os"
)

// ToolInput はClaude Codeから渡されるツール入力を表します
type ToolInput struct {
	FilePath     string `json:"file_path"`
	NotebookPath string `json:"notebook_path"`
}

// InputData はフック入力の全体構造を表します
type InputData struct {
	ToolInput ToolInput `json:"tool_input"`
}

// addNewlineIfNeeded はファイル末尾に改行がなければ追加します
func addNewlineIfNeeded(filePath string) error {
	if filePath == "" {
		return nil
	}

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return nil // ファイルが存在しない、ディレクトリ、または空の場合はスキップ
	}

	file, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// 最後の1バイトを確認
	if _, err := file.Seek(-1, io.SeekEnd); err != nil {
		return err
	}

	buf := make([]byte, 1)
	if _, err := file.Read(buf); err != nil {
		return err
	}

	// 改行でない場合は追加
	if buf[0] != '\n' {
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			return err
		}
		if _, err := file.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var input InputData
	decoder := json.NewDecoder(os.Stdin)

	// JSONデコードに失敗しても静かに終了（non-blocking hook）
	if err := decoder.Decode(&input); err != nil {
		return
	}

	// ファイルパスを取得して処理
	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolInput.NotebookPath
	}

	// エラーがあっても静かに終了（non-blocking hook）
	_ = addNewlineIfNeeded(filePath)
}
