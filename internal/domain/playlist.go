package domain

import (
	"errors"
	"math/rand"
	"time"
)

type Track struct {
	Path  string
	Title string
}

type PlaylistManager struct {
	tracks         []Track
	originalTracks []Track
	currentIndex   int
	isShuffle      bool
	isLoop         bool
	rng            *rand.Rand
}

func NewPlaylistManager(tracks []Track) *PlaylistManager {
	orig := make([]Track, len(tracks))
	copy(orig, tracks)
	return &PlaylistManager{
		tracks:         tracks,
		originalTracks: orig,
		currentIndex:   0,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *PlaylistManager) CurrentTrack() (Track, error) {
	if len(m.tracks) == 0 {
		return Track{}, errors.New("la playlist está vacía")
	}
	return m.tracks[m.currentIndex], nil
}

func (m *PlaylistManager) Next() bool {
	if len(m.tracks) == 0 { return false }
	if m.currentIndex < len(m.tracks)-1 {
		m.currentIndex++
		return true
	}
	if m.isLoop {
		m.currentIndex = 0
		return true
	}
	return false
}

func (m *PlaylistManager) Prev() {
	if len(m.tracks) == 0 { return }
	if m.currentIndex > 0 {
		m.currentIndex--
	} else {
		m.currentIndex = len(m.tracks) - 1
	}
}

func (m *PlaylistManager) ToggleShuffle() {
	m.isShuffle = !m.isShuffle
	current, _ := m.CurrentTrack()
	if m.isShuffle {
		m.rng.Shuffle(len(m.tracks), func(i, j int) {
			m.tracks[i], m.tracks[j] = m.tracks[j], m.tracks[i]
		})
		for i, t := range m.tracks {
			if t.Path == current.Path { m.currentIndex = i; break }
		}
	} else {
		m.tracks = make([]Track, len(m.originalTracks))
		copy(m.tracks, m.originalTracks)
		for i, t := range m.tracks {
			if t.Path == current.Path { m.currentIndex = i; break }
		}
	}
}

func (m *PlaylistManager) ToggleLoop() { m.isLoop = !m.isLoop }
func (m *PlaylistManager) Status() (bool, bool) { return m.isShuffle, m.isLoop }