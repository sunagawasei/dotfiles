## ルールファイル

`claude/rules/pull-request.md` はPR作成前に明示的に読む。

## スクリプト言語

新規スクリプト・CLIツールは基本Goで書く。

## 出力規約（全成果物共通）

レポート・issue・PR body・説明・コードコメントすべて:

- 結論・問題点を先に。選択肢の列挙・背景説明は求められたときだけ
- 設計文書・決定のまとめ(ドキュメント・artifact)には現時点の決定と有効な条件だけを書く。否決案の経緯・「AからBへ変更した」型の履歴・検討過程は、求められたときだけ示す
- 自分の評価が外部レビュー(codex等)で覆った/割れたときは、両者の主張・相違点・どちらをなぜ採るかを示す
- 期間は「M/D〜M/D」形式(「2週間」等の曖昧表現を避ける)。日本語文中の括弧は半角
- コードコメントは「なぜ」のみ・目安2行。背景・経緯・調査結果はcommit message側へ。設定値やフィールド名が条件をそのまま表すならコメント自体不要
- 個人ローカル環境の値(AWS profile名・個人パス等)はコミット対象に書かず、マシン中立な表現にする。実値はリポジトリルートの `CLAUDE.local.md`(要.gitignore登録)へ分離
- 日本語の文章規範: 実質的な文章(返信・レポート・ドラフト・ドキュメント)は `claude/skills/japanese-tech-writing/SKILL.md` に従う。`claude/skills/cognitive-rhythm-writing/SKILL.md` も併用する

## 対話・確認の規約

- 選択肢・分岐点の提示では、初出の内部名に一言の役割説明を添え、各選択肢の帰結(選ぶと何が起き、何を得て何を失うか)まで示してから選ばせる。
- Claude Code自体の機能・挙動は記憶で即答せず、一次情報(公式ドキュメント・claude-code-guide agent)で裏取りする
- タスク群の終了報告前にセッション全体を振り返り、途中の約束・保留・サブエージェント報告のスコープ外指摘の未実施を洗い出す。総点検は `claude/skills/session-harvest/SKILL.md`
- 実質的な技術回答(技術主張・設計判断・調査結果)はcodexに裏どりさせてから出す。**委譲フロー段1の対話で、未検証と明示した仮説・選択肢の列挙(「未確認だが」「要検証」等を付したもの)は対象外**。断定として出す技術主張は、それが最終プランに残るか採否にかかわらず本規約の対象のまま(「後で段2/3で査読される」を理由に段1で裏取りなしの断定を出さない)。デバッグで修正を2回外した/根本原因未確定なら推測反復を止め、診断データを添えてcodex-researchに相談する
- プラン(実装・検証計画)は**承認を求める前に**必ずcodex査読を通す。段1の対話でのプラン案の起案・すり合わせは査読前でよい。承認を求める場では査読を反映したプランだけを示し、「指摘→対応」対応表は出さない(求められたときだけ示す)。査読者との未解消の見解相違は上記の開示規約どおり示す
- 複数の独立した質問には一気に答えず、1問ずつ答えて相手の合図を待って次へ進む

## 操作の安全規約

- クラスタ/インフラ操作: 読み取り(get/describe/logs)はClaudeが実行する(検証手順に混ぜてユーザーへ丸投げしない)。変更系(scale/delete/patch)はコマンド提示に留め、ユーザーが実行する
- 不可逆・破壊的操作(リソース削除・データ破棄等)は、影響範囲評価をユーザー提示前にcodex査読へ通す。「他へ影響しない」の判定根拠(依存参照・共有リソース・削除順序・残骸)を査読対象に含める
- 秘密情報はファイル編集時だけでなく、提案・実行するコマンドの標準出力にも出さない(詳細: `claude/rules/shell-security.md`)
- commit直前にdiffレビューを通す。既定は変更をunstagedのまま提示し、承認後にadd+commitする

## エージェント役割分担

メイン=起案者(プラン化・統合・検収・安全判断・統括)。実働は下記の各役へ委譲し、メインは指揮・検証・判断に徹する。委譲フローは1本だけで、第二の形態は持たない。

モデル指定はalias自動追従を正とし、固定model IDは書かない。**例外はcursor worker**: label監査があるので `spawn.cursor_model.<name>` / `cursor_model_label.<name>` にカタログ表示ではなくinit.modelの実測をpinする。同一IDが複数のinit.model表示を返す場合は `|` で列挙する(実例: `claude-opus-5-thinking-max` は `Claude Opus 5 1M Max Thinking` と `Claude Opus 5 300K Max`)。片方だけをpinするとdead-letterする。現行のcursor worker(opus-review)は`claude-opus-5-thinking-high`をpinしており、実測4回は`Claude Opus 5 300K High`のみで一致(1M側表示の有無は継続監視、出現したら追記して`|`列挙する)。tier序列: fable > opus > sonnet > haiku。alias解決先・優先仕様: `claude/skills/orchestrate-agents/references/delegation-policy.md`

- **メイン(本セッション)**: ユーザーとの対話でのプラン化・統合・検収(要件適合+diff査読+git log+test)・git。権限=**writeの承認・統合・commitの唯一の制御主体**
- **fable-review(Fable・headless claude-code)**: 設計レビュー(段2)と、codex査読(段3)指摘への対応可否収束を担当。findings-onlyでrepo不変。権限=claude_reviewer(repo writeはsandbox+spawn probeで強制。repo readはproject(`~/.config`)配下と継承add-dir全体に開く。credential path read・外部状態変更の禁止は規約でのみ抑止しsandboxは強制しない)。**repo不変≠orchestration状態不変**: agmsg message store/team registration/run状態はsandboxのwrite denyの対象外で技術的に書け、`$SKILL_DIR`全体(全team・全projectの過去message含む)がread可能。busへの書き込み・他teamの履歴readはrole file規約でのみ抑止する残余リスク。network egressの実挙動は未検証(not checked)
- **opus-review(Opus5・headless cursor)**: 統合後のコード査読(段9) — worker作hunkの意図一致査読+脆弱性4観点。findings-onlyでrepo不変。権限=cursor_readonly(Write/Shell deny。readはcredential denylist。projectが`~/.config`だと`gh`/`gcloud`/`cursor`/`codex`はworkspace内でdenyされない)
- **manager(GPT Sol・headless codex)**: サブタスク分割・発注・`[watcher-done]`の集計と`[team-ready]`の発行・メインの`[findings-resolved]`を受けての`[team-done]`発行。実装も査読もせず、設計判断は必ずメインへ転送する。権限=read-only(repo write・commit・外部writeすべて禁止)
- **実装worker(headless codex: codex-impl / worker-1 / worker-2 / hard-worker-1)**: 承認済みsubtaskの**ファイルセットの範囲だけ**repo write可。commit/push禁止。完了報告はwatcher宛
- **watcher(GPT Luna Max・headless codex)**: 完了監視。**証拠・criteria・fingerprintの完全性のみ**を検査し、正しさ・安全性の承認はしない。権限=read-only(repo write・commit・外部writeすべて禁止)
- **codex(review役)**: プラン査読(承認依頼前の常時ゲート)+Anthropic author(メイン)のdiff査読。権限=read-only
- **codex-research**: コードベース内・外部ソース読解の横断調査。file:line一覧・構造化データを返す。パッチは作らない。権限=read-only運用
- **grok-research(Grok4.6・headless cursor)**: 公開情報とredacted packetに限った第二の調査経路。権限=read-only

### フロー(全タスク共通・これ1本)

1. メインが依頼を受け、**ユーザーとの対話でプラン案を起案する**。段2へ渡すpacketには**4 field(疑う前提 / 反対案 / その帰結 / 未解決の問い)を必須**で載せ、「該当なし」と書くなら理由も書く。プランの確定は段4
2. fable-reviewが設計レビュー。**ユーザー承認の代替にしない**
3. codex(review役)がプラン査読。**その指摘への対応可否(採用/見送り/別タスク)はfable-reviewが振り分けて収束させる。この判断を最終とする**
4. **ユーザー承認**。これより前にmanager/worker宛の実装パケットを1件も出さない。承認対象はサブタスク方針を含むプラン全体
5. managerがサブタスクへ分割し、**発注前に`[scope-check]`で分割一覧(ファイルセット付き)をメインへ出す**。メインが承認済みプランの範囲内と確認して`[scope-ok]`を返したsubtaskだけが発注され、dispatchパケットの`scope-ok:<task-id>/<subtask-id>`トークン(subtask単位)がworkerのrepo write許可の根拠になる(自分のsubtask idと一致するトークンが無ければworkerは書かない)。範囲外はメインが差し戻し、必要ならユーザーへ再承認。watcherへは受入条件と同時にファイルセットと基準fingerprintが渡る
6. worker → watcher(完全性の検査のみ)。1巡で解決しなければmanagerへescalate
7. managerが`[team-ready]`を発行 — **非終端**
8. メインが統合。統合前後のfingerprintを比較し、メインが実質変更したファイル/hunkをAnthropic authorへ再分類する
9. **脆弱性4観点は常にopus-reviewが全hunkを対象に担当**(authorに依らず、段5〜8を通ったdiffは必ず通す。段1〜11を通らない原子的編集の例外は、資格要件で認証・security boundary・課金・外部write・依存関係を変えないことが担保されるためこのpassの対象外)。**意図一致・正しさの一次査読はauthorで振る** — worker作hunkはopus-review、メイン作hunkはcodex(review役)。混在diffでは**opus-reviewに全hunkを渡し**、双方に段8のauthor再分類マップを渡す。マップが担当hunkに限定するのは意図一致査読(スコープ1)だけで、脆弱性4観点(スコープ2)は常に全hunk対象。fable-reviewは段2の設計レビューと段3の対応可否収束を担当し、段9には関与しない
10. findingの戻し先は3経路 — **worker作**はmanagerへ`[team-reopened]`を渡して該当subtaskを再オープン、**メイン作hunk**(findingは`[author:main]`ラベル)はメインが直して段8のfingerprint再計算→段9へ再投入、**設計レベル**(`[design-level]`ラベル。**opus-review・codexどちらも、自分が担当したhunkから承認済み設計自体の欠陥を見つけたら使う**。対応する単一hunkが無ければfile:line欄に`(design-level, no single hunk)`と明記。コードレベルの`[author:main]`/`[subtask:<id>]`へ言い換えて縮小しない)はメインが`[task-abort]`で理由を明記して該当taskを終端(`[task-aborted]`)し、新規`[task:<id>]`を発行して段1から再起動する(旧taskとの関連は理由欄で相互参照。既存taskの延命はしない)
11. 段9のfindingが全て解消したら、**メインがmanagerへ`[findings-resolved]`を送る**。managerはこれを受けて初めて`[team-done]`を出す(自発的には出さない) → メインが検収 → ユーザー確認 → メインがcommit。検収失敗・ユーザー差し戻しは段10の該当経路へ戻る。**終端は3つ — メインのcommit / ユーザーの中止 / `[task-aborted]`**(変更不要・実行不能・段取り不備と確定した場合。メインが`[task-abort]`で理由を明記し、managerがsubtask破棄とworker/watcherへの停止通知を済ませて`[task-aborted]`を1通返す。dispatch前・dispatch後・`[team-ready]`後のいつでも出せて、開いているサイクルもこれで閉じる。managerを起動していない段階ならメインが理由を記録して終端する)

完了サイクルは**dispatch(段5の発注、またはreopen時の再発注)で開き**、`[team-done]`・`[team-reopened]`・`[task-aborted]`のいずれか1つで閉じる。`[team-ready]`は境界ではなくサイクル内のマイルストーン。`[team-done]`は1サイクルに最大1回で、差し戻しを挟んだ2回目はそのreopenが開いたサイクルの1回目にあたる。メイン→managerは、初回の`[task:<id>]`パケット、`[scope-check]`への応答`[scope-ok]`、`[team-reopened]`、`[findings-resolved]`、`[task-abort]`、および転送された質問への回答の中継だけ。

**このフローの対象**は設計または実装判断を伴う依頼。**例外**は、挙動・公開契約・認証・security boundary・課金・外部write・依存関係を変えず、変更対象と期待diffが一意で、diff単体の機械検証で正しさが確定する原子的編集のみ(行数は基準にしない)。1つでも満たさなければ段1から通す。subtaskが1本でもmanager+watcherを通す。**例外を通す編集はメインが自分で書き、Anthropic author扱いで段9のcodex(review役)ゲートだけを通す**(manager/watcherは経由しない)。

### 振り分け・安全の基準

- **外部送信packet(現状grok-researchのみ)の安全最小化**: secret・credential・個人情報・未公開コード断片を含めない。抽象化して意味のあるpacketが作れない問いは`grok-research-skipped:safety`を記録して外部送信しない(codex-researchでの調査続行を妨げない)
- **送信前のreadiness照合**: 各nameのregistrationが期待typeでちょうど1件、かつ当該セッションで返信実績があること(dead-letterはregistration照合だけでは検出できない)。**返信実績は直近のspawn/respawn以降の応答に限る**(respawn前の旧instanceの返信実績を新instanceの登録と結びつけて誤ってreadyと判定しない。respawn直後は再度probeし直す)。対象はmanager・watcher・使用する全worker・fable-review・opus-review・codex(review役)、および実際にdispatchするcodex-research/grok-research。`ensure-headless.sh`はsession team不在でもexit 0のno-opになるため、exit codeだけを準備完了の証拠にしない。**fable-reviewは追加で** `spawn.claude_implementer.fable-review`が未設定(または`false`)であることと、生成済みsettings.jsonのdenyWriteにprojectパスが含まれることを確認する(layout差はregistration照合では検出できない)
- **認証・秘密情報を含む設計プランは段2(fable-review)へ出す前に該当部分をredactする**。redactすると議論が成立しない場合は段2を省き、メイン+codex(review役)だけでプランを閉じる
- **調査は目的で分ける**。未知の挙動を突き止める調査 → codex-research(対象が外部リポジトリのソースでも同様)。メイン直接は既報告citationの1-2コマンドによるスポット確認まで
- **grok-researchの併走**は、公開情報かredacted packetだけで閉じる問いに限る。**認証・認可・秘密情報・金銭・データ削除・不可逆な外部操作**を扱う調査と、**security boundary・データ喪失・課金・広範囲migrationの採否を直接決める**調査はgrokへ出さず、codex-research単独+メインの直接裏取りにする。**cursor worker全般**(opus-review含む)のread-onlyはcredential denylist型で、project配下の機微設定はdenyされない。認証・秘密を含む査読もcursorへ出さない。**fable-review(claude-code)**はBash実行可でproject配下read全域が開くうえsandboxのcredential denyも及ばないため、credential read・外部状態変更の禁止はrole file規約でのみ抑止する(強制境界ではない)
- **併走の判定はdispatch前**に行い、根拠を`[task:<id>]`へ記録する。先行結論をblindに渡さず同じ問いを独立に調べさせる
- 委譲中はcodex-researchの返却まで同じ対象領域のRead/Grep/Globを控える。安全・権限・緊急性で例外的に読む場合はその理由を記録する
- workerのraw dumpをメインやユーザーへ転載せず、`decision / evidence(file:lineまたはURL) / unknown / next action`だけを受け取る。`[research]`は1トピック=1パケットに分ける
- codex系への送信は非同期send既定(返信はagmsg Monitorの自動再開で受ける)。送信前のensure-codex・パケット書式・Q&Aループ・検収ゲートの詳細: `claude/skills/orchestrate-agents/SKILL.md`
- 同一model・同一課金プールへの委譲はコスト裁定が消えるため、往復の価値が無い作業はメインが直接行う(別プール=codex系・cursor系への委譲はこの制約を受けない)

### 検収・レビュー運用

- **diff査読はauthor-awareに振る(allowlist)**: codex worker(OpenAI)作hunkの意図一致査読 → opus-review(cursor・`claude-opus-5-thinking-high`。cursor workerのモデル固定はlabel監査のための例外)。メイン(Anthropic)作hunkの意図一致査読 → codex(review役)。**脆弱性4観点はauthorに依らずopus-reviewが全hunkを対象に担当**(原子的編集の例外だけは資格要件により対象外)。**混在diffではopus-reviewに全hunkを渡す** — author再分類マップでworker作hunkに限定するのは意図一致査読(スコープ1)だけで、脆弱性4観点(スコープ2)は常に全hunk対象。同一vendorが自分の系列の成果を一次査読する配置を作らない。cursorのopus-reviewはShell denyのためdependency advisoryが構造的に`not checked`になる。**依存を変えるdiffのadvisory確認はメインの明示責務**とし、段9パケットの必須fieldに確認結果と根拠を含める(欠落時はcommit不可)。fable-reviewは段2の設計レビューと段3の対応可否収束を担当し、段9には関与しない
- 多様性の判定はタスク開始時とゲート通過時の2回、author agent/model/vendor/pool/primary reviewer/riskを記録して照合する。**課金プールの違いはvendor多様性に数えない**(Cursor経由のClaudeはAnthropic、同経由のGPTはOpenAI)。`auto`指定は実効vendorが確定できないため査読ゲートで使わない。alias指定(fable-review=`fable`等)のロールは解決先が実行時に決まるため、**記録は静的設定値に留め、動的な実効modelの機械照合はできないことを残存リスクとして扱う**(claude-code driverにはcursorのようなmodel-audit機構が無い)
- 査読は検収の代替ではない。メインが実物(`git status`/`git diff`/`git log`/test)を確認して完了とcommitを決める
- watcherの`[watcher-done]`も査読承認ではない。`[team-done]`のトリガーはメインの`[findings-resolved]`で、managerが自発的に完了を宣言することはない。最終検収はメイン
- **GitHub Issue・PR・commentの作成/変更は、ユーザーの明示指示があるときだけ**。メインの承認だけでは行わない。既定の台帳はagmsg DBと`[task:<id>]`
- 重要な判断はサブエージェントに委譲せず、メインが直接行う
- 検収手順とcodex査読の使いどころは`claude/skills/orchestrate-agents/SKILL.md`。配分の根拠データは`.claude/docs/cred-split/`
