---
name: dockerfile
description: "Use when creating or modifying Dockerfiles. Apply security best practices with distroless images, multi-stage builds, and non-root users."
---

# Dockerfile ガイドライン

Dockerfileを作成・編集する際のベストプラクティス集。セキュリティと効率性を重視した構成を提供。

## ベースイメージ

- **distroless使用**: シェルやパッケージマネージャを排除し、攻撃対象領域を最小化

## ビルド構成

- **マルチステージビルド**: ビルド環境と実行環境を分離
- **レイヤーキャッシュ最適化**: 依存関係ファイル → ソースコードの順でCOPY

## セキュリティ

- **非rootユーザー実行**: `USER 65532` で実行
- **タグ固定**: latestタグは使用禁止（再現性確保）

## ファイル管理

- **.dockerignore**: 不要ファイルを除外（.git、ドキュメント等）

## コマンド設計

- **ENTRYPOINT**: メインコマンドを固定
- **CMD**: 可変引数を指定

## 品質管理

- **Hadolint**: リントツールで静的解析を実行
