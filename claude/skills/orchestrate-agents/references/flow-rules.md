# フロー規則(全文)

グローバル`CLAUDE.md`の「フロー」節には不変条件だけを置き、段ごとの規則・振り分け・安全の基準・検収運用の全文はここに置く。段ごとの手順とパケット書式は`../SKILL.md`。

### フロー(全タスク共通・これ1本)

1. メインが依頼を受け、**ユーザーとの対話でプラン案を起案する**。段2へ渡すpacketには**5 field(疑う前提 / 反対案 / その帰結 / 未解決の問い / この種の指摘が生じうる経路の列挙)を必須**で載せ、「該当なし」と書くなら理由も書く。プランの確定は段4
2. codex(review役)がプラン査読(**節約モードでも送付先は変わらない**)。**ユーザー承認の代替にしない**
3. **その指摘への対応可否(採用/見送り/別タスク)はfable-review(節約モードではdesign-review)が振り分けて収束させる。この判断を最終とする**。認証・秘密情報を含むプランは振り分け依頼前にredactする。redactすると振り分けが成立しない場合は段3を省き、メインが直接振り分ける
4. **ユーザー承認**。これより前に実装subagentへdispatchを1件も出さない。承認対象はサブタスク方針を含むプラン全体
5. **メインがサブタスクへ分割し、実装subagentへdispatchする**。各subtaskは自己完結パケットで、**スコープのファイルセットを明示する**。並列で走らせる場合はファイルセットを重複させず、**dispatch中はメインが同じrepoを編集しない**。subagentは実行中にメインへ問い合わせられないため、判断が要る場面では変更を加えず`question`で返る。メインが回答してSendMessageで継続させる
6. **メインが受け入れ検査をする**。完了報告の変更ファイル一覧と`git diff --stat`を突き合わせ、**申告外のパスに差分があればそのsubagentの逸脱として差し戻す**(並列時は帰属が確定しないので、説明のつかない差分が1つでもあればゲートを止める)。報告された検証は1-2コマンドでスポット再現する
7. 全subtaskが受け入れ検査を通ったら段8へ。ここは終端ではない
8. メインが統合し、**メインが実質変更したファイル/hunkをAnthropic author(メイン作)へ再分類する**。author再分類マップは「メイン作 / subagent作」の2区分で、findingの差し戻し先を決めるために使う
9. **codex(review役)が全hunkを査読する** — 意図一致・正しさと、脆弱性4観点(認証/認可境界・secret出力・外部write・dependency advisory)の両方。実装がAnthropic側に寄ったため、cross-vendorの査読はcodex1本になる。**codexが応答不能な場合はgrok-reviewへフォールバックする**(判定基準・waiver・停止条件は`claude/skills/orchestrate-agents/SKILL.md`「段9査読者フォールバックチェーン」節が正本)。段8のauthor再分類マップを渡すが、マップは差し戻し先の判定に使うもので査読対象は常に全hunk。段1〜11を通らない原子的編集の例外は、資格要件で認証・security boundary・課金・外部write・依存関係を変えないことが担保されるため脆弱性4観点の対象外。**依存を変えるdiffのadvisory確認はメインの明示責務**とし、段9パケットの必須fieldに確認結果と根拠を含める(欠落時はcommit不可)。**そのfindingsへの対応可否(採用/見送り/別タスク)は、段3と同じくfable-review(節約モードではdesign-review)が振り分けて収束させる。この判断を最終とする**(手順は`claude/skills/orchestrate-agents/SKILL.md`「段9の指摘振り分け」節)
10. **段9で採用と振り分けられたfinding**の戻し先は、codexが付けたラベルに従う3経路 — **subagent作hunk**(`[subtask:<id>]`ラベル)はSendMessageで当該subagentへ差し戻す。セッションを跨いで文脈が失われている場合は新規spawnし、**findingに加えて元のsubtaskパケット(ゴール・制約・ファイルセット)を同梱する**。**メイン作hunk**(`[author:main]`ラベル)はメインが直して段9へ再投入。**設計レベル**(`[design-level]`ラベル。承認済み設計自体の欠陥を担当hunkから見つけたら使う。対応する単一hunkが無ければfile:line欄に`(design-level, no single hunk)`と明記。コードレベルのラベルへ言い換えて縮小しない)はメインが理由を明記して該当taskを終端し、新規`[task:<id>]`を発行して段1から再起動する(旧taskとの関連は理由欄で相互参照。既存taskの延命はしない)。**見送り**と振り分けられたfindingは対応せず、理由を段11の検収報告に記録する。**別タスク**と振り分けられたfindingは今回のtaskを終端せず、新規`[task:<id>]`として別途扱う(スコープ外への先送りであり、`[design-level]`ラベル採用時の「task中止→新task発行」とは別物: 別タスクは今回のスコープ外への先送り、`[design-level]`+採用は今回の承認済み設計自体の欠陥)。**通っている実装まで捨てる必要はない** — 欠陥に対応する差分だけを新規taskとして段1から通し、他の実装は現状のまま保持して両方が揃ってからcommitする。**同種の設計レベル指摘が2周以上続いたら、その周で個別修正を止め「なぜ毎回1件ずつ出るのか」を先に潰す**(列挙の方法が間違っているというシグナルとして読む。次のdispatchは修正でなく調査にし、何を全数列挙するのか・どの軸で数えるのかを明示させる)。網羅性を機械保証できないと判明したら、その旨と再調査手順を成果物の中に記録し、commit messageにも書く(実例: 2026-09-08、gh-dash配色のuse-site登録で3周続き、走査の軸が片方向だけだったのが原因だった)
11. 段9のfindingが全て解消したら、**メインが最終diffを固定し、必須検証を自分で再実行してから検収する**(差し戻しで変わった後のdiffが、査読を通したdiffと同一であることを確認する) → ユーザーのdiffチェック(提示の手順は`claude/skills/orchestrate-agents/SKILL.md`段11) → メインがcommit。検収失敗・ユーザー差し戻しは段10の該当経路へ戻る。**終端は3つ — メインのcommit / ユーザーの中止 / taskの中止**(変更不要・実行不能・段取り不備と確定した場合。メインが理由を記録し、走っているsubagentを停止して終端する)

完了サイクルは**段5のdispatch(またはreopen時の再dispatch)で開き**、段11の検収完了・段10の差し戻し・taskの中止のいずれか1つで閉じる。段7は境界ではなくサイクル内のマイルストーン。

### 実装レーンの2形態(subagent / team)

既定は**subagent**。teamを使うのは、サブタスク同士が作業中に情報をやり取りする必要があるとき、または3人以上を並行させて共有タスクリストで進行管理する規模のときだけ。**この環境ではteammateは別プロセスにならない**(2026-09-06実測。tmux/iTerm2でないとin-processにフォールバックする)ため、「独立した文脈で並行に考えさせる」も「あとから呼び戻して継続させる」もsubagentで取れる。有効化・制約・実効model/effortの確認方法は`claude/skills/orchestrate-agents/SKILL.md`が正本。

Anthropicプランの消費を抑える節約モード(段3と段9のfable-review振り分けをdesign-review、段5をcodex-implへ切り替える)の手順は`claude/skills/orchestrate-agents/references/economy-mode.md`が正本。**このモードは段9の査読をcodexのままにするため、codex-implが書いたコードをcodexが一次査読する。「同一vendorが自分の系列の成果を一次査読する配置を作らない」に反する例外で、2026-09-12にユーザーが受容を明示した。節約モードの中だけで成立し、通常モードには適用しない。**

**このフローの対象**は設計または実装判断を伴う依頼。**例外**は、挙動・公開契約・認証・security boundary・課金・外部write・依存関係を変えず、変更対象と期待diffが一意で、diff単体の機械検証で正しさが確定する原子的編集のみ(行数は基準にしない)。1つでも満たさなければ段1から通す。subtaskが1本でも段5〜11を通す。**例外を通す編集はメインが自分で書き、Anthropic author扱いで段9のcodex(review役)ゲートだけを通す**(subagentへdispatchしない)。

### 振り分け・安全の基準

- **外部送信の安全境界は役割で2軸に分ける**(2026-08-27確定)。(i)調査役(grok-research)への送信packetは secret・credential・個人情報・未公開コード断片を含めない最小化義務を負う。抽象化して意味のあるpacketが作れない問いは`grok-research-skipped:safety`を記録して外部送信しない(codex-researchでの調査続行を妨げない)。(ii)cursor harness経由の査読役(opus-review・grok-review)へは、ユーザーが受容したデータ境界としてprivate diff送付を許容する。対象はCLAUDE.mdが適用される全project、個別repositoryのローカル外部送信禁止規約があればそれを優先、ユーザーが受容を撤回した時点で即時無効。秘密の実値(credential・token等)を含むpayloadは両軸とも対象外(判定は機械的scannerでなくメインの目視確認。詳細はSKILL.md「段9査読者フォールバックチェーン」節)
- **送信前のreadiness照合**: 各nameのregistrationが期待typeでちょうど1件、かつ当該セッションで返信実績があること(dead-letterはregistration照合だけでは検出できない)。**返信実績は直近のspawn/respawn以降の応答に限る**(respawn前の旧instanceの返信実績を新instanceの登録と結びつけて誤ってreadyと判定しない。respawn直後は再度probeし直す)。対象はfable-review・codex(review役)、および実際にdispatchするcodex-research/grok-research/grok-review(agmsg経由の相手のみ。実装subagentはagmsgに乗らないのでこの照合の対象外)。`ensure-headless.sh`はsession team不在でもexit 0のno-opになるため、exit codeだけを準備完了の証拠にしない。**fable-reviewは追加で** `spawn.claude_implementer.fable-review`が未設定(または`false`)であることと、生成済みsettings.jsonのdenyWriteにprojectパスが含まれることを確認する(layout差はregistration照合では検出できない)
- **認証・秘密情報を含む設計プラン/diffは、段3・段9いずれもfable-reviewへ振り分けを依頼する前に該当部分をredactする**。redactすると振り分けが成立しない場合はその段を省き、codexのfindingsの振り分けはメインが直接行う
- **調査は目的で分ける**。未知の挙動を突き止める調査 → codex-research(対象が外部リポジトリのソースでも同様)。メイン直接は既報告citationの1-2コマンドによるスポット確認まで。公式ドキュメント等を読めば答えが出る単純な事実確認はcodex-researchの対象外で、誰が調べるかはClaude Codeの既定に任せる(経路を指定する設定は足さない)
- **grok-researchの併走**は、公開情報かredacted packetだけで閉じる問いに限る。**認証・認可・秘密情報・金銭・データ削除・不可逆な外部操作**を扱う調査と、**security boundary・データ喪失・課金・広範囲migrationの採否を直接決める**調査はgrokへ出さず、codex-research単独+メインの直接裏取りにする。**cursor worker全般**(opus-review・grok-review含む)のread-onlyはcredential denylist型で、project配下の機微設定はdenyされない。**秘密の実値**(credential・token等の値そのもの。認証ロジックに触れるdiff全般ではない — 段9の脆弱性4観点は認証/認可境界を含み、これはopus-review/grok-reviewが査読する対象そのもの)を含む査読はcursorへ出さない(機械的secret scannerは無く、メインがdispatch直前に最終outbound payload全体を目視確認する運用。誤判定は残存リスクとして受容)。**fable-review(claude-code)**はBash実行可でproject配下read全域が開くうえsandboxのcredential denyも及ばないため、credential read・外部状態変更の禁止はrole file規約でのみ抑止する(強制境界ではない)
- **併走の判定はdispatch前**に行い、根拠を`[task:<id>]`へ記録する。先行結論をblindに渡さず同じ問いを独立に調べさせる
- **併走の目的は発見の多様性であり、判定の統計的検証(不一致率・閾値・判定不能率等)の道具にしない**。統合はメインが両トラックの結果を目視で比較し、食い違いは成果物に明記する運用を既定にする。定量的な検証機構を(自分であれ査読者であれ)提案したくなったら、対象規模(件数・システムの重要度)に見合うかを先に自問する(実例: 2026-09-16、issue triageの併走に不一致率3割の閾値を作り込もうとしてcodex査読が2巡連続でHighを出し、advisor()の指摘で目視比較まで簡素化)
- 委譲中はcodex-researchの返却まで同じ対象領域のRead/Grep/Globを控える。安全・権限・緊急性で例外的に読む場合はその理由を記録する
- workerのraw dumpをメインやユーザーへ転載せず、`decision / evidence(file:lineまたはURL) / unknown / next action`だけを受け取る。`[research]`は1トピック=1パケットに分ける
- codex系への送信は非同期send既定(返信はagmsg Monitorの自動再開で受ける)。**ただしagent teamのteammateは、idle中に自分のMonitorの通知では起きない**(sunagawasei/agmsg#60)。teammateはforegroundの`until`ループで待ち、leadが中継する(`claude/skills/parallel-handoff/references/teammate-mode.md`)。送信前のensure-codex・パケット書式・Q&Aループ・検収ゲートの詳細: `claude/skills/orchestrate-agents/SKILL.md`
- 同一model・同一課金プールへの委譲はコスト裁定が消えるため、往復の価値が無い作業はメインが直接行う(別プール=codex系・cursor系への委譲はこの制約を受けない)。**実装subagentはAnthropicプール内だが、メインOpus / subagent sonnetの単価差が裁定になる**。メイン自身がsonnetで動くセッションでは委譲しない

### 検収・レビュー運用

- **段9のdiff査読はcodex(review役)が全hunkを担当する**: 意図一致・正しさと脆弱性4観点(認証/認可境界・secret出力・外部write・dependency advisory)の両方。実装がAnthropic側(メイン・実装subagent)に寄ったため、cross-vendorの査読はcodex1本になる。**同一vendorが自分の系列の成果を一次査読する配置を作らない**の原則は、これで満たす。**codex応答不能時はgrok-review(cursor・`cursor-grok-4.6-xhigh`、暫定運用)へフォールバックする**(詳細はSKILL.md「段9査読者フォールバックチェーン」節)。opus-reviewはAnthropicなので平常時の査読には使わず、高リスク変更でメインが必要と判断したときの第2意見として休眠させる。**依存を変えるdiffのadvisory確認はメインの明示責務**とし、段9パケットの必須fieldに確認結果と根拠を含める(欠落時はcommit不可)。**codexのfindingsへの対応可否(採用/見送り/別タスク)は、段3と同じくfable-review(節約モードではdesign-review)が振り分けて収束させる**(振り分け後の差し戻しはメインが行う。詳細は`claude/skills/orchestrate-agents/SKILL.md`「段9の指摘振り分け」節)
- 多様性の判定はタスク開始時とゲート通過時の2回、author agent/model/vendor/pool/primary reviewer/riskを記録して照合する。**課金プールの違いはvendor多様性に数えない**(Cursor経由のClaudeはAnthropic、同経由のGPTはOpenAI)。`auto`指定は実効vendorが確定できないため査読ゲートで使わない。alias指定(fable-review=`fable`等)のロールは解決先が実行時に決まるため、**記録は静的設定値に留め、動的な実効modelの機械照合はできないことを残存リスクとして扱う**(claude-code driverにはcursorのようなmodel-audit機構が無い)
- 査読は検収の代替ではない。メインが実物(`git status`/`git diff`/`git log`/test)を確認して完了とcommitを決める
- 段9の査読も検収の代替ではない。findingが全て解消したらメインが最終diffを固定し、必須検証を自分で再実行してから検収する(差し戻しで変わった後のdiffが査読を通したdiffと同一であることを確認する)
- **GitHub Issue・PR・commentの作成/変更は、ユーザーの明示指示があるときだけ**。メインの承認だけでは行わない。既定の台帳はagmsg DBと`[task:<id>]`
- 重要な判断はサブエージェントに委譲せず、メインが直接行う。実装subagentは実行中に問い合わせられないため、判断が要る場面では変更を加えず`question`で返し、メインが回答して継続させる
- 検収手順とcodex査読の使いどころは`claude/skills/orchestrate-agents/SKILL.md`。配分の根拠データは`.claude/docs/cred-split/`
