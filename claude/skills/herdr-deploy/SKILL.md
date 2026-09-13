---
name: herdr-deploy
description: "herdrのローカルパッチを実装し配備するときの手順。パッチファイル(home-manager/patches/)の作成、herdr.nixへの登録、packages.nixとの関係、darwin-applyによる反映、herdr serverの再起動までを扱う。herdrの挙動を変える作業、パッチの追加や変更、herdr.nixやpackages.nixの編集、配備手順の確認に入ったときに使う。"
---

# herdrのローカル変更を配備する

herdrは flake input のソースへローカルパッチを当ててビルドする。**commitしただけでは稼働バイナリは変わらない。** パッチファイル化、登録、`darwin-apply`、herdr serverの再起動を経て初めて反映される。

## 配備の経路

各手順の末尾に、対応する検査コマンドの段番号(段0〜段3の定義は後述)を添える。

1. 開発ツリー(git worktree、branch `v082-upgrade`)で実装してcommitする(段0)。
2. その差分を`.patch`ファイルとして`home-manager/patches/`へ置く(段1)。
3. `home-manager/herdr.nix`の`patches`リストへ登録する(段1)。**`herdr-combined-frame-digest.patch`は必ず最後に置く。** 先行するUIパッチ全部の合成描画を固定するintegration fixtureなので、UIパッチを変えたらdigestを再生成する。並び順は開発branchのcommit順と一致させる。各パッチは親commit時点のツリーに対するdiffなので、順を崩すとoffset依存になる。
4. パッチファイルと`herdr.nix`を`git add`する(段0)。Nix flakeはgit追跡ファイルしか見ないため、未追跡だと"Path ... is not tracked by Git"でNix評価が失敗する。
5. `darwin-rebuild build --flake ~/.config`で評価とビルドだけ先に検証する(sudo不要)。
6. `darwin-apply`で適用する(段2)。
7. herdr serverを再起動して新バイナリを有効化する(段3)。**全pane が終了する重大操作なので、ここはユーザーが実行する。** 適用後から再起動までの間は、herdrのCLIが`protocol_mismatch`で全て偽を返し、herdr依存のhookが止まる。

手順2から手順7のどれか一つが欠けても稼働バイナリは変わらない。

## 検査コマンド

`scripts/cmd/verify-herdr-deploy`が、次の段0から段3までを判定する。**落ちた段が、次にすべき操作と1対1で対応する。**

```bash
cd scripts && go run ./cmd/verify-herdr-deploy --dev-tree <開発ツリーのパス>
```

- **段0(前提の検査)**: 開発ツリーの指定、開発ツリーに未commitの変更が無いこと(手順1)、`home-manager/patches/`と`home-manager/herdr.nix`がgit add済みであること(手順4)、開発ツリーが`flake.lock`のpin revを含むこと、開発ツリーの`nix/`が上流revと同一であることを見る。
- **段1(開発ツリー↔パッチファイル)**: 落ちたら、差分の向きで次の操作が変わる(向きの内訳は後述「開発ツリーが両方向にずれうること」を参照)。開発ツリー側にだけある実装をこれから配備する場合に限り、手順2(パッチファイル化)と手順3(登録)を行う。**過去2回(2026-07-13、2026-09-10)の反映漏れは、いずれもこの向きでの対応漏れだった。**
- **段2(設定↔ビルド済みシステム)**: 落ちたら手順6(`darwin-apply`)を実行する必要がある。
- **段3(ビルド済みシステム↔稼働プロセス)**: 落ちたら手順7(herdr serverの再起動)を実行する必要がある。

`--stage N`で段を単独実行できる(段0は前提として常に実行される)。開発ツリーは`--dev-tree`、設定リポジトリは`--flake`で指定する。

## 開発ツリーのパスの設定

開発ツリーのパスは個人ローカルの値なので、環境変数`HERDR_DEV_TREE`か`--dev-tree`フラグから渡す。実値はこのスキルにも他のcommit対象ファイルにも書かない。`HERDR_DEV_TREE`に何を設定するかは、リポジトリルートの`CLAUDE.local.md`(`.gitignore`登録済み)に書く。

## パッチ列の途中へ新しいパッチを挿す

`combined-frame-digest`は必ず最後尾なので、新しいパッチは常にその手前へ入る。開発branchのcommit順も
配備順と一致させる必要があるため、末尾にcommitを積むだけでは済まない。**`git reset --hard`は使わない**
(未commitの変更を巻き込む事故の経路になるうえ、元のtipが名前付きrefから外れる)。

1. 事前検証(読み取りのみ): `git -C <dev> rev-parse --abbrev-ref HEAD` / `git -C <dev> status --porcelain`が空 /
   `git -C <dev> rev-parse <配備branch>`を記録する。
2. `git branch <配備branch>-pre-<slug> <現tip>` で**元のtipを名前付きrefで保全する**。
   `git branch --contains <現tip>` にこのbranchが出ることを確認する。
3. `git checkout -b <slug> <digestの1つ前のcommit>` → 実装 → commit。
4. `git cherry-pick <digest commitのsha>` でdigestを最後尾へ戻す。
5. `<slug>`をチェックアウトした状態で `git branch -f <配備branch> <slug>` → `git checkout <配備branch>`。
6. `git diff <新commit>^ <新commit> -- src/` でパッチファイル化する。
7. 作業用branch(`<slug>`等)は配備branchへ畳んだ後に `git branch -D` で掃除する。保全用の
   `<配備branch>-pre-<slug>` は、稼働確認が済むまで残す。

digestの期待値を再生成する必要があるかは、**そのパッチが`Mode::Terminal`と`Mode::Navigate`のフレームを
変えるか**で決まる。digest fixture(`src/ui/tab_surface.rs`)はこの2モードしか描かないので、
ダイアログのoverlayだけを変えるパッチなら再生成は不要。判定はcherry-pick後に
`nix develop --command cargo test --bin herdr ui::` が緑かどうかで行う。

段9の査読で修正が入って実装commitを差し替えるときは、作業用branchを実装commitに置き直して
amendし、手順4以降をやり直す。

## テストの走らせ方

`cargo`は`nix develop`の中にしかない。`nix develop --command cargo test --bin herdr <filter>`。
`--lib`は無い(binary crate)。フィルタ無しの全件実行は`pty::actor`付近でSIGPIPEにより中断するので
ゲートに使えない。モジュールフィルタ(`app::` / `ui::` / `detect::` / `config::`)で回す。
**合否は件数でなく失敗テスト名の集合の一致で判定する** — 詳細は
`claude/projects/-Users-s23159--config/memory/reference_herdr_test_non_hermetic.md`。

## 上流バージョンを上げる

flake inputのpin(タグ)を上げ、9本のパッチを新しい上流commitへ移植する手順。各手順の末尾に対応する検査段を添える。

1. 新しい上流タグからworktreeを作る。
2. 現行の開発branchのcommitを配備順にcherry-pickする。
3. conflictを解消してcommitする(段0)。
4. 開発ツリーで`cargo test`を走らせ、コンパイルと意味の両方を検査する。`nix/package.nix`は`doCheck = false`のため、`darwin-rebuild build`はテストを走らせない — ここが唯一のコンパイルゲートになる。取り込みの確認には`server::headless::tests`の件数(素の上流に暗転パッチ7本分が加わった件数)が使える。
5. UIパッチが描画へ影響していれば、`combined-frame-digest`の期待値を8パッチ適用後の実測値へ書き直す。
6. `git diff <sha>^ <sha> -- src/`でパッチファイルを再生成する。`-- src/`は`confirm-close-running-process`が変更する`docs/next/website/src/data/config-reference.json`を落とすために要る(nixのsrc filesetが`website/`を含まず、含めるとpatchPhaseが失敗する)。`git diff`の出力にコメントは含まれないため、コメントヘッダを持つパッチは手で戻す。
7. `flake.nix`のrefと`flake.lock`を新しい上流revへ更新する(段0)。
8. `home-manager/herdr.nix`のコメント中のbranch名を新しいbranch名へ更新する。
9. `darwin-rebuild build --flake ~/.config`で評価とビルドを検証する(sudo不要)。
10. `darwin-apply`で適用する(段2)。
11. `verify-herdr-deploy`で段0〜段2を確認する。
12. ユーザーがherdr serverを再起動する(段3)。
13. `herdr --skill > claude/skills/herdr/SKILL.md`で同梱skillを再生成し、日本語の追記節を戻す。

## 完了報告の規約

herdrのローカル変更を扱うタスクの完了報告には、「配備した」と書く代わりに`verify-herdr-deploy`の出力そのものを貼る。走らせずに完了と書く行為が、報告書式の欠落として見えるようにするためである。

## このスキルが防げないこと

段1と段2には自動で走る起動点が無い。`darwin-apply`の中から呼べるのは段3相当の確認だけである。コマンドを走らせなかった場合の誤認は、このスキル自体では防げない。段0から段2を走らせる動機は、上の完了報告の規約に依存する。

## 開発ツリーが両方向にずれうること

配備リスト(`home-manager/herdr.nix`の`patches`)と開発branchは独立にずれうる。現在は配備9本と開発ツリーの9commitが1対1で対応し、段1は差分なしで通る。`panel-contrast-fg-bright`は開発branchにも含めず、パッチファイルだけ未登録で据え置いている。次のupgradeで開発branchをrebase元にする前に、段1を通しておく。
