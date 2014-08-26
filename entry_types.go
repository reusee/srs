package main

import "encoding/gob"

func init() {
	gob.Register(new(AudioToWordEntry))
	gob.Register(new(WordToAudioEntry))
	gob.Register(new(SentenceEntry))
	gob.Register(new(DialogEntry))
}

// audio to word

type AudioToWordEntry struct {
	WordIndex int
	word      *Word
}

func (e *AudioToWordEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*AudioToWordEntry); ok && t.WordIndex == e.WordIndex {
		return true
	}
	return false
}

func (e *AudioToWordEntry) Load(data *Data) {
	e.word = data.Words[e.WordIndex]
}

func (e *AudioToWordEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
}

func (e *AudioToWordEntry) PracticeOrder() int {
	return 1
}

func (e *AudioToWordEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press any key to show answer")
	input()
	ui("set-text", e.word.Text)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		ui("set-hint", "playing...")
		playAudio(e.word.AudioFile)
		ui("set-hint", "")
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}

// word to audio

type WordToAudioEntry struct {
	WordIndex int
	word      *Word
}

func (e *WordToAudioEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*WordToAudioEntry); ok && t.WordIndex == e.WordIndex {
		return true
	}
	return false
}

func (e *WordToAudioEntry) Load(data *Data) {
	e.word = data.Words[e.WordIndex]
}

func (e *WordToAudioEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
}

func (e *WordToAudioEntry) PracticeOrder() int {
	return 3
}

func (e *WordToAudioEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-text", e.word.Text)
	ui("set-hint", "press any key to play audio")
	input()
repeat:
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}

// sentence

type SentenceEntry struct {
	AudioFile string
}

func (e *SentenceEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*SentenceEntry); ok && t.AudioFile == e.AudioFile {
		return true
	}
	return false
}

func (e *SentenceEntry) Load(*Data) {}

func (e *SentenceEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.AudioFile)[0]
}

func (e *SentenceEntry) PracticeOrder() int {
	return 2
}

func (e *SentenceEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.AudioFile)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		ui("set-hint", "playing...")
		playAudio(e.AudioFile)
		ui("set-hint", "")
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}

// dialog

type DialogEntry struct {
	AudioFile string
}

func (e *DialogEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*DialogEntry); ok && t.AudioFile == e.AudioFile {
		return true
	}
	return false
}

func (e *DialogEntry) Load(*Data) {}

func (e *DialogEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.AudioFile)[0]
}

func (e *DialogEntry) PracticeOrder() int {
	return 4
}

func (e *DialogEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.AudioFile)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		ui("set-hint", "playing...")
		playAudio(e.AudioFile)
		ui("set-hint", "")
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}