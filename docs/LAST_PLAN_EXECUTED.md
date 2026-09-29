---
PLAN: "feat: tts contract — Synthesizer and StreamSynthesizer"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase 4b** of
> [`AGENT_ECOSYSTEM_MASTER_PLAN.md`](https://github.com/webtyp/agent/blob/main/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md).
> **Blocked until `webtyp.com/audio` v0.1.0 is published** (phase 4a). Runs in parallel with `webtyp/stt`.

# Plan — `webtyp.com/tts`: the contract for turning text into speech

## 0. Context

**Text-to-speech (TTS)**, also called speech synthesis, turns written text into spoken audio.
It is planned for version 2 of the webtyp agent: the agent's answer is spoken aloud. Version 1
is text only, but the contract is declared now so the pipeline is wired from the start.

This plan creates only the contract. The model that implements it (candidates in
`docs/TEXT_TO_SPEECH_BROWSER_MODELS.md`) runs in the browser in Go/TinyGo, like
`webtyp/embed`, and lives in its own repository once chosen.

## Development rules (inline)

- **Contract library:** interface only. No model, no browser call, no mock outside `_test.go` files.
- Every file compiles under `GOOS=js GOARCH=wasm` and TinyGo.
- **Never import:** `fmt`/`errors`/`strings`/`strconv` (use `webtyp.com/fmt`), `context`
  (use `webtyp.com/context`), `encoding/json`, `time`, `map[K]V`, `os`, `log`, `net/http`.
- Dependencies: `webtyp.com/audio` (v0.1.0) and `webtyp.com/context`; the test file also uses `webtyp.com/fmt`. Nothing else.
- Flat layout, tests with `testing` only. Do **not** run `gopush`/`codejob`.

## Design gate (api-design — five answers)

1. **Prior art.** **Web Speech API** `speechSynthesis.speak(utterance)`: browser-owned
   voices, no access to the audio. **Google Cloud TTS**: `SynthesizeSpeech(input, voice,
   audioConfig)` → audio bytes. **Piper**: `synthesize(text)` → PCM with a voice fixed at load
   time. **Coqui/Kokoro**: `tts(text, voice)` → waveform. We follow Piper: the **voice is chosen
   when the implementation is constructed**, because a voice is a model file that is expensive
   to load, so switching per call is not a real option in the browser. The output is
   `audio.PCM`, which Web Audio plays without conversion. For streaming (speaking while the
   model is still writing), **ElevenLabs input streaming** and **Azure Speech** text streams
   both accept text pieces and return audio per completed sentence. We use the same
   `NewStream` → `Write` → `Close` shape as `webtyp/stt`, so both voice directions read the
   same way. It is a separate interface (`StreamSynthesizer`), and nothing consumes it in
   version 1. It is declared now so the voice loop is wired from the start.
2. **Novice-name test.** `tts.Synthesizer.Synthesize(ctx, text)` → `audio.PCM` reads as
   "synthesize this text into audio". It is the word Google, Microsoft and the Web Speech API
   all use.
3. **Complexity ledger.** Concepts +3 (`Synthesizer`, `StreamSynthesizer`, `Stream`) / −0. Ways to do the same thing +0.
4. **Where it belongs.** TTS has its own models, licences and quality trade-offs, separate from
   speech-to-text. It is a contract repo, so importing it adds no model to a binary.
5. **What it deletes.** Nothing. This is new capability.

## Stage 1 — the contract

**`tts.go`**

```go
// Package tts is the text-to-speech contract: text in, PCM audio out. Implementations live in
// their own repositories.
package tts

import (
	"webtyp.com/audio"
	"webtyp.com/context"
)

// Synthesizer turns text into spoken audio, in the voice the implementation was built with.
type Synthesizer interface {
	Synthesize(ctx *context.Context, text string) (audio.PCM, error)
}

// StreamSynthesizer speaks text while it is still arriving, e.g. from llm.Streamer.
type StreamSynthesizer interface {
	NewStream(ctx *context.Context) (Stream, error)
}

// Stream is one answer being spoken as its text arrives.
type Stream interface {
	// Write adds the next piece of text and returns the audio of every sentence that piece
	// completed, in order. Text of an unfinished sentence waits for the next Write or Close.
	Write(text string) ([]audio.PCM, error)

	// Close speaks whatever text is left and returns its audio. Write after Close is an error.
	Close() ([]audio.PCM, error)
}
```

## Stage 2 — consumer-shaped test

**File:** `example_test.go`, package `tts_test`. Declare in the file a
`type toneSynthesizer struct{}` whose `Synthesize` returns 100 ms of silence per word at
22050 Hz mono, and an error `fmt.Err("tts: empty text")` for empty text.

- `Example_speakAnswer`: `"abrimos a las ocho"` → prints `DurationMS()`. `// Output: 400`
- `TestSynthesize_EmptyTextErrors`: `""` → error.
- `TestStream_SpeaksPerSentence`: a `sentenceStream` in the test file buffers text and, on each
  `.`, returns one `audio.PCM` of 100 ms per word of that sentence. `Write("Abrimos a")` → 0
  clips; `Write(" las ocho. Hasta")` → 1 clip of 400 ms; `Close()` → 1 clip of 100 ms; `Write`
  after `Close` → error. Also assert that `var _ tts.StreamSynthesizer = sentenceSynthesizer{}` compiles.

## Stages

| Stage | Files | Acceptance |
|---|---|---|
| 1 | `tts.go`, `go.mod` | builds for host, wasm, TinyGo |
| 2 | `example_test.go` | passes under `gotest` and `gotest -tinygo` |
