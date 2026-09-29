package tts_test

import (
	"testing"

	"webtyp.com/audio"
	"webtyp.com/context"
	"webtyp.com/fmt"
	"webtyp.com/tts"
)

const sampleRate = 22050

// silence returns 100 ms of silence per word.
func silence(words int) audio.PCM {
	return audio.PCM{SampleRate: sampleRate, Channels: 1, Samples: make([]float32, words*sampleRate/10)}
}

func countWords(s string) int {
	n, in := 0, false
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '.' {
			in = false
		} else if !in {
			in = true
			n++
		}
	}
	return n
}

type toneSynthesizer struct{}

func (toneSynthesizer) Synthesize(ctx *context.Context, text string) (audio.PCM, error) {
	if text == "" {
		return audio.PCM{}, fmt.Err("tts: empty text")
	}
	return silence(countWords(text)), nil
}

func Example_speakAnswer() {
	var s tts.Synthesizer = toneSynthesizer{}
	pcm, _ := s.Synthesize(context.Background(), "abrimos a las ocho")
	fmt.Println(pcm.DurationMS())
	// Output: 400
}

func TestSynthesize_EmptyTextErrors(t *testing.T) {
	if _, err := (toneSynthesizer{}).Synthesize(context.Background(), ""); err == nil {
		t.Fatal("empty text must be an error")
	}
}

type sentenceSynthesizer struct{}

func (sentenceSynthesizer) NewStream(ctx *context.Context) (tts.Stream, error) {
	return &sentenceStream{}, nil
}

type sentenceStream struct {
	pending string
	closed  bool
}

func (s *sentenceStream) Write(text string) ([]audio.PCM, error) {
	if s.closed {
		return nil, fmt.Err("tts: write after close")
	}
	s.pending += text
	var out []audio.PCM
	start := 0
	for i := 0; i < len(s.pending); i++ {
		if s.pending[i] == '.' {
			out = append(out, silence(countWords(s.pending[start:i])))
			start = i + 1
		}
	}
	s.pending = s.pending[start:]
	return out, nil
}

func (s *sentenceStream) Close() ([]audio.PCM, error) {
	s.closed = true
	if countWords(s.pending) == 0 {
		return nil, nil
	}
	return []audio.PCM{silence(countWords(s.pending))}, nil
}

func TestStream_SpeaksPerSentence(t *testing.T) {
	var _ tts.StreamSynthesizer = sentenceSynthesizer{}
	st, _ := sentenceSynthesizer{}.NewStream(context.Background())

	if clips, _ := st.Write("Abrimos a"); len(clips) != 0 {
		t.Fatalf("an unfinished sentence must wait, got %d clips", len(clips))
	}
	clips, _ := st.Write(" las ocho. Hasta")
	if len(clips) != 1 || clips[0].DurationMS() != 400 {
		t.Fatalf("want 1 clip of 400 ms, got %d clips", len(clips))
	}
	clips, _ = st.Close()
	if len(clips) != 1 || clips[0].DurationMS() != 100 {
		t.Fatalf("Close must speak the rest: want 1 clip of 100 ms, got %d", len(clips))
	}
	if _, err := st.Write("más"); err == nil {
		t.Fatal("Write after Close must be an error")
	}
}
