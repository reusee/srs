package main

import (
	"encoding/gob"
	"regexp"
)

func init() {
	gob.Register(new(AudioToWordEntry))
	gob.Register(new(WordToAudioEntry))
	gob.Register(new(SentenceEntry))
	gob.Register(new(DialogEntry))
}

var (
	lessonPattern = regexp.MustCompile("[0-9]+")
)

// audio to word

type AudioToWordEntry struct {
	*HistoryImpl
	WordIndex int
	word      *Word
}

func (e *AudioToWordEntry) Signature() string {
	return s("atw-%d", e.WordIndex)
}

func (e *AudioToWordEntry) Init(data *Data) {
	e.word = data.Words[e.WordIndex]
}

func (e *AudioToWordEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
}

func (e *AudioToWordEntry) PracticeOrder() int {
	return 1
}

func (e *AudioToWordEntry) Weight() int {
	return 10
}

func (e *AudioToWordEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press any key to show answer")
	input()
	ui("set-text", e.word.Text)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		ui("set-hint", "playing...")
		playAudio(e.word.AudioFile)
		ui("set-hint", "")
		goto repeat
	}
}

// word to audio

type WordToAudioEntry struct {
	*HistoryImpl
	WordIndex int
	word      *Word
}

func (e *WordToAudioEntry) Signature() string {
	return s("wta-%d", e.WordIndex)
}

func (e *WordToAudioEntry) Init(data *Data) {
	e.word = data.Words[e.WordIndex]
}

func (e *WordToAudioEntry) Lesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
}

func (e *WordToAudioEntry) PracticeOrder() int {
	return 3
}

func (e *WordToAudioEntry) Weight() int {
	return 10
}

func (e *WordToAudioEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-text", e.word.Text)
	ui("set-hint", "press any key to play audio")
	input()
repeat:
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto repeat
	}
}

// sentence common

type sentenceCommon string

func (sen sentenceCommon) Signature() string {
	return s("sen-%s", sen)
}

func (s sentenceCommon) Lesson() string {
	return lessonPattern.FindStringSubmatch(string(s))[0]
}

func (s sentenceCommon) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(string(s))
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		ui("set-hint", "playing...")
		playAudio(string(s))
		ui("set-hint", "")
		goto repeat
	}
}

// sentence

type SentenceEntry struct {
	*HistoryImpl
	sentenceCommon
	AudioFile string
}

func (e *SentenceEntry) Init(*Data) {
	e.sentenceCommon = sentenceCommon(e.AudioFile)
}

func (e *SentenceEntry) PracticeOrder() int {
	return 2
}

func (e *SentenceEntry) Weight() int {
	return 10
}

// dialog

type DialogEntry struct {
	*HistoryImpl
	sentenceCommon
	AudioFile string
}

func (e *DialogEntry) Init(*Data) {
	e.sentenceCommon = sentenceCommon(e.AudioFile)
}

func (e *DialogEntry) PracticeOrder() int {
	return 4
}

func (e *DialogEntry) Weight() int {
	return 10
}
