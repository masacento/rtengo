# Benchmarks

wazero → wasm2go 移行の前後比較用ベンチマーク記録。

ベンチマーク本体は `benchmark_test.go`。モデルとトークナイザは環境変数で指定する。

```sh
RTENGO_BENCH_MODEL=/path/to/model.onnx \
RTENGO_BENCH_TOKENIZER=/path/to/tokenizer.json \
go test -run '^$' -bench . -benchmem -count 5
```

`-tags rtengopure` を付けると wasm2go の pure-Go バックエンド
（`internal/genwasm_pure`）に切り替わる。asm との比較は本ドキュメント末尾の
「asm バックエンド vs pure-Go バックエンド」セクション参照。

## 計測対象

| ベンチマーク | 内容 |
|---|---|
| `BenchmarkRuntimeInit` | WASM コンパイル + インスタンス化 + `rten_init`（コールドスタート相当） |
| `BenchmarkModelLoad` | モデルを WASM メモリへコピーして `rten_load_model` を実行 |
| `BenchmarkEmbed/Short` | 短いテキスト（`hello`）の埋め込み 1 回（正規化あり） |
| `BenchmarkEmbed/Long` | 長めの日本語テキストの埋め込み 1 回（正規化あり） |

## ベースライン: wazero 版（2026-07-23）

- commit: `3fdc76c`
- 実装: wazero v1.10.1（実行時ロード + JIT 実行）
- 環境: Apple M1 Pro / 32GB / darwin arm64 / go1.26.1
- モデル: ruri `model.onnx`（504MB）、`tokenizer.json`（6.4MB）
- コマンド: `go test -run '^$' -bench . -benchmem -count 5`

| ベンチマーク | 時間/op（5 回中央値） | メモリ/op | allocs/op |
|---|---|---|---|
| RuntimeInit | ~789 ms | 42.5 MB | ~59,234 |
| ModelLoad | ~1.14 s（574ms〜1.52s とばらつき大） | ~4.0 GB | 16 |
| Embed/Short | ~88.6 ms | 429 KB | 568 |
| Embed/Long | ~1.57 s | 1.15 MB | 2,734 |

生データ:

```
BenchmarkRuntimeInit-10    	       2	 781953062 ns/op	42511936 B/op	   59236 allocs/op
BenchmarkRuntimeInit-10    	       2	 791136521 ns/op	42509356 B/op	   59235 allocs/op
BenchmarkRuntimeInit-10    	       2	 794825208 ns/op	42509348 B/op	   59234 allocs/op
BenchmarkRuntimeInit-10    	       2	 789447208 ns/op	42509220 B/op	   59233 allocs/op
BenchmarkRuntimeInit-10    	       2	 788849542 ns/op	42509340 B/op	   59234 allocs/op
BenchmarkModelLoad-10      	       1	1524013042 ns/op	3442236880 B/op	      17 allocs/op
BenchmarkModelLoad-10      	       2	1141785562 ns/op	3996200400 B/op	      16 allocs/op
BenchmarkModelLoad-10      	       2	 574873708 ns/op	3996200400 B/op	      16 allocs/op
BenchmarkModelLoad-10      	       2	 628416875 ns/op	3996200456 B/op	      17 allocs/op
BenchmarkModelLoad-10      	       4	1354297114 ns/op	3775518160 B/op	      15 allocs/op
BenchmarkEmbed/Short-10    	      13	  88606500 ns/op	  428840 B/op	     568 allocs/op
BenchmarkEmbed/Short-10    	      13	  87173042 ns/op	  428840 B/op	     568 allocs/op
BenchmarkEmbed/Short-10    	      12	  86952278 ns/op	  428840 B/op	     568 allocs/op
BenchmarkEmbed/Short-10    	      13	  91680356 ns/op	  428840 B/op	     568 allocs/op
BenchmarkEmbed/Short-10    	      12	  91075677 ns/op	  431926 B/op	     568 allocs/op
BenchmarkEmbed/Long-10     	       1	1576196417 ns/op	 1147048 B/op	    2734 allocs/op
BenchmarkEmbed/Long-10     	       1	1526584833 ns/op	 1147048 B/op	    2734 allocs/op
BenchmarkEmbed/Long-10     	       1	1598960166 ns/op	 1147048 B/op	    2734 allocs/op
BenchmarkEmbed/Long-10     	       1	1581362875 ns/op	 1147048 B/op	    2734 allocs/op
BenchmarkEmbed/Long-10     	       1	1540437375 ns/op	 1147048 B/op	    2734 allocs/op
```

## wasm2go 版（2026-07-23）

- 実装: wasm2go v0.4.9 による AOT 変換（`internal/genwasm`）。wazero 依存は削除
- 注意: wasm2go は SIMD（v128）非対応のため、wasm は `+simd128` なしでリビルドした。推論カーネルはスカラー版になる
- 環境: Apple M1 Pro / 32GB / darwin arm64 / go1.26.1（ベースラインと同一）
- モデル: ruri `model.onnx`（504MB）、`tokenizer.json`（6.4MB）
- コマンド: `go test -run '^$' -bench . -benchmem -count 5`

| ベンチマーク | 時間/op（5 回中央値） | メモリ/op | allocs/op |
|---|---|---|---|
| RuntimeInit | ~0.11 ms | 1.6 MB | 13 |
| ModelLoad | ~60 ms（初回のみ ~1.1 s） | ~210 MB | 4 |
| Embed/Short | ~148 ms | 70 KB | 459 |
| Embed/Long | ~3.41 s | 1.2 MB | 2,630 |

生データ:

```
BenchmarkRuntimeInit-10    	   10000	    114529 ns/op	 1606699 B/op	      13 allocs/op
BenchmarkRuntimeInit-10    	    9847	    108901 ns/op	 1606694 B/op	      13 allocs/op
BenchmarkRuntimeInit-10    	   10000	    111249 ns/op	 1606698 B/op	      13 allocs/op
BenchmarkRuntimeInit-10    	   10000	    103707 ns/op	 1606698 B/op	      13 allocs/op
BenchmarkRuntimeInit-10    	   10000	    100644 ns/op	 1606700 B/op	      13 allocs/op
BenchmarkModelLoad-10      	       1	1097040416 ns/op	3711303840 B/op	       8 allocs/op
BenchmarkModelLoad-10      	      16	  66005745 ns/op	231956528 B/op	       4 allocs/op
BenchmarkModelLoad-10      	      16	  63693391 ns/op	231956521 B/op	       4 allocs/op
BenchmarkModelLoad-10      	      19	  56469024 ns/op	195331813 B/op	       4 allocs/op
BenchmarkModelLoad-10      	      18	  60312410 ns/op	206183584 B/op	       4 allocs/op
BenchmarkEmbed/Short-10    	       7	 153486446 ns/op	   69608 B/op	     459 allocs/op
BenchmarkEmbed/Short-10    	       7	 150933476 ns/op	   80190 B/op	     459 allocs/op
BenchmarkEmbed/Short-10    	       7	 147726690 ns/op	   69608 B/op	     459 allocs/op
BenchmarkEmbed/Short-10    	       7	 146118321 ns/op	   74899 B/op	     459 allocs/op
BenchmarkEmbed/Short-10    	       7	 148266869 ns/op	   69608 B/op	     459 allocs/op
BenchmarkEmbed/Long-10     	       1	3418209750 ns/op	 1220184 B/op	    2630 allocs/op
BenchmarkEmbed/Long-10     	       1	3402867208 ns/op	 1183144 B/op	    2627 allocs/op
BenchmarkEmbed/Long-10     	       1	3492674375 ns/op	 1220184 B/op	    2630 allocs/op
BenchmarkEmbed/Long-10     	       1	3406336083 ns/op	 1183144 B/op	    2627 allocs/op
BenchmarkEmbed/Long-10     	       1	3397484750 ns/op	 1220184 B/op	    2630 allocs/op
```

## 比較サマリ（wazero → wasm2go）

| ベンチマーク | wazero | wasm2go | 倍率 |
|---|---|---|---|
| RuntimeInit | 789 ms | 0.11 ms | **約 7,000 倍高速** |
| ModelLoad | ~1.1 s / 4.0 GB | 60 ms / 210 MB | **約 18 倍高速・メモリ 1/19** |
| Embed/Short | 89 ms | 148 ms | 1.7 倍低速 |
| Embed/Long | 1.57 s | 3.41 s | 2.2 倍低速 |

所感:

- コールドスタート（RuntimeInit）とモデルロードは劇的に改善。実行時コンパイルが消え、メモリも wasm エンジン分だけ削減された
- 推論（Embed）の低速化は SIMD 無効化が原因。wazero 版は `+simd128` ビルドを JIT 実行していたが、wasm2go は v128 非対応のためスカラーカーネルになる。トレードオフとして受け入れるかは用途次第
- 数値の正しさは確認済み（`hello` の埋め込みベクトルが新旧実装で 1e-6 オーダーで一致）

## asm バックエンド vs pure-Go バックエンド（2026-07-24）

`wasm2go` は `-pure` フラグで asm バンドル（`amd64.s` / `arm64.s`）を出力しない
pure-Go の参照実装を生成できる（`the ABIInternal reference for benchmarking`）。
両者を比較するため、同一 wasm から 2 つの生成パッケージを作って切り替え可能にした。

- 生成物:
  - asm 版 …… 従来の `internal/genwasm`（gcasm バックエンド、`.s` + arch ビルドタグ付き）
  - pure 版 …… `internal/genwasm_pure`（`wasm2go -pure` 出力、arch タグ・`.s` なし）
- 切り替え: `internal/rtenwasm` を `rtengopure` ビルドタグで切り替え。公開 API（`rtengo.NewRuntime` 等）はそのまま。

```sh
# asm（デフォルト）
go test -run '^$' -bench . -benchmem -count 5

# pure-Go
go test -tags rtengopure -run '^$' -bench . -benchmem -count 5
```

> 注意: このセクションは **Apple M3** + **ruri-v3-30m**（`model.onnx` 140MB /
> `tokenizer.json` 6.4MB）で計測した。メインの wazero/wasm2go 表とは
> マシン（M1 Pro）もモデル（504MB ruri）も異なるため、**絶対値ではなく
> asm vs pure の比だけが意味を持つ**。

### 計測結果（5 回中央値）

| ベンチマーク | asm（gcasm） | pure-Go | pure/asm |
|---|---|---|---|
| RuntimeInit | 73.9 µs / 1.53 MB / 13 allocs | 64.5 µs / 1.45 MB / 13 allocs | **0.87×（pure が高速）** |
| ModelLoad | 17.3 ms / 15.0 MB / 4 allocs | 16.6 ms / 14.3 MB / 4 allocs | **0.96×（pure が高速）** |
| Embed/Short | 15.71 ms / 57 KB / 459 allocs | 16.01 ms / 57 KB / 459 allocs | 1.02× |
| Embed/Long | 389.4 ms / 780 KB / 2627 allocs | 394.0 ms / 778 KB / 2627 allocs | 1.01× |

### 生データ

asm:

```
BenchmarkRuntimeInit-8   	   16776	     74049 ns/op	 1606725 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   16356	     75026 ns/op	 1606726 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   16401	     73881 ns/op	 1606725 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   15924	     73800 ns/op	 1606725 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   16131	     71865 ns/op	 1606726 B/op	      13 allocs/op
BenchmarkModelLoad-8     	      67	  17327956 ns/op	15488057 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      66	  17574852 ns/op	15722722 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      66	  17576421 ns/op	15722722 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      66	  17337614 ns/op	15722722 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      66	  17295253 ns/op	15722722 B/op	       4 allocs/op
BenchmarkEmbed/Short-8   	      75	  15756199 ns/op	   56789 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      74	  15710990 ns/op	   56796 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      75	  15710089 ns/op	   57283 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      75	  15708508 ns/op	   56789 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      72	  15810309 ns/op	   57324 B/op	     459 allocs/op
BenchmarkEmbed/Long-8    	       3	 391089736 ns/op	  772514 B/op	    2627 allocs/op
BenchmarkEmbed/Long-8    	       3	 387908944 ns/op	  784861 B/op	    2628 allocs/op
BenchmarkEmbed/Long-8    	       3	 391670125 ns/op	  797208 B/op	    2629 allocs/op
BenchmarkEmbed/Long-8    	       3	 389448986 ns/op	  772514 B/op	    2627 allocs/op
BenchmarkEmbed/Long-8    	       3	 388439208 ns/op	  772514 B/op	    2627 allocs/op
```

pure-Go:

```
BenchmarkRuntimeInit-8   	   18717	     63630 ns/op	 1524806 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   18636	     64117 ns/op	 1524806 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   18842	     64494 ns/op	 1524806 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   18175	     65812 ns/op	 1524806 B/op	      13 allocs/op
BenchmarkRuntimeInit-8   	   17924	     66560 ns/op	 1524806 B/op	      13 allocs/op
BenchmarkModelLoad-8     	      70	  16675654 ns/op	14824283 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      67	  16544381 ns/op	15488055 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      68	  16649817 ns/op	15260290 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      69	  16523271 ns/op	15039204 B/op	       4 allocs/op
BenchmarkModelLoad-8     	      69	  16628987 ns/op	15039127 B/op	       4 allocs/op
BenchmarkEmbed/Short-8   	      72	  15997126 ns/op	   57324 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      73	  16068201 ns/op	   56803 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      74	  15963268 ns/op	   56296 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      74	  16012233 ns/op	   57297 B/op	     459 allocs/op
BenchmarkEmbed/Short-8   	      69	  16053890 ns/op	   56296 B/op	     459 allocs/op
BenchmarkEmbed/Long-8    	       3	 395047000 ns/op	  772514 B/op	    2627 allocs/op
BenchmarkEmbed/Long-8    	       3	 392465792 ns/op	  772514 B/op	    2627 allocs/op
BenchmarkEmbed/Long-8    	       3	 393785639 ns/op	  772514 B/op	    2627 allocs/op
BenchmarkEmbed/Long-8    	       3	 398288194 ns/op	  784866 B/op	    2628 allocs/op
BenchmarkEmbed/Long-8    	       3	 394005917 ns/op	  784861 B/op	    2628 allocs/op
```

### バイナリサイズ

同一プログラム（`examples/embedding`）を両バックエンドでビルドして比較。
どちらも `darwin/arm64`、Mach-O 実行ファイル。

| ビルド | asm（gcasm） | pure-Go | 差 |
|---|---|---|---|
| 通常（DWARF/シンボル付き） | 23.41 MB | 23.39 MB | ±0（誤差） |
| `-ldflags='-s -w' -trimpath` | 18.76 MB | 15.40 MB | **asm が +3.36 MB（22% 大きい）** |

`size(1)` でセグメントを見ると、差はほぼすべて `__TEXT`（実行コード）由来:
asm 17.84 MB / pure 14.34 MB。gcasm が wasm をほぼリテラルにネイティブアセンブリに
落とすのに対し、pure-Go は `cmd/compile` の最適化（インライン化・DCE 等）が効くため、
コードサイズでも pure が小さく収まる。速度が同程度（上記）なのにバイナリも小さいので、
スカラー限定ビルドでは pure 版のほうが配布効率も良い。

### 所感

- **推論（Embed）は pure-Go と asm が事実上同速**（pure が 1〜2% 遅いだけでノイズ圏内）。
  この wasm は SIMD 無効ビルドで推論カーネルがスカラー固定のため、pure-Go の
  直線コードを `cmd/compile` が最適化（インライン化・レジスタ割当・境界チェック除去）した
  結果が、gcasm の wasm スタックマシン模倣コードに匹敵する、という形。
- **RuntimeInit / ModelLoad は pure-Go がわずかに高速**。asm トラポリンや keepalive
  データ分の省略が効いている。メモリ使用量も pure が一貫して少ない。
- asm 版の存在意義は「移植性」ではなく「SIMD 等の pure-Go で表現しづらい命令が使える
  ようになったとき」。現状のスカラー限定ビルドでは、pure 版は**性能ペナルティなしに
  全 GOARCH でビルドできる・配布が単純**という実利的な利点がある。

## 移行時に見つかった既存バグ

`model.go` の `LoadModel` は、ロード成功時に wasm メモリ上のステージングバッファ
（モデルサイズ全体）を解放しておらず、ロードを繰り返すと 4GiB の wasm 線形メモリを
使い切って `wasm: unreachable` トラップで落ちていた（wazero 版でも潜在的に存在）。
`rten_load_model` は内部でデータをコピーするため、呼び出し後に必ず `Deallocate`
するよう修正済み。回帰テスト: `TestRepeatedLoadModel`（`modelload_test.go`）。
