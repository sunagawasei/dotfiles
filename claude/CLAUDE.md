# グローバルCLAUDE設定

全プロジェクトに適用されるグローバル設定。汎用ルールは `claude/rules/`（グローバル設定dir）に配置する。プロジェクト固有のルールは各プロジェクトの `.claude/rules/` に置く。

## ルールファイル

以下のルールが全プロジェクトに適用されます：

- **Go言語規約**: 詳細は `claude/rules/golang.md` を参照
- **セキュリティガイドライン**: 詳細は `claude/rules/shell-security.md` を参照

これらのルールは、該当するファイルを編集する際に自動的に適用されます。

## スクリプト言語

- スクリプトやCLIツールを新規作成する際は、**常にGo言語（Golang）を使用**すること
- Python、Bash、Node.jsなどの他の言語は、ユーザーが明示的に指定した場合のみ使用可

## エージェント役割分担

**Claude=唯一の「手」と安全ゲート / cursor=実装前プラン / codex=実装後レビュー**。振り分けはClaudeが判断。狙いは高単価な Opus の出力を「実装」に集中させ、前段の調査・計画は別subscriptionの cursor に、レビューは codex に外注してトークンを節約すること（codex は前段の調査・計画には使わない＝review 専任）。

| 役 | 担当 | 権限 |
|---|---|---|
| **Claude（本セッション）** | 実装・実行・git・権限/sandbox/auto-mode・最終統合・cursor/codex 結果の採否 | write（唯一の実行主体） |
| **cursor** | 実装**前**の横断調査＋プラン（変更箇所特定・依存把握・パッケージ比較 → file:line 手順＋代替案）。強み=幅 | read-only |
| **codex** | 実装**後**の diff 査読（正しさ・エッジ・回帰リスク → Findings のみ）。強み=深さ | read-only |

振り分け基準:
- **実装・実行・安全判断** — 本セッション自身（手放さない）
- **小さな調査（1-2ファイル/明確な grep）** — Explore サブエージェント or Claude 直接
- **中〜大タスクの計画・横断調査（3+ファイル/新機能/refactor/パッケージ選定/アーキ不明）** — cursor に先行 `ask`
- **実装後レビュー** — codex に `ask`（下記ループ）
- **実装で迷う/問題に直面** — cursor（別解・発散の案出し）と codex（候補案の検証・反証）の2視点。codex には計画でなく「この案で問題ないか」の検証を投げる（review の一種で専任と整合）

トークン節約ガード（中〜大タスク時）:
- cursor プラン返却まで対象領域の Read/Grep/Glob を控え、実装は cursor が cite したファイルを起点に開く（不足・安全確認に必要なら Claude が追加で開いてよい。最終的な正しさ/安全判断は Claude が持つ）
- cursor/codex の出力は要約してユーザーに転載せず、**採用項目だけ実装に反映**（コンテキスト膨張＝トークン増を防ぐ）
- 同一 diff を cursor と codex の両方には回さない（例外: 大規模 refactor のみ cursor 事前影響分析 → 実装 → codex 査読）

### cursor 依頼の鉄則（実装前プラン）

- cursor は **read-only**。プランを返すだけで**実装は Claude**。cursor に diff レビューを振らない（codex の領域）
- agmsg: 宛先 cursor・prefix `[plan]`。自己完結パケット = GOAL / CONSTRAINTS / SCOPE（対象パス）/ OUTPUT（手順 file:line・trade-off 2-3案・risks）/ DO NOT implement

### codex 依頼の鉄則（実装後レビュー）

- codex は **read-only**。findings を返すだけで、**fix は Claude が適用**する。依頼は review/verify/findings/test-plan のみ。codex に計画を振らない
- agmsg: 宛先 codex・prefix `[review]`。自己完結パケット = `git diff` か対象 `file:line` ＋ 意図 ＋（ループ時）前回指摘→対応の対応表（codex はメッセージ本文しか見ない）
- 出力形式: Findings / Required tests / Residual risk / Confidence。severity 順・推測は明記
- 対象は **実質的な実装のみ**。1行修正など些細な編集はレビュー不要

### 実装後レビュー（クリーンになるまでループ）

**実質的な**実装（機能・タスク完了の節目。1行修正など些細な編集は除く）をしたら、その diff を codex に `ask` してレビューを受ける。

**指摘の反映（再修正）それ自体も「実装」なので、反映後の差分を再び codex に `ask` する。** これを codex が substantive な findings を返さなくなる（「Findings なし」等）まで繰り返してから完了とする。1回のレビュー＋修正で止めない。

- ループの各巡で「前回指摘 → 対応」の対応表を自己完結パケットに含め、修正の検証と新規問題の両方を見てもらう。
- 純粋に些末な nit のみになったら収束とみなしてよい（鉄則の「些細な編集はレビュー不要」と整合）。
