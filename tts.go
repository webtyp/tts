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
