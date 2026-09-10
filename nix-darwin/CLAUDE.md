# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 適用コマンド

設定変更を反映するには、フレークルート (`~/.config`) から実行：

```bash
darwin-apply
```

`darwin-apply`（`../home-manager/packages.nix`）は `sudo darwin-rebuild switch --flake ~/.config`
のラッパー。flake attrを省略しているため hostname（`scutil --get LocalHostName`）で解決される。
**Claudeもこれを実行できる**: permission ruleは deny→ask→allow の順で評価され specificity が
効かないため、`Bash(sudo:*)` deny を残したまま sudo を含む形で例外は作れない。sudo を含まない
コマンド名にして `Bash(darwin-apply)` のみ allow することで、他の sudo は一律拒否のまま維持している。
引数は受け取らない（渡すと exit 2）。sudoers 側の NOPASSWD は `configuration.nix` の
`security.sudo.extraConfig` で呼び出し形の完全一致のみ許可している。

`zsh.nix` のエイリアス：
- `nupdate` — `nix flake update` + `darwin-apply`（パッケージ更新 + 適用）

> home-manager は nix-darwin の darwinModule として統合されているため、単独の `home-manager switch` は使わない。

## パッケージ追加の判断基準

- **Nix パッケージ（一般ツール・言語）** → `../home-manager/` 配下の適切なモジュール（詳細は `../home-manager/CLAUDE.md`）
- **Homebrew cask（GUI アプリ）** → `homebrew.nix` の `casks` リスト
- **Homebrew brew（nixpkgs にない CLI）** → `homebrew.nix` の `brews` リスト

## 注意点

- **Homebrew GitHub トークン**: `homebrew.nix` には `activationScript` のオーバーライドがある。`nix-darwin` の activation は `#!/usr/bin/env -i bash` で環境変数を消去するため、`HOMEBREW_GITHUB_API_TOKEN` を直接渡せない。`gh auth token` をユーザーセッションから動的取得して注入する回避策を実装済み（触らない）。
- **`nix.enable = false`**: Nix デーモン管理は nix-darwin に委ねず別管理。
- **アーキテクチャ**: `aarch64-darwin`（Apple Silicon）固定。
- **`darwinConfigurations` の attr 名**: hostname と一致させる。PC 交換時は `flake.nix` 側も改名する。

## Karabiner-Elements(尊師スタイルの on/off)

「尊師スタイル」は MacBook 内蔵キーボードの上に roBa(ZMK・BLE)を載せる運用。内蔵キーが roBa の
底面で押されるのを防ぐため、Karabiner-Elements の profile を切り替えて内蔵キーボードを無効化する。

```bash
sonshi on      # 内蔵キーボードのキーを常に無効にする(profile: sonshi)
sonshi off     # 内蔵キーボードを使う(profile: normal)
sonshi status  # 選択中の profile を表示する(引数なしも同じ)
```

`sonshi` は `../home-manager/packages.nix` の `writeShellApplication`。profile を選んだあと
`karabiner_cli --show-current-profile-name` で読み戻し、期待した名前でなければ非0で終了する。
`--select-profile` は profile が存在しなくても終了コード0を返すため、この読み戻しが無いと
「on にしたつもりで内蔵が生きている」状態に気づけない。

**`sonshi status` が保証するのは profile 名だけで、profile の中身は検査しない。** GUI から
ルールを消しても `sonshi` と表示される。中身の検査は `[task:sonshi-doctor]` で別途扱う。

### roBa が使えないときの戻し方

`sonshi on` の状態では roBa の接続に関係なく内蔵キーボードが効かない。roBa の電池切れ・BLE 不調・
別 Mac へのペアリング切り替えで roBa が使えなくなると、内蔵キーボードで `sonshi off` を打てない。
キーボードを使わない復旧経路は3つ。

1. Finder → アプリケーション → Karabiner-Elements を開き、Profiles タブで `normal` を選ぶ
   (トラックパッドだけで完結する)
2. roBa を再接続する

**メニューバーの Karabiner アイコンはあてにしない。** このマシンはメニューバーの項目が多く、
Karabiner のアイコンが MacBook のノッチの裏へ回り込んで見えない。アイコンが見えているときは
そこから Profiles を選んでもよいが、復旧経路としては数えない。

**Settings ウィンドウを閉じても終了しても内蔵キーは戻らない。** キー入力を掴んでいるのは root で
動く `Karabiner-Core-Service` と `Karabiner-Console-User-Server` で、Settings アプリ
(`Karabiner-Elements.app`)はそれとは別プロセス。Settings が起動していなくても profile は効いている。

### 設定本体

`~/.config/karabiner/` は .gitignore 対象で git 追跡外。GUI が操作のたびに書き換えるうえ BLE
デバイスの識別子を含むため。追跡しない代わりに、profile 2本の構成をここに記録する。

profile `sonshi`:

- `complex_modifications` に1ルール。`device_if` の `is_built_in_keyboard: true` を条件に、
  `from.any` が `key_code` / `consumer_key_code` / `apple_vendor_keyboard_key_code` /
  `apple_vendor_top_case_key_code` の manipulator 4本。各 `from` に `modifiers.optional: ["any"]`
  を付け、修飾キー保持中の入力も対象にする。`to` は書かずイベントを捨てる
- `pointing_button` は含めない。内蔵トラックパッドを生かすため
- `devices` に roBa(vendor_id 7504 / product_id 24926)の `disable_built_in_keyboard_if_exists: true`
  を残す。上のルールの `from.any` に漏れがあっても roBa 接続中はデバイス設定側で落ちる

profile `normal`:

- `complex_modifications` なし。roBa の `disable_built_in_keyboard_if_exists` は設定しない(既定の `false`)

両 profile 共通:

- roBa の `devices` エントリに `ignore: true`(GUI の "Modify events" off)。roBa は keyboard と
  pointing が統合された1デバイスなので、これを外すと PMW3610 トラックボールのポインタ入力が
  Karabiner の仮想 HID を経由する
- 内蔵キーボードの `devices` エントリは作らない。既定の "Modify events" on のままにする必要がある
  (Karabiner がデバイスを grab しないとキー入力を落とせない)。`ignore: true` を付けると `sonshi`
  profile のルールごと効かなくなる
- `virtual_hid_keyboard.keyboard_type_v2: "jis"`

`karabiner.json` の全体は次のとおり(`selected` は選択中の profile 側が `true`)。**Karabiner は
保存のたびにキーを並べ替え、既定値のフィールドを落とす。** `selected: false` や
`disable_built_in_keyboard_if_exists: false` を手で書いても消えるので、下は保存後の形。

```json
{
  "profiles": [
    {
      "complex_modifications": {
        "rules": [
          {
            "description": "尊師モード: 内蔵キーボードのキー入力を捨てる",
            "manipulators": [
              {
                "conditions": [
                  {
                    "identifiers": [
                      {
                        "is_built_in_keyboard": true
                      }
                    ],
                    "type": "device_if"
                  }
                ],
                "from": {
                  "any": "key_code",
                  "modifiers": {
                    "optional": [
                      "any"
                    ]
                  }
                },
                "type": "basic"
              },
              {
                "conditions": [
                  {
                    "identifiers": [
                      {
                        "is_built_in_keyboard": true
                      }
                    ],
                    "type": "device_if"
                  }
                ],
                "from": {
                  "any": "consumer_key_code",
                  "modifiers": {
                    "optional": [
                      "any"
                    ]
                  }
                },
                "type": "basic"
              },
              {
                "conditions": [
                  {
                    "identifiers": [
                      {
                        "is_built_in_keyboard": true
                      }
                    ],
                    "type": "device_if"
                  }
                ],
                "from": {
                  "any": "apple_vendor_keyboard_key_code",
                  "modifiers": {
                    "optional": [
                      "any"
                    ]
                  }
                },
                "type": "basic"
              },
              {
                "conditions": [
                  {
                    "identifiers": [
                      {
                        "is_built_in_keyboard": true
                      }
                    ],
                    "type": "device_if"
                  }
                ],
                "from": {
                  "any": "apple_vendor_top_case_key_code",
                  "modifiers": {
                    "optional": [
                      "any"
                    ]
                  }
                },
                "type": "basic"
              }
            ]
          }
        ]
      },
      "devices": [
        {
          "disable_built_in_keyboard_if_exists": true,
          "identifiers": {
            "is_keyboard": true,
            "is_pointing_device": true,
            "product_id": 24926,
            "vendor_id": 7504
          },
          "ignore": true
        }
      ],
      "name": "sonshi",
      "virtual_hid_keyboard": {
        "keyboard_type_v2": "jis"
      }
    },
    {
      "devices": [
        {
          "identifiers": {
            "is_keyboard": true,
            "is_pointing_device": true,
            "product_id": 24926,
            "vendor_id": 7504
          },
          "ignore": true
        }
      ],
      "name": "normal",
      "selected": true,
      "virtual_hid_keyboard": {
        "keyboard_type_v2": "jis"
      }
    }
  ]
}
```

Expert タブの `enable_cgeventtap_fallback` は off のまま維持する。on にするとイベントの発生元
デバイスを判別できず、内蔵キーボードだけを対象にできない。

`karabiner_cli` は `/Library/Application Support/org.pqrs/Karabiner-Elements/bin/karabiner_cli`。
`--lint-complex-modifications <file>` は未知の `any` 値も未知の識別子キーも実際に弾くので、ルールを
足すときの検算に使える。

### 設定を書き換えるとき

Karabiner は `karabiner.json` の外部からの変更を検知して自動で読み直す(再起動は不要)。**壊れた JSON を
保存すると、こちらが検証するより先に daemon が読み込む。** そのため live のファイルを直接編集せず、
同じディレクトリに候補を作って検証し、`mv` で置き換える。同一ディレクトリの `mv` は rename なので、
`karabiner.json` は常に旧版か新版のどちらかで、途中の状態が見えることはない。GUI の Settings ウィンドウを
開いたまま置き換えると書き戻される可能性があるので、閉じてから行う。

```bash
d=~/.config/karabiner
t=$(mktemp -d)
cp "$d/karabiner.json" "$d/karabiner.json.bak" &&
  cp "$d/karabiner.json" "$d/karabiner.json.new" &&
  "${EDITOR:-vim}" "$d/karabiner.json.new" &&
  python3 -m json.tool "$d/karabiner.json.new" > /dev/null &&
  python3 -c 'import json,sys; c=json.load(open(sys.argv[1])); json.dump({"title":"lint","rules":[r for p in c["profiles"] for r in p.get("complex_modifications",{}).get("rules",[])]}, open(sys.argv[2],"w"))' \
    "$d/karabiner.json.new" "$t/rules.json" &&
  karabiner_cli --lint-complex-modifications "$t/rules.json" &&
  mv "$d/karabiner.json.new" "$d/karabiner.json" &&
  karabiner_cli --list-profile-names
```

gate は2段ある。`python3 -m json.tool` が JSON 構文を見て、`--lint-complex-modifications` が
Karabiner のルールとしての妥当性を見る(未知の `any` の値も、`identifiers` の未知のキーも exit 1 で
弾く)。`--lint-complex-modifications` は `karabiner.json` の形を受け付けないので、その手前の `python3 -c`
で全 profile のルールを `{"title", "rules"}` 形に抜き出している。`&&` で繋いであるので、どちらかで
落ちれば `mv` に到達せず live は元のまま残る。戻すときも
`mv "$d/karabiner.json.bak" "$d/karabiner.json"` で置き換える(こちらも rename)。内蔵キーが効かず
roBa も使えない状態になったら、上の「roBa が使えないときの戻し方」の経路で `normal` へ戻す。

`karabiner_cli` は cask が `/opt/homebrew/bin/karabiner_cli` に symlink を張るので PATH で引ける。
実体は `/Library/Application Support/org.pqrs/Karabiner-Elements/bin/karabiner_cli`。`sonshi` ラッパーは
Nix 環境から呼ばれて PATH に依存できないため、そちらは実体の絶対パスを直接使っている。

### 選択の永続性

profile の選択は `karabiner.json` の `selected` に書かれるので、ログイン後は macOS を再起動しても
残る。ログインウィンドウ(ログイン前)と Karabiner 起動前の短時間は対象外。
