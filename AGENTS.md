# Agent Guide — `webtyp/tts`

Constraints for agents working on this library. **Read this before any change.**
The current work order, when one exists, is [docs/PLAN.md](docs/PLAN.md). The ecosystem plan
is [`agent/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md`](https://github.com/webtyp/agent/blob/main/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md).

## What this library is

The text-to-speech **contract**. It holds an interface and its value types only. It must never contain a
model, weights, a tokenizer, a browser call, or a mock outside `_test.go` files. The model
that implements it lives in its own repository, so importing this contract adds no model to a
binary.

## The builds that define "done"

```bash
go vet ./...
gotest
gotest -tinygo
GOOS=js GOARCH=wasm go build ./...
```

## Never import these

| Never | Use instead | Why |
|---|---|---|
| `fmt`, `errors`, `strings`, `strconv` | `webtyp.com/fmt` | isomorphic, small under TinyGo |
| `context` (stdlib) | `webtyp.com/context` | every webtyp API takes `*context.Context` |
| `time` | `webtyp.com/time` | |
| `encoding/json` | nothing | reflection JSON costs ~1 MB of wasm |
| `map[K]V` | a slice | TinyGo's map runtime is a size tax |
| `os`, `log`, `net/http`, `syscall/js` | nothing | a contract never touches the environment |

## Common mistakes to avoid

- Adding a streaming mode as a flag on the batch method. Streaming is `StreamSynthesizer` and its
  `Stream`: a separate interface, so a batch-only model never has to fake it.
- Declaring a local audio type. Use `audio.PCM`.
