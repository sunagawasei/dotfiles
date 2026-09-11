---
name: herdr-deploy
description: "herdrのローカルパッチを実装し配備するときの手順。パッチファイル(home-manager/patches/)の作成、herdr.nixへの登録、packages.nixとの関係、darwin-applyによる反映、herdr serverの再起動までを扱う。herdrの挙動を変える作業、パッチの追加や変更、herdr.nixやpackages.nixの編集、配備手順の確認に入ったときに使う。"
---

# herdrのローカル変更を配備する

herdrは flake input のソースへローカルパッチを当ててビルドする。**commitしただけでは稼働バイナリは変わらない。** パッチファイル化、登録、`darwin-apply`、herdr serverの再起動を経て初めて反映される。

## 配備の経路

各手順の末尾に、対応する検査コマンドの段番号(段0〜段3の定義は後述)を添える。

1. 開発ツリー(git worktree、branch `v080-upgrade`)で実装してcommitする(段0)。
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

## 完了報告の規約

herdrのローカル変更を扱うタスクの完了報告には、「配備した」と書く代わりに`verify-herdr-deploy`の出力そのものを貼る。走らせずに完了と書く行為が、報告書式の欠落として見えるようにするためである。

## このスキルが防げないこと

段1と段2には自動で走る起動点が無い。`darwin-apply`の中から呼べるのは段3相当の確認だけである。コマンドを走らせなかった場合の誤認は、このスキル自体では防げない。段0から段2を走らせる動機は、上の完了報告の規約に依存する。

## 開発ツリーが両方向にずれうること

配備リスト(`home-manager/herdr.nix`の`patches`)と開発branchは独立にずれる。2026-09-11時点では、非アクティブpaneの暗転パッチが配備側に登録されている一方で開発ツリーに対応するcommitが無く、`panel-contrast-fg-bright`は開発ツリーにcommitがある一方で配備リストからは外れている(2026-09-02のcommit `e138f54`で意図的に外した)。段1はこの両方向のずれを同時に検出する。次のupgradeで開発branchをrebase元にする前に、段1を通しておく。
