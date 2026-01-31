# Cursor CLI 設定

Cursor CLIのカスタム設定ディレクトリです。

## ファイル構成

```
cursor/
├── README.md              # このファイル
├── cli-config.json        # Cursor CLIメイン設定
└── chats/                 # チャット履歴（自動生成）
```

## 設定概要

### エディタ設定

- **Vimモード**: 有効（`vimMode: true`）
- **行番号表示**: 有効（`showLineNumbers: true`）

### モデル設定

- **使用モデル**: GPT-5.2 Codex Extra High
- **maxMode**: 有効（最大コンテキスト長を使用）

### セキュリティ設定

- **承認モード**: allowlist（許可リスト方式）
- **サンドボックス**: 無効（`mode: disabled`）
- **ネットワークアクセス**: allowlist（許可リスト方式）
- **プライバシーモード**: Ghost Mode有効

### 権限設定

- **許可されたコマンド**: `Shell(ls)`のみ
- **コミット・PR署名**: Agentに帰属（`attributeCommitsToAgent: true`, `attributePRsToAgent: true`）

## カラーテーマ

Cursor CLIは現在、カスタムカラーテーマ設定をサポートしていません。
エディタUIの色は、ターミナルエミュレーター（WezTerm）のANSI色設定に従います。

### WezTermとの統合

このdotfilesリポジトリでは、WezTermで統一カラーテーマを設定しています：

- **背景色**: `#1A201E`（モノクロ背景）
- **前景色**: `#D7E2E1`（前景）
- **アクセント色**:
  - Cyan: `#5AAFAD`（成功・追加）
  - Magenta: `#8C83A3`（変更・警告）
- **選択背景**: `#2A2F2E`（濃い影）

詳細は `/Users/s23159/.config/COLOR-SYSTEM.md` を参照してください。

## 使用方法

### 起動

```bash
# 基本起動
cursor

# 特定のファイルを開く
cursor path/to/file.go

# ディレクトリを開く
cursor path/to/directory
```

### Vimキーバインド

Cursor CLIはVimモードが有効なため、標準的なVimキーバインドが使用できます：

- `i`: Insert モード
- `v`: Visual モード
- `ESC`: Normal モード
- `:w`: 保存
- `:q`: 終了
- `:wq`: 保存して終了

## 関連設定

- **WezTerm**: `/Users/s23159/.config/wezterm/`
- **Neovim**: `/Users/s23159/.config/nvim/`
- **カラーシステム**: `/Users/s23159/.config/COLOR-SYSTEM.md`

## 注意事項

- `cli-config.json`には認証情報（`authInfo`）が含まれています
- バージョン管理時は機密情報を除外するか、`.gitignore`に追加してください
- `statsigBootstrap`セクションは動的に生成されるため、手動編集は不要です
