# Magic Cylinder - WebTransport Ping-Pong Experiment

このプロジェクトは、[quic-go/webtransport-go](https://github.com/quic-go/webtransport-go)を使用した実験的なWebTransportアプリケーションです。

## 概要

2つのサーバーインスタンスが互いにWebTransportを使ってpingpong通信を繰り返す実験です：

```
client -> server1 -> server2 -> server1 -> server2 -> ...
```

- **Client**: 初期pingをserver1に送信するだけ
- **Server1** (ポート 8443): リクエストを受け、レスポンスを返した後、server2にエコーを送信
- **Server2** (ポート 8444): リクエストを受け、レスポンスを返した後、server1にエコーを送信

サーバーコードは1つで、異なるポートとターゲットURLで2つのプロセスとして起動します。

## 必要な要件

- Go 1.21以上
- OpenSSL（自己署名証明書生成用）

## セットアップと実行

### 1. 依存関係のインストール

```bash
make deps
```

### 2. 自己署名証明書の生成

```bash
make certs
```

### 3. アプリケーションのビルド

```bash
make build
```

### 4. サーバーの実行

#### ターミナル1: Server1を起動 (ポート8443)

```bash
./bin/server -port 8443 -name server1 -target https://localhost:8444/webtransport
```

#### ターミナル2: Server2を起動 (ポート8444)

```bash
./bin/server -port 8444 -name server2 -target https://localhost:8443/webtransport
```

#### ターミナル3: Clientを実行

```bash
./bin/client -server https://localhost:8443/webtransport
```

## 動作確認

Clientが初期pingを送信すると、server1とserver2の間で自動的にpingpongが繰り返されます。

### Server1のログ例
```
=== server1 Starting ===
Server configuration: Port=8443, Name=server1, Target=https://localhost:8444/webtransport
[Controller] WebTransport connection established
[Controller] Received ping message: Initial ping from client (seq: 1)
[Repository] Processing ping: Initial ping from client (seq: 1)
[Controller] Sent pong message (seq: 1)
[Repository] Sending echo to target: https://localhost:8444/webtransport
[Repository] Echo sent to https://localhost:8444/webtransport
```

### Server2のログ例
```
=== server2 Starting ===
Server configuration: Port=8444, Name=server2, Target=https://localhost:8443/webtransport
[Controller] WebTransport connection established
[Controller] Received pong message (seq: 1)
[Repository] Processing pong: Pong response to: Initial ping from client (seq: 1)
[Controller] Sent ping message (seq: 2)
[Repository] Sending echo to target: https://localhost:8443/webtransport
```

## プロジェクト構造

```
magic-cylinder/
├── cmd/
│   ├── client/          # Clientアプリケーション
│   └── server/          # Serverアプリケーション（1つのコードで複数起動）
├── internal/
│   ├── base.go          # サーバー起動管理
│   ├── router.go        # ルーティングと依存性注入
│   ├── config/          # 設定管理
│   ├── controller/      # コントローラー層（ビジネスロジック）
│   ├── repository/      # リポジトリ層（通信ロジック）
│   └── entity/          # データモデル
├── certs/               # TLS証明書（生成後）
├── bin/                 # ビルド後のバイナリ（生成後）
└── Makefile             # ビルドとタスクの自動化
```

## 技術的な詳細

### WebTransport
- HTTP/3上で動作するWebTransport プロトコルを使用
- 双方向ストリーミング通信をサポート
- 自己署名証明書を使用（実験用途）

### アーキテクチャ
- **Interface-based Design**: 疎結合でテスタブルな実装
- **Controller層**: WebTransport接続管理とメッセージハンドリング
- **Repository層**: メッセージ処理とエコー送信
- **1つのエンドポイント**: `/webtransport` で統一

### コマンドラインオプション

#### Server
```bash
./bin/server -port <PORT> -name <NAME> -target <TARGET_URL>
```
- `-port`: サーバーのポート番号（デフォルト: 8443）
- `-name`: サーバーの名前（ログ用）
- `-target`: エコー先のサーバーURL（空の場合はエコーしない）

#### Client
```bash
./bin/client -server <SERVER_URL>
```
- `-server`: 接続先のサーバーURL（デフォルト: https://localhost:8443/webtransport）

## Makefileコマンド

```bash
make deps           # 依存関係のインストール
make certs          # 自己署名証明書の生成
make build          # ビルド
make clean          # ビルド成果物の削除
make test           # テストの実行
```

## 注意事項

- 自己署名証明書を使用しているため、証明書の警告が表示されます
- 実験用途のコードのため、本番環境での使用は推奨されません
- WebTransportはまだ実験的な技術です
- pingpongループは手動で停止する必要があります（Ctrl+C）