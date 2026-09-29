# wazero から wasm2go への移行ガイド

このドキュメントは、Go アプリケーションで wazero を使って WebAssembly を実行している場合に、`wasm2go` に移行する方法を説明します。

## 移行を検討するケース

wazero から wasm2go への移行は、次のような場合に有効です。

- **コールドスタートを短縮したい** — wasm の JIT / AOT コンパイルを実行時にやめ、ビルド時に済ませたい
- **呼び出しオーバーヘッドを削りたい** — ホスト ↔ ゲストの境界をまたぐ処理が多い
- **実行時のメモリを減らしたい** — wasm エンジンの内部状態を持ちたくない
- **配布をシンプルにしたい** — 生成物が普通の Go パッケージなので、単一バイナリに埋め込みやすい

## アーキテクチャの違い

| | wazero | wasm2go |
|---|---|---|
| 実行方式 | 実行時に wasm をロードして JIT / インタプリタ実行 | ビルド時に wasm を Go コードに AOT 変換 |
| ランタイム依存 | `github.com/tetratelabs/wazero` を実行時に使う | 生成コードは Go 標準ライブラリのみ |
| 配布形態 | `.wasm` ファイルを実行時に読み込む | `.wasm` を含めた Go コードとして静的に配布 |
| ビルド時間 | 短い（wasm は実行時にコンパイル） | 長い（wasm2go のコード生成＋ Go ビルドが必要） |

## 移行手順

### 1. wasm2go をインストールする

```sh
go install github.com/goccy/wasm2go/cmd/wasm2go@latest
```

### 2. wasm を Go コードに変換する

wazero では実行時に `.wasm` を読み込んでいましたが、wasm2go ではビルド前にコード生成を行います。

```sh
wasm2go \
  -i module.wasm \
  -o ./internal/genwasm/module.go \
  -pkg genwasm \
  -import example.com/myproj/internal/genwasm
```

生成されるファイルは `-o` で指定した `.go` ファイルだけではありません。同じディレクトリに次のようなファイルが追加で出力されます。

- `amd64.s` / `arm64.s` — ネイティブアセンブリ
- `decls_amd64.go` / `decls_arm64.go` — 関数宣言
- `*_pure.go` — その他の GOARCH 用 pure-Go フォールバック
- `sharedimage*.go` — メモリ管理ヘルパー

これらはすべて同じ Go パッケージとして一緒にコミットしてください。

### 3. 呼び出しコードを書き換える

#### wazero の典型的なコード

```go
package main

import (
    "context"
    "os"

    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
    ctx := context.Background()

    // wasm バイナリを実行時に読み込む
    wasmBytes, _ := os.ReadFile("module.wasm")

    // ランタイムを作成
    r := wazero.NewRuntime(ctx)
    defer r.Close(ctx)

    // WASI を有効化
    wasi_snapshot_preview1.MustInstantiate(ctx, r)

    // モジュールをインスタンス化
    mod, err := r.Instantiate(ctx, wasmBytes)
    if err != nil {
        panic(err)
    }
    defer mod.Close(ctx)

    // エクスポート関数を呼ぶ
    result, err := mod.ExportedFunction("add").Call(ctx, 3, 5)
    if err != nil {
        panic(err)
    }
    println(result[0])
}
```

#### wasm2go のコード

```go
package main

import (
    "example.com/myproj/internal/genwasm"
)

func main() {
    // モジュールをインスタンス化（Go 関数の呼び出しだけ）
    m := genwasm.New()

    // エクスポート関数を呼ぶ（普通の Go メソッド）
    result := m.Add(3, 5)
    println(result)
}
```

## API 対応表

| wazero | wasm2go | 備考 |
|---|---|---|
| `wazero.NewRuntime(ctx)` | 不要 | wasm2go はランタイムを持たない |
| `wazero.NewRuntimeWithConfig(ctx, cfg)` | 不要 | 設定は生成時に決定される |
| `r.Instantiate(ctx, wasmBytes)` | `genwasm.New()` | 返り値は `*Module` |
| `r.InstantiateWithConfig(ctx, wasmBytes, cfg)` | `genwasm.New()` または `genwasm.NewWithMemory(...)` | メモリを外部から与えたい場合は `NewWithMemory` |
| `r.InstantiateWithConfig(ctx, wasmBytes, cfg)` | `genwasm.New()` または `genwasm.NewWithWASI(...)` | WASI をカスタマイズしたい場合は `NewWithWASI` |
| `mod.Close(ctx)` | 不要 | Go の GC に任せる。明示的な Close はない |
| `mod.ExportedFunction("add").Call(ctx, 3, 5)` | `m.Add(3, 5)` | 関数名は PascalCase に変換される |
| `mod.ExportedMemory("memory")` | `m.Memory()` | `[]byte` を直接返す |
| `memory.Read(offset, size)` | `m.Memory()[offset:offset+size]` | 直接スライスでアクセス |
| `memory.Write(offset, data)` | `copy(m.Memory()[offset:], data)` | 直接スライスでコピー |
| `wasi_snapshot_preview1.MustInstantiate(ctx, r)` | 自動生成 | `wasi_snapshot_preview1` の import は wasm2go がネイティブ実装を生成する |

## エクスポート名の変換ルール

wasm のエクスポート名は、Go のメソッド名として使えるように PascalCase に変換されます。

- `add` → `Add`
- `say_hello` → `SayHello`
- `run_query_v2` → `RunQueryV2`

## WASI の移行

wazero では WASI を使うために `wasi_snapshot_preview1.MustInstantiate(ctx, r)` が必要でした。

wasm2go では、wasm が `wasi_snapshot_preview1` を import している場合、生成時にネイティブ Go 実装が自動的に出力されます。ユーザーが明示的に初期化する必要はありません。

ただし、WASI のファイルシステムや環境変数などの挙動は、wasm2go の実装に依存します。wazero の `wazero.ModuleConfig` で細かく設定していた場合は、生成コード側で同様の対応が必要かどうかを確認してください。
### デフォルトの WASI 実装を使う

`wasi_snapshot_preview1` を import している wasm の場合、生成されるコンストラクタは `NewWithWASI` になり、`New()` はデフォルト実装を使う薄いラッパーになります。

```go
// デフォルト実装（stdin/stdout/stderr は OS のものを使う）
m := genwasm.New()
```

### WASI の挙動をカスタマイズする

`DefaultWASI()` の代わりに、生成される `WasiStubs` を自分で組み立てて渡せます。

```go
wasi := genwasm.DefaultWASI()
wasi.SetPreopenDir("/tmp")          // preopen ディレクトリを変更
wasi.SetArgs([]string{"prog", "-v"}) // argv を差し替える
wasi.SetEnv([]string{"KEY=VALUE"})   // 環境変数を差し替える

m := genwasm.NewWithWASI(wasi)
```

ファイルシステムを差し替えたい場合は `SetFS`、アクセス制御をしたい場合は `SetFSAccessHook` / `SetNetAccessHook` などのメソッドを確認してください。

## ホスト関数（カスタム import）の移行

wazero では `r.NewHostModuleBuilder("env")` などでホスト関数を提供できました。

wasm2go では、wasm の import が関数の場合、生成される `Module` に対応するインターフェースが定義され、Go 側で実装を注入できる形になります。
wasm2go では、wasm の import が関数の場合、生成される `Module` に対応するインターフェースが定義され、Go 側で実装を注入する形になります。

生成コードに出てくる `Env` のようなインターフェースを実装し、`NewWithImports` のようなコンストラクタがあればそこへ渡してください。具体的な生成形態は wasm の import 内容に依存するため、`wasm2go` 出力のインターフェース定義を確認してください。
### wazero の例

```go
hostModule := r.NewHostModuleBuilder("env").
    NewFunctionBuilder().
    WithFunc(func(v uint32) { fmt.Println(v) }).
    Export("log").
    Instantiate(ctx)

mod, _ := r.InstantiateWithConfig(ctx, wasmBytes,
    wazero.NewModuleConfig().WithName("env"))
```

### wasm2go の例

wasm が `"env"` モジュールから `log` を import している場合、生成コードには次のようなインターフェースが出力されます。

```go
type envImports interface {
    Log(m *Module, v uint32)
}
```

そしてコンストラクタはこのインターフェースを引数に取ります。

```go
type myEnv struct{}

func (e myEnv) Log(m *genwasm.Module, v uint32) {
    fmt.Println(v)
}

m := genwasm.New(myEnv{})
```

複数の import モジュールがある場合は、コンストラクタに複数の引数が並びます。生成コードの `type XxxImports interface` を確認し、それぞれ実装してください。

## メモリの扱い

wazero では `mod.Memory()` 経由でバッファを扱っていましたが、wasm2go では `m.Memory()` で直接 `[]byte` を返します。

```go
// wazero
mem := mod.Memory()
buf, _ := mem.Read(0, 16)

// wasm2go
buf := m.Memory()[0:16]
```

メモリサイズの変更（`memory.grow`）は生成コード内部で処理されます。外部からメモリを触る場合は、`accessMemory` に相当する排他制御が必要な場合があるため、生成コードのコメントを参照してください。

## エラーハンドリングの違い

wazero では `Call` が `error` を返しますが、wasm2go ではエクスポート関数は値だけを返します。

wasm の `unreachable` や `out of bounds` などのトラップは、wasm2go では Go の `panic` として通知されます。

```go
// wasm2go では panic として扱う
defer func() {
    if r := recover(); r != nil {
        fmt.Println("wasm trap:", r)
    }
}()
result := m.Run()
```

wazero の `error` ベースのエラーハンドリングから移行する場合は、`panic` / `recover` ベースに書き換える必要があります。

## ビルド・配布の違い

### wazero の場合

- `.wasm` ファイルを実行時に読み込む
- `go:embed` でバイナリに埋め込むことも可能
- 実行時に wasm のコンパイルが発生する

### wasm2go の場合

- `.wasm` はビルド時に `wasm2go` で Go コードに変換する
- 生成された Go コードを `go build` で普通にビルドする
- 実行時に `.wasm` ファイルは不要
- amd64 / arm64 以外では自動的に pure-Go フォールバックが使われる

### CI / ビルドスクリプトに組み込む

`Makefile` や CI スクリプトにコード生成を組み込むのが一般的です。

```makefile
gen-wasm:
	wasm2go -i module.wasm -o internal/genwasm/module.go -pkg genwasm -import example.com/myproj/internal/genwasm
```

## 移行チェックリスト

- [ ] `wasm2go` をインストールした
- [ ] 対象の `.wasm` を `wasm2go` で Go コードに変換した
- [ ] 生成された `.go` / `.s` / `decls_*.go` / `sharedimage*.go` をすべて同じディレクトリにコミットした
- [ ] wazero の `NewRuntime` / `Instantiate` を `genwasm.New()` に置き換えた
- [ ] `ExportedFunction(...).Call` を生成メソッド呼び出しに置き換えた
- [ ] メモリアクセスを `m.Memory()` スライスに置き換えた
- [ ] WASI 初期化コードを削除した（wasm2go が自動生成する場合）
- [ ] エラーハンドリングを `panic` / `recover` ベースに移行した
- [ ] wazero の依存を `go.mod` から削除した

## 制限事項

wasm2go はすべての wasm を変換できるわけではありません。以下は非対応または制限があります。

- SIMD（`v128`）命令
- Reference Types の完全なサポート
- Tail Calls（`return_call` など）
- Memory64
- 一部の新しい wasm プロポーザル

変換前に `wasm2go -i module.wasm -o /dev/null -pkg tmp -import tmp` で変換可能か確認することを推奨します。
