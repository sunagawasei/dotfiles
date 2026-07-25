---
name: split-commits
description: staged済みの大きな変更を論理単位(意思決定単位)のコミット列に分割する。hunk帰属の検分→中間スナップショット→git plumbing構築→tree一致検証。working tree無変更で安全
---

# コミット分割ワークフロー

## 概要

- 用途: レビュー可能性のため、staged(または未コミット)の大規模変更を「1 コミット = 1 意思決定」の列に分割する
- 原則: working tree には一切触れない(git plumbing で構築)。分割合計が元 diff と完全一致することを tree ハッシュで機械検証する

## ワークフロー

### 1. 論理単位の定義と hunk 帰属の検分

- 変更を論理単位 U1..Un(意思決定単位)として仮定義し、`git diff --cached` の全 hunk を各単位へ帰属判定する(大きい場合は codex-research 等の調査役へ委譲: ファイル別テーブル/A=ファイル丸ごと単一単位・B=hunk 分割必要・C=同一 hunk 内混在の分類/依存シンボルの静的判定/推奨分割数)
- 分割限度の判断基準: (a)各中間コミットが単独でコンパイル通過するか(シンボル参照・import) (b)GitOps 環境では中間 revision が deploy されうるため「その revision が同期されても壊れないか」(実例: chart フォーク削除と代替注入の実装は runtime 意味論で併合必須だった) (c)テストの統合 assertion が複数単位を 1 つの ConsistOf で検証していると分割不能 → コンパイル整合に必要かつ意味的帰属が一致する行だけ前倒しする
- テスト実行の全 pass は最終コミットのみで保証し、中間は build/vet 通過まで、という妥協が現実的

### 2. 保険と中間スナップショット構築

- `git diff --cached > full-staged.patch` を 2 箇所に保存(復元保険)
- 「commit k 適用後」のファイル実体(HEAD から変化するファイルのみ、repo 相対パス)を stage1..stageN-1 として構築(実装役 codex-impl 等へ委譲可。混入自己チェックと各 stage の GOWORK=off go build/vet 等のコンパイル検証を必須にする)
- 検収: 本体 git status/log の無傷確認、stage への他単位混入を grep でスポットチェック、コンパイル検証のスポット再現

### 3. tree 構築と stat 提示(commit はまだ作らない)

```bash
export GIT_INDEX_FILE="$TMPDIR/split-idx"
git read-tree HEAD
for f in $(cd stage1 && find . -type f | sed 's|^\./||'); do
  h=$(git hash-object -w "stage1/$f") && git update-index --add --cacheinfo "100644,$h,$f"
done
tree1=$(git write-tree)
# stage2 以降も同様に積む
unset GIT_INDEX_FILE
tree3=$(git write-tree)   # 本体 index から最終 tree
git diff-tree --stat HEAD $tree1 && git diff-tree --stat $tree1 $tree2 && git diff-tree --stat $tree2 $tree3
git diff-tree --stat HEAD $tree3   # 合計整合: 元 staged diff の stat と一致するはず
```

- 一時 index(GIT_INDEX_FILE)を使うので本体 index・working tree は無傷

### 4. コミットメッセージ(ユーザーと確定してから)

- **タイトルは手段でなく目的/対象システムを主語にする**(悪い例: "Switch Harbor to PVC storage" → 良い例: "Set up the apne1 stg environment with ..."。PR タイトルも同じ原則)
- **社内固有名詞はタイトルで避ける**。本文初出で「一般名詞 (固有名詞)」形式で一度定義し、以後はコード上のシンボルとの対応が取れる固有名詞を使ってよい(実例: "the in-house S3-compatible object storage (S4)")
- repo の既存コミットの言語・prefix 慣習に合わせる。auto-close したくない issue は Closes でなく Refs
- **チーム向け成果物(コミットメッセージ・PR・issue)に書き手のローカル文脈を書かない**: 「現在の作業ブランチにある〜は削除する」のような書き手の作業状態・未来形は読者に通じない。対象が何かの定義と、完了済みの事実を書く
- 手順・段取りを書くときは一文に押し込まず番号リストにする

### 5. GO 後の commit 構築と検証

```bash
c1=$(git commit-tree $tree1 -p HEAD -F msg1.txt)
c2=$(git commit-tree $tree2 -p $c1 -F msg2.txt)
c3=$(git commit-tree $tree3 -p $c2 -F msg3.txt)
git update-ref -m "split staged changes into logical commits" refs/heads/<branch> $c3
```

- update-ref 後、index の中身が最終 tree と一致していれば git status は自然にクリーンになる
- 検証: `git status` クリーン / `git diff HEAD` 空 / 最終状態でテスト実行
- 復元: `git update-ref refs/heads/<branch> <旧HEAD>` だけで「未コミット・全量 staged」に戻る(index は独立)

## 注意点

- 途中で commit を作らない(勝手 commit 禁止の運用と衝突する)。commit 構築は必ずユーザーの明示 GO 後
- go.work.sum 等の checksum ファイルは「不足はエラー・余分は無害」なので早いコミットに随伴させるのが安全側
- 分割後の各コミットは bisect・revert の単位になる。無理な分割(中間状態が壊れる)はしない — 分割不能なら 1 コミット+PR body の Review order 案内で代替する

## 関連

- 検分・スナップショット構築の委譲は `/orchestrate-agents` の [research]/[implement] パケット書式を使う
