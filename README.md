# lib-go

Shared Go utility packages for Thruput services: logging, worker pools,
buffer recycling, XDR encoding, and the [goxdr](https://github.com/xdrpp/goxdr)
RPC driver setup they share.

## Use

Requires Go 1.24 or later. The packages build on Linux and macOS; `libnet` and
`libruntime` use Unix socket options and process groups, so they do not build
on Windows.

```sh
go get github.com/thruputchain/lib-go
```

```go
import "github.com/thruputchain/lib-go/liblog"
```

## Packages

| Package | Contents |
| --- | --- |
| `libflow` | Worker pools, with and without a queue, and a batching buffer that flushes when it reaches a size limit or a maximum wait. |
| `liblock` | `KeyedMutex`, a mutex per key that frees a key's entry once no one holds or waits on it. |
| `liblog` | Context-scoped [zerolog](https://github.com/rs/zerolog) logging with labels and `Enterf`/`Leavef` spans. |
| `libmath` | `DivideRoundUp` for integer types. |
| `libmem` | Borrowed values and a byte buffer pool whose buffers return to the pool when recycled. |
| `libmetrics` | [OpenTelemetry](https://opentelemetry.io/) metric attributes carried on a `context.Context`. |
| `libnet` | `Network`, `Dialer` and `Listener` interfaces over TCP, including a variant that sets `SO_REUSEADDR` and `SO_REUSEPORT`. |
| `librand` | Independently seeded `math/rand` sources and a goroutine-safe wrapper. |
| `librpc` | Serial and parallel goxdr RPC drivers with pooled messages and calls, and switches for call and stats logging. |
| `libruntime` | Signal hooks: dump goroutine stacks, or kill the process group, on `SIGINT`/`SIGTERM`. |
| `libstat` | Counters, latency meters, ring buffers, and statistics over a moving window of the intervals between events. |
| `libtext` | Byte encodings for JSON and strings, filename-safe hashes, text forms, templates and a line-prefixing writer. |
| `libtime` | Time and duration formatting, and a `Time` interface with a manually advanced `TestTime`. |
| `libxdr` | XDR marshaling, SHA-256 hashing of XDR values, and XDR to JSON and text conversion. |
| `must` | Assertions and panic-on-error helpers whose errors carry a stack trace. |

## Test

```sh
go vet ./...
go test ./...
```

The `libxdr` tests use the protocol types from
[thruput-protocol-go](https://github.com/thruputchain/thruput-protocol-go).

## License

Apache License 2.0; see [LICENSE](LICENSE).
