// publish-claude-settings は ~/.config/claude/settings.json から、個人パスと社内識別子を
// 除いた公開用の claude/settings.public.json を生成する。
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// 配列の要素に現れたら、その要素ごと落とす語。社内の CLI・バケット名と、個人環境のパス
var dropTerms = regexp.MustCompile(`cycloud|containerregistry|/poc/|\bpup\b|\bgws\b`)

// 公開物に残ってはならない値。生成後の走査で1つでも見つかれば出力しない
type secrets struct {
	home string
	uid  string
	user string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "publish-claude-settings:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	in := filepath.Join(home, ".config/claude/settings.json")
	out := filepath.Join(home, ".config/claude/settings.public.json")
	if len(args) == 2 {
		in, out = args[0], args[1]
	} else if len(args) != 0 {
		return errors.New("usage: publish-claude-settings [input output]")
	}

	raw, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	s := secrets{home: home, uid: strconv.Itoa(os.Getuid()), user: filepath.Base(home)}
	pub, err := publicize(doc, s)
	if err != nil {
		return err
	}
	return os.WriteFile(out, pub, 0o644)
}

// publicize は公開用の JSON を返す。残留する個人値があればエラーにする
func publicize(doc map[string]any, s secrets) ([]byte, error) {
	// 社内バケット名・リポジトリ運用など、自然文に個人の事情が入る
	if auto, ok := doc["autoMode"].(map[string]any); ok {
		delete(auto, "environment")
	}
	cleaned := clean(doc, s)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cleaned); err != nil {
		return nil, err
	}
	for _, term := range []string{s.uid, s.user} {
		if strings.Contains(buf.String(), term) {
			return nil, fmt.Errorf("個人値 %q が公開版に残っている", term)
		}
	}
	if m := dropTerms.FindString(buf.String()); m != "" {
		return nil, fmt.Errorf("除外語 %q が公開版に残っている(配列の要素以外にある)", m)
	}
	return buf.Bytes(), nil
}

func clean(v any, s secrets) any {
	switch v := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, x := range v {
			out[k] = clean(x, s)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, x := range v {
			if str, ok := x.(string); ok && (dropTerms.MatchString(str) || strings.Contains(str, s.uid)) {
				continue
			}
			out = append(out, clean(x, s))
		}
		return out
	case string:
		// Read(//Users/...) のように先頭へスラッシュが付く形を先に置換する
		return strings.NewReplacer("/"+s.home, "~", s.home, "~").Replace(v)
	default:
		return v
	}
}
