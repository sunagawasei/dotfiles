---
name: ebook
description: note.comのマガジン・連載をKindle向けの日本語EPUBにする。「noteの連載をkindleで読めるebookに」「記事をEPUBにして」で使う。文庫・新書の組版(段落冒頭1字下げ・括弧始まりは下げない・和文中の空白除去)をpandocのLuaフィルタとCSSで掛ける。
---

# note.com連載 → Kindle向け日本語EPUB

個人の閲覧用に限る。再配布しない。**無料公開の期間限定記事は取得できる間に取る**(2026-10-03の連載は「今週末限定」の無料公開だった)。

## 1. 記事一覧と本文の取得

`magazine_key`はURL `note.com/<user>/m/<magazine_key>` の末尾。

```bash
curl -sL "https://note.com/api/v1/magazines/<magazine_key>/notes" | jq -r '.data.notes[] | [.key,.publish_at,.price,.name]|@tsv'
curl -sL "https://note.com/api/v3/notes/<key>" > <key>.json   # .data.body がHTML、.data.can_read が読めるか
```

- 一覧は `v1` のmagazines API。`v3/magazines/...` は404
- `price`が0かつ`can_read`がtrueなら全文が取れる。有料記事は本文が途中で切れる
- 連続取得は1秒空ける。告知・宣伝の記事は除き、公開日順に並べる
- 各本文の先頭に `<h1>記事タイトル</h1>` を足して `ch/NN.html` にする

## 2. pandocでEPUB化

pandocは入っていない。`nix build nixpkgs#pandoc --no-link --print-out-paths` で得たパスの `bin/pandoc` を使う。

```bash
pandoc -f html -t epub3 --lua-filter=jp.lua --css=style.css --metadata-file=meta.yaml \
  --epub-cover-image=cover.jpg --toc --toc-depth=1 --split-level=1 -o out.epub ch/*.html
```

`meta.yaml`は `title` / `author` / `lang: ja`。表紙はマガジンの `cover_image_url`(横長バナーなので書影としては縦長にならない)。`jp.lua`と`style.css`はこのskillのディレクトリにある。

## 3. 検証(出力XHTMLを展開して数える)

```bash
for f in $(unzip -Z1 out.epub | grep text/ch); do unzip -p out.epub $f; done > all.xhtml
grep -c $'\xef\xbf\xbd' all.xhtml    # 文字化け(U+FFFD)は0
grep -c '<p></p>' all.xhtml          # 空段落は0
```

元の外部カード数(`embedded-service="external-article"`)と出力の `class="noindent card"` が一致すること。`Send to Kindle` に入れて実機で字下げ・画像幅・章ごとの改ページを見る。

## 落とし穴

- **Luaの文字クラスはバイト単位**: `[\u{3000}%s]` と書くと「さ」(E3 81 95)の先頭バイトまで削って文字化けする。全角空白の除去は `t:sub(1, 3) == "\u{3000}"` で3バイトずつ見る
- **タグ数はタグ名の境界まで見る**: `grep '<table'` は note の目次ウィジェット `<table-of-contents>`(中身は空)にも当たる。実際の表は0個だった
- 他記事の埋め込み(`embedded-service="note"`)は中身が空なので削除する。Amazonなどの外部カード(`external-article`)はタイトルのリンク1行だけ残す(`jp.lua`の`Figure`)
- 元HTMLの改行位置に空白が入る。和文どうしの`SoftBreak`は空白にしない(`jp.lua`の`Inlines`)
- Kindleは `line-break: strict` と `max-width` を無視する。禁則は端末任せにして、画像は `width:100%` で収める。`line-height` は端末設定で上書きされるが読みやすさを優先して残している
- 傍点は個別プロパティ(`text-emphasis-style`)で書く。2026-10-03の連載は本文に強調がなく、影響は出ていない
- 横書きのみ。縦書きは「p.152」「4,180円」などの縦中横の指定が要り、実機確認なしでは入れない

## 査読

2026-10-03にcodexが`jp.lua`と`style.css`を査読済み(カード抽出を再帰探索へ、傍点・画像のCSS、開き括弧の追加を反映。`<span>`で包まれた先頭の全角空白は残る)。Kindle Previewer/epubcheckは未実施。
