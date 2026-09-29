# tts
<img src="docs/img/badges.svg">

The text-to-speech contract of webtyp: `Synthesizer.Synthesize(ctx, text)` → `audio.PCM`. It is a contract only. The model that implements it runs in
the browser in Go/TinyGo and lives in its own repository.

> Voice ships in version 2 of the agent; version 1 is text only. The contract is declared now so
> the voice pipeline is wired from the start.

## Getting started

| I want to… | Use |
|---|---|
| speak the agent's answer | `tts.Synthesizer` |
| start speaking before the whole answer is written | `tts.StreamSynthesizer` → `Stream.Write` / `Close` (version 2) |
| implement a model for it | implement the interface in your own repository |
| pass audio around | `webtyp.com/audio` (`audio.PCM`) |

## Documentation

- [Architecture](docs/ARCHITECTURE.md): the contract, where it sits in the voice pipeline, and the candidate models.
- [Browser model research](docs/TEXT_TO_SPEECH_BROWSER_MODELS.md): the model comparison this project started from.
- [Agent guide](AGENTS.md): rules for anyone changing this library.
