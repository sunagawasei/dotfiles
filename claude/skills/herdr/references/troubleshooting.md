# herdr トラブルシューティング

`claude`をスポーン・委譲するときの実践知見。socket API自体の一般的な使い方（split/run/wait等）は`../SKILL.md`を参照。ここは「動いているようで動いていない」を避けるための落とし穴集。

## claude を別リポジトリに spawn して委譲する（実践知見 2026-07-21）

```bash
PANE=$(herdr workspace create --cwd /path/to/repo --no-focus | python3 -c 'import sys,json; print(json.load(sys.stdin)["result"]["root_pane"]["pane_id"])')
herdr pane run "$PANE" "claude"
sleep 5   # 起動待ち（wait output --match ">" は使えない。下記参照）
herdr pane run "$PANE" 'タスクプロンプト（1行。シングルクォートで囲み、$・バッククォート・シングルクォートを含めない）'
herdr pane send-keys "$PANE" Enter   # 長文はペースト扱いで pane run の Enter では submit されないため必須
```

- **長文プロンプトはペースト扱いになる**: 入力欄に `[Pasted text #1]` と表示されたまま止まり、`pane run` が送る Enter では submit されない。追いで `pane send-keys <pane> Enter` を送る。submit確認は `pane get` の `terminal_title` がタスク内容に変わったこと（既定の「Claude Code」のままなら未送信）
- **Claude Code の入力プロンプトは `>` ではなく `❯`**（fullscreen TUI）。`../SKILL.md`の「spawn a new agent」レシピの `--match ">"` は claude にはマッチせずタイムアウトする。起動・受理の確認は `pane read --source recent` で status line を見るか、`pane get` の `display_agent` / `terminal_title` がタスク内容になったことで行う
- **完了検知は agent_status ポーリングが確実**: `herdr pane get <pane>` の `.result.pane.agent_status` が `working` から `idle` / `blocked` / `done` に変わったら入力待ち。`wait agent-status` は単一 status しか待てないため、複数 status のいずれかを待つ場合は Claude Code の Monitor ツールに 10 秒間隔のポーリングループ（条件成立で exit）を渡して1通知で受ける
- herdr socket（`$XDG_CONFIG_HOME/herdr/herdr.sock`）は sandbox の `allowUnixSockets` に登録が必要。`Operation not permitted` が出たら user settings のこの項目を確認

## claude を一時 spawn して設定・権限挙動を実測検証する（実践知見 2026-07-22）

設定変更（permissions 等）の効果を、使い捨ての claude セッションで前後比較するパターン。

```bash
PANE=$(herdr pane split <自pane> --direction right --no-focus | python3 -c 'import sys,json; print(json.load(sys.stdin)["result"]["pane"]["pane_id"])')
herdr pane run "$PANE" "claude --permission-mode plan"   # plan mode で直接起動できる
i=0; until herdr pane read "$PANE" --source recent --lines 15 | grep -q 'plan mode'; do i=$((i+1)); [ $i -gt 15 ] && break; sleep 1; done
herdr pane run "$PANE" 'テスト指示（1行）' && sleep 2 && herdr pane send-keys "$PANE" Enter
```

- **起動待ちは statusline の文字列 grep**: plan mode 起動なら `plan mode`。`❯` 単体は Try 例文行にもマッチするので注意
- **許可プロンプトは `agent_status=blocked` として検出できる**: permission dialog 表示中の pane は blocked になる。「プロンプトが出るか」の検証はこれで機械判定できる（出ない場合は working→idle/done に直行）
- blocked になった時点で「プロンプトが出る」という検証結果は確定している。**許可プロンプトへの応答（Yes 選択）をエージェントが send-keys で代行しない** — 権限ゲートの迂回にあたるため、続行が必要ならユーザーに判断を仰ぐ。検証だけならそのまま `pane close` してよい
- **status 待ちは2フェーズ**: まず working への遷移（submit確認）、次に working からの離脱（結果）。両方を1つの until/while ループスクリプトにして background Bash で実行すると1通知で受けられる
- 判定は `pane read` で実物を必ず確認する（「Ran 1 shell command」等の実行痕跡）。検証後は `pane close` で片付ける（セッションごと終了）
