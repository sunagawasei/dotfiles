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
- **`idle` は「失敗」ではなく「入力待ち」。status だけで成否を判定しない**（2026-07-29）: `idle` には「作業を終えた」と「agent が自発的に質問して待っている」の両方が含まれる。`blocked` は権限プロンプト表示中を指すので、**確認待ちを `blocked` で拾おうとすると取りこぼす**。監視スクリプトで `idle` を失敗に分類したところ、実際には agent が指示どおり「switch 前に確認をお願いします」と停止していただけだった。`idle`/`done` を受けたら必ず `pane read` で画面を読んでから判断する
- **成否の判定条件は agent の status ではなく客観的な事実に置く**（2026-07-29）: 「パッケージを入れる」なら agent の完了ではなく**バイナリの実在**（`[ -x /etc/profiles/per-user/<user>/bin/<cmd> ]` 等）を条件にする。agent の `done` は自己申告に過ぎず、home-manager の `switch` が未実行なら何も起きていない。加えて `idle`/`done` かつ目的物が無いケースを明示的に拾わないと、その状態が沈黙として見過ごされる
- **claude を spawn する workspace / tab に `--label` を付けない**（2026-07-29ユーザー指示）: label を付けると手動リネーム扱いになり、**root pane 由来の自動命名（= claude のセッションタイトル）が反映されなくなる**。agent が起動すればタイトルは依頼内容（例: `Add buf to home-manager packages.nix`）に変わり、固定文字列より情報量が多い。spawn 前の視点で分かりやすい名前を付けたくなるが、起動後はタイトルに任せる
- herdr socket（`$XDG_CONFIG_HOME/herdr/herdr.sock`）は sandbox の `allowUnixSockets` に登録が必要。`Operation not permitted` が出たら user settings のこの項目を確認

## statusline の codex busy アイコンが出ない（実践知見 2026-07-29）

`~/.config/claude/statusline.sh` は agmsg の run ディレクトリを読んで codex bridge の稼働を検出し、2段目左端に `󰚩` を出す。bridge が実際に動いているのにアイコンが出ない場合、**agmsg 側の形式変更に statusline が追従できていない**可能性が高い。プロデューサ（agmsg スクリプト群）とコンシューマ（statusline.sh）が別リポジトリにあるため、形式変更が伝播しない。

検出は4段のゲートを通る。切り分けは上から順に手で再現する。

1. **meta ファイルの team/type 一致** — `run/codex-bridge.<team>.<name>.meta` を読む
2. **pidfile + `kill -0`** — プロセス生存確認
3. **`ps` で `codex-bridge.js` の実体と識別子の一致**
4. **bridge ログの lifecycle 解析** — `armed` 後に `wakeup` があれば busy

2026-07-29 に壊れていたのは 1 と 3 で、原因は agmsg の形式変更だった。

| ゲート | 旧形式（statusline が期待） | 新形式（agmsg の実際） |
|---|---|---|
| 1. meta | `team=<team>` 行 | **`identities=<team>/<name>`**（`team=` 行は無い） |
| 3. プロセス照合 | `--team <team> --name <name>` 引数 | **`--identity-key <base64(team\tname)>:`** |

`identity-key` は `printf '%s\t%s' "$team" "$name" | base64 | tr -d '\r\n' | tr '+/' '-_'` に終端 `:` を付けた値（`scripts/lib/identity-key.sh` の `agmsg_identity_key`）。修正は両ゲートを**新旧どちらでも通る形**にする（`identities` の `/` 前を team として読む / identity-key と旧引数を OR で照合）。

- **廃止の痕跡は変更した側のコメントに残る**: `ensure-codex.sh` に「rather than the retired `--team/--name` signature」と明記されていた。消費側が壊れたら、生産側スクリプトの最近のコメントを grep するのが早い
- **sandbox 内では `kill -0` と `ps` が `operation not permitted` で失敗するため、ゲート2・3を自分で検証できない**。statusline 本体は sandbox 外で走るので、ゲート1の修正を確認したら実際の表示で見てもらう
- 反映は `statusLine.refreshInterval`（設定されていれば数秒）で自動的に起こる

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
