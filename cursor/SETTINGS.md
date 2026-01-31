# Cursor CLI 設定ガイド

Cursor CLIの設定項目と推奨値の詳細ドキュメントです。

## 設定ファイル

### cli-config.json

Cursor CLIのメイン設定ファイル。JSON形式で以下のセクションを含みます：

```json
{
  "version": 1,
  "editor": {},
  "display": {},
  "model": {},
  "permissions": {},
  "sandbox": {},
  "attribution": {},
  "approvalMode": "",
  "maxMode": true,
  "network": {}
}
```

## エディタ設定（editor）

### vimMode

- **型**: `boolean`
- **デフォルト**: `false`
- **推奨値**: `true`
- **説明**: Vimキーバインドを有効化
- **効果**: Vim互換の操作が可能になり、キーボードのみでの効率的な編集が実現

```json
{
  "editor": {
    "vimMode": true
  }
}
```

### その他のエディタオプション

Cursor CLIは現在、以下のエディタ設定をサポートしていません：

- カスタムカラーテーマ（WezTermのANSI色設定を使用）
- フォント設定（ターミナルエミュレーターの設定を使用）
- タブサイズ（プロジェクト設定やEditorConfigに従う）

## 表示設定（display）

### showLineNumbers

- **型**: `boolean`
- **デフォルト**: `true`
- **推奨値**: `true`
- **説明**: 行番号を表示
- **効果**: コードの行位置が把握しやすくなり、デバッグやコードレビューが効率化

```json
{
  "display": {
    "showLineNumbers": true
  }
}
```

## モデル設定（model）

### modelId

- **型**: `string`
- **推奨値**: `"gpt-5.2-codex-xhigh"`
- **説明**: 使用するAIモデルのID
- **効果**: コード生成・補完の品質と速度を決定

### maxMode

- **型**: `boolean`
- **推奨値**: `true`
- **説明**: 最大コンテキスト長を使用
- **効果**: より長いコードコンテキストを理解し、精度の高い提案が可能

```json
{
  "model": {
    "modelId": "gpt-5.2-codex-xhigh",
    "maxMode": true
  },
  "maxMode": true
}
```

## セキュリティ設定

### approvalMode

- **型**: `string`
- **値**: `"allowlist"` | `"denylist"` | `"auto"`
- **推奨値**: `"allowlist"`
- **説明**: コマンド実行の承認モード
- **効果**: allowlistは明示的に許可されたコマンドのみ実行（最も安全）

### permissions

- **型**: `object`
- **説明**: 実行を許可/拒否するシェルコマンドのリスト

```json
{
  "permissions": {
    "allow": [
      "Shell(ls)",
      "Shell(cat)",
      "Shell(grep)",
      "Shell(find)"
    ],
    "deny": [
      "Shell(rm -rf)",
      "Shell(sudo)"
    ]
  }
}
```

**推奨される許可コマンド**:
- `Shell(ls)`: ディレクトリ一覧表示
- `Shell(cat)`: ファイル内容表示
- `Shell(grep)`: テキスト検索
- `Shell(find)`: ファイル検索
- `Shell(git)`: Gitコマンド

**推奨される拒否コマンド**:
- `Shell(rm -rf)`: 再帰的削除
- `Shell(sudo)`: 管理者権限実行
- `Shell(chmod)`: ファイル権限変更

### sandbox

- **型**: `object`
- **推奨値**:

```json
{
  "sandbox": {
    "mode": "disabled",
    "networkAccess": "allowlist"
  }
}
```

- **mode**: サンドボックスモード（`"disabled"` | `"enabled"`）
- **networkAccess**: ネットワークアクセス制御（`"allowlist"` | `"denylist"` | `"all"`）

## 帰属設定（attribution）

### attributeCommitsToAgent

- **型**: `boolean`
- **推奨値**: `true`
- **説明**: AIエージェントによるコミットを明示的に記録
- **効果**: `Co-Authored-By: Cursor AI`がコミットメッセージに追加される

### attributePRsToAgent

- **型**: `boolean`
- **推奨値**: `true`
- **説明**: AIエージェントによるPRを明示的に記録
- **効果**: PR作成者としてAIエージェントが記録される

```json
{
  "attribution": {
    "attributeCommitsToAgent": true,
    "attributePRsToAgent": true
  }
}
```

## プライバシー設定（privacyCache）

### ghostMode

- **型**: `boolean`
- **推奨値**: `true`（個人開発） | `false`（チーム開発）
- **説明**: チャット履歴をローカルのみに保存
- **効果**: AIとの対話がCursorのサーバーに送信されない

### privacyMode

- **型**: `number`
- **値**: `0`（無効） | `1`（部分） | `2`（完全）
- **推奨値**: `2`（最大プライバシー）
- **説明**: コードスニペットの送信を制御

```json
{
  "privacyCache": {
    "ghostMode": true,
    "privacyMode": 2
  }
}
```

## ネットワーク設定（network）

### useHttp1ForAgent

- **型**: `boolean`
- **推奨値**: `false`
- **説明**: HTTP/1.1を使用（HTTP/2の代わり）
- **効果**: 一部のプロキシ環境で互換性が向上するが、パフォーマンスは低下

```json
{
  "network": {
    "useHttp1ForAgent": false
  }
}
```

## 統合設定

### WezTermとの連携

Cursor CLIのカラーテーマはWezTermのANSI色設定に依存します。
統一テーマを実現するため、以下を確認してください：

1. WezTermのカラー設定（`wezterm.lua`）が正しく設定されている
2. ANSI 16色が統一カラーパレット（`COLOR-SYSTEM.md`）に準拠している
3. 背景・前景色がモノクロ基調である

### Gitとの連携

Cursor CLIでGitコミットを行う場合、以下の設定が推奨されます：

```json
{
  "permissions": {
    "allow": [
      "Shell(git)",
      "Shell(git status)",
      "Shell(git add)",
      "Shell(git commit)",
      "Shell(git push)"
    ]
  },
  "attribution": {
    "attributeCommitsToAgent": true
  }
}
```

## 推奨設定テンプレート

### 個人開発向け

```json
{
  "version": 1,
  "editor": {
    "vimMode": true
  },
  "display": {
    "showLineNumbers": true
  },
  "model": {
    "modelId": "gpt-5.2-codex-xhigh",
    "maxMode": true
  },
  "maxMode": true,
  "approvalMode": "allowlist",
  "permissions": {
    "allow": [
      "Shell(ls)",
      "Shell(cat)",
      "Shell(grep)",
      "Shell(find)",
      "Shell(git)"
    ],
    "deny": [
      "Shell(rm -rf)",
      "Shell(sudo)"
    ]
  },
  "sandbox": {
    "mode": "disabled",
    "networkAccess": "allowlist"
  },
  "attribution": {
    "attributeCommitsToAgent": true,
    "attributePRsToAgent": true
  },
  "privacyCache": {
    "ghostMode": true,
    "privacyMode": 2
  },
  "network": {
    "useHttp1ForAgent": false
  }
}
```

### チーム開発向け

```json
{
  "version": 1,
  "editor": {
    "vimMode": true
  },
  "display": {
    "showLineNumbers": true
  },
  "model": {
    "modelId": "gpt-5.2-codex-xhigh",
    "maxMode": true
  },
  "maxMode": true,
  "approvalMode": "allowlist",
  "permissions": {
    "allow": [
      "Shell(ls)",
      "Shell(cat)",
      "Shell(grep)",
      "Shell(find)",
      "Shell(git status)",
      "Shell(git diff)",
      "Shell(git log)"
    ],
    "deny": [
      "Shell(git push)",
      "Shell(rm -rf)",
      "Shell(sudo)"
    ]
  },
  "sandbox": {
    "mode": "enabled",
    "networkAccess": "allowlist"
  },
  "attribution": {
    "attributeCommitsToAgent": true,
    "attributePRsToAgent": true
  },
  "privacyCache": {
    "ghostMode": false,
    "privacyMode": 1
  },
  "network": {
    "useHttp1ForAgent": false
  }
}
```

## トラブルシューティング

### 設定が反映されない

1. Cursor CLIを再起動してください
2. 設定ファイルのJSON構文が正しいか確認してください（`jq`コマンドで検証可能）
3. 権限設定（`permissions`）が正しいか確認してください

### カラーテーマが適用されない

Cursor CLIは独自のカラーテーマをサポートしていません。
WezTermのANSI色設定を確認してください。

### Vimモードが動作しない

`cli-config.json`の`editor.vimMode`が`true`になっているか確認してください。
変更後、Cursor CLIを再起動してください。

## 関連ドキュメント

- **README**: `/Users/s23159/.config/cursor/README.md`
- **KEYMAPS**: `/Users/s23159/.config/cursor/KEYMAPS.md`
- **COLOR-SYSTEM**: `/Users/s23159/.config/COLOR-SYSTEM.md`
- **WezTerm設定**: `/Users/s23159/.config/wezterm/`
