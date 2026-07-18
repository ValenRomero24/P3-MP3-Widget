package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)


type BeepEngine struct {
	mu			sync.Mutex
	streamer 	beep.StreamSeekCloser
	format		beep.Format
	ctrl		*beep.Ctrl
	volume		*effects.Volume
	sampleRate	beep.SampleRate
}

func NewBeepEngine() *BeepEngine {
	sr := beep.SampleRate(44100)
	_ = speaker.Init(sr, sr.N(time.Second/10))
	return &BeepEngine{sampleRate : sr}
}

func (e * BeepEngine) Play(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer != nil {
		e.streamer.Close()
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var streamer beep.StreamSeekCloser
	var format	 beep.Format
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	case ".flac":
		streamer, format, err = flac.Decode(file)
	default:
		file.Close()
		return fmt.Errorf("formato no soportado")
	}

	if err != nil {
		file.Close()
		return err
	}

	e.streamer	= streamer
	e.format 	= format

	resampled	:= beep.Resample(4, format.SampleRate, e.sampleRate, streamer)
	
	e.ctrl	= &beep.Ctrl{
		Streamer:	resampled,
		Paused:		false,
	} 

	e.volume = &effects.Volume{
		Streamer: 	e.ctrl,
		Base:		2,
		Volume:		0,
	}

	speaker.Clear()
	speaker.Play(e.volume)
	return nil
}


func (e *BeepEngine) TogglePause() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ctrl != nil {
		speaker.Lock()
		e.ctrl.Paused = !e.ctrl.Paused
		speaker.Unlock()
		return e.ctrl.Paused
	}
	return false
}

func (e *BeepEngine) Seek(offset time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer == nil { return }
	
	speaker.Lock()
	newPos := e.streamer.Position() + e.format.SampleRate.N(offset)
	if newPos < 0 {
		newPos = 0
	}
	if newPos > e.streamer.Len() {
		newPos = e.streamer.Len()
	}
	_ = e.streamer.Seek(newPos)
	speaker.Unlock()
}

func (e *BeepEngine) GetProgress() (time.Duration, time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer == nil || e.format.SampleRate == 0 {
		return 0, 0
	}

	pos := e.format.SampleRate.D(e.streamer.Position())
	tot := e.format.SampleRate.D(e.streamer.Len())

	return pos, tot
}