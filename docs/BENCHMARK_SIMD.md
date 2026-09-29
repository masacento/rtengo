# RTen WASM SIMD ベンチマーク

`rtenwasm` crate を SIMD (`simd128`) でビルドし、`wasm2go` の各家・フラグで変換して生成した Go バインディングの推論性能を比較した記録です。

## 目的

- RUST 側で `-C target-feature=+simd128` を有効にした RTen WASM が、`wasm2go` で変換後も期待どおり高速化するか検証する。
- `-fast-math` / `-simd-unroll` / `-fuse-loops` / `-pure` などの設定を変えて、性能への影響を比較する。
- 結論として、どの設定が実用的かを明らかにする。

## 環境

- CPU: Apple M1 Pro (arm64)
- Go: 1.24.x
- Rust: cargo 1.96.0
- target: `wasm32-wasip1`
- RTen: `rten 0.24.0`
- wasm2go: ローカルインストール版（`-fast-math` 等の SIMD フラグ対応）

## モデル・手順

- モデル: `sirasagi62_ruri-v3-30m-ONNX/model.onnx`（30M 埋め込みモデル、出力 256 次元）
- トークナイザ: 同ディレクトリの `tokenizer.json`
- 測定: `benchmark_test.go` の `BenchmarkEmbed`（`-benchmem -count N`）
  - `Short` = `"hello"`
  - `Long` = ある程度の長さの日本語テキスト
- 各設定で **埋め込み出力の一致**（`[0]==-1.129579`, `[4]==8.546244` 等）を事前確認してから測定。

## 設定一覧と結果

| # | RUST ビルド | wasm2go フラグ | Short (hello) | Long | 出力一致 |
|---|---|---|---|---|---|
| A | non-SIMD | asm（デフォルト） | **29.3 ms** | **~799 ms** | ✅ |
| B | **SIMD** | asm + `-fast-math` | 289 ms | 8266 ms | ✅ |
| C | **SIMD** | asm（`-fast-math` なし） | ~292–330 ms | ~8391–8456 ms | ✅ |
| D | **SIMD** | asm + `-fast-math -simd-unroll 4 -fuse-loops` | 290 ms | 8258 ms | ✅ |
| E | **SIMD** | `-pure`（pure-Go のみ） | ~289–292 ms | ~8443–8452 ms | ✅ |
| F | non-SIMD | `-pure`（pure-Go のみ） | **23.5–23.7 ms** | **586–733 ms** | ✅ |

（A は count=3、F は count=2、B/C/D/E は SIMD が遅いため count=1〜2。中央値で代表値を記載。）

## 補足データ（抜粋）

### A: non-SIMD + asm（デフォルト）
```
BenchmarkEmbed/Short-10   40  29292756 ns/op  56296 B/op  459 allocs/op
BenchmarkEmbed/Long-10     2  793452354 ns/op  772512 B/op  2627 allocs/op
```

### B: SIMD + asm + `-fast-math`
```
wasm2go: gcasm SIMD splice: 76538 call sites inlined, 0 kept as calls
BenchmarkEmbed/Short-10    4  289743323 ns/op  65556 B/op  459 allocs/op
BenchmarkEmbed/Long-10     1  8266676708 ns/op  772520 B/op  2627 allocs/op
```

### F: non-SIMD + `-pure`
```
BenchmarkEmbed/Short-10   49  23736758 ns/op  56296 B/op  459 allocs/op
BenchmarkEmbed/Long-10     2  586592730 ns/op  772512 B/op  2627 allocs/op
```

## 考察

1. **SIMD ビルドは ~10 倍遅い**
   - SIMD wasm（v128 命令）を wasm2go で変換したものは、バックエンド（asm / pure）やフラグ（`-fast-math` / `-simd-unroll 4` / `-fuse-loops`）に関係なく、**すべて ~10 倍遅い**（Short 約 290ms、Long 約 8.3s）。
   - 原因は、RTen が `simd128` で生成した v128 ベクトル命令を wasm2go が非効率にスカラ化するためと推定される。SIMD 化した Rust コードはスカラ実行だとむしろ遅くなる。

2. **`-fast-math` の影響はほぼ無い**
   - B（`-fast-math` あり）と C（なし）で差は誤差範囲。SIMD の大幅な劣化は `-fast-math` 由来ではない。

3. **SIMD チューニング（`-simd-unroll` / `-fuse-loops`）でも改善しない**
   - D（`-simd-unroll 4 -fuse-loops` + `-fast-math`）も同程度の ~10 倍遅さ。根本的な SIMD 変換の非効率は解消できない。

4. **non-SIMD + `-pure` がデフォルト（asm）より速い**
   - 非 SIMD 前提では、`-pure`（pure-Go バックエンド）が asm バックエンドより **約 15–25% 高速**（Short 23.5ms vs 29.3ms、Long ~660ms vs ~799ms）。
   - ただし `-pure` は wasm2go ヘルプ上「ベンチマーク用の ABIInternal 参照実装」とされており、asm バックエンドが本番想定。本件のワークロードでは asm スプライスが必ずしも最適でないことを示唆する。

## 結論 / 推奨

- **`rtenwasm` は SIMD (`+simd128`) でビルドしない。** 現行設定（SIMD 無効）のままが正しい。
  - 理由: wasm2go が v128 を非効率変換するため ~10 倍の性能劣化となる。README/Makefile の「SIMD は無効にすること」という注意は正しい。
- 生成フラグは現行のまま（デフォルト asm、`-fast-math` なし）で問題ない。
- 高速化を目指すなら、次点として **non-SIMD + `-pure`** が有効（約 15–25% 速い）。ただし `-pure` はベンチマーク参照実装の位置づけなので、本番採用する場合は出力・他プラットフォームでの挙動を追加確認のこと。

## 付録: 再現手順

```sh
# SIMD ビルド（非推奨、性能が ~10 倍落ちる）
cd rtenwasm
RUSTFLAGS="-C target-feature=+simd128" cargo build --release --target wasm32-wasip1
wasm2go -i target/wasm32-wasip1/release/rtenwasm.wasm \
  -out-dir ../internal/genwasm -pkg genwasm \
  -import github.com/masacento/rtengo/internal/genwasm \
  -fast-math

# 現行（推奨）: SIMD 無効 + asm
cd rtenwasm
cargo build --release --target wasm32-wasip1
wasm2go -i target/wasm32-wasip1/release/rtenwasm.wasm \
  -out-dir ../internal/genwasm -pkg genwasm \
  -import github.com/masacento/rtengo/internal/genwasm

# ベンチマーク
RTENGO_BENCH_MODEL=/path/to/model.onnx \
RTENGO_BENCH_TOKENIZER=/path/to/tokenizer.json \
go test -run '^$' -bench BenchmarkEmbed -benchmem -count 3
```