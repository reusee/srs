package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	commandHandlers["add-words"] = AddWords
	commandHandlers["add-sentences"] = AddSentences
}

func AddWords(data *Data) {
	for _, audioFile := range os.Args[2:] {
		audioFile, err := filepath.Abs(audioFile)
		if err != nil {
			panic(err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		index := data.GetWordIndex(audioFile, "")
		// audio to word entry
		entry := &Entry{
			IsEntry: &AudioToWordEntry{
				WordIndex: index,
			},
			History: []HistoryEntry{
				{
					Level: 0,
					Time:  time.Now(),
				},
			},
		}
		added := data.AddEntry(entry)
		if added {
			p("added AudioToWordEntry %s\n", audioFile)
		} else {
			p("skip %s\n", audioFile)
		}
		// word to audio entry
		entry = &Entry{
			IsEntry: &WordToAudioEntry{
				WordIndex: index,
			},
			History: []HistoryEntry{
				{
					Level: 0,
					Time:  time.Now(),
				},
			},
		}
		added = data.AddEntry(entry)
		if added {
			p("added WordToAudioEntry %s\n", audioFile)
		} else {
			p("skip %s\n", audioFile)
		}
	}
	data.Complete()
}

func AddSentences(data *Data) {
	for _, audioFile := range os.Args[2:] {
		audioFile, err := filepath.Abs(audioFile)
		if err != nil {
			panic(err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		// add sentence
		entry := &Entry{
			IsEntry: &SentenceEntry{
				AudioFile: audioFile,
			},
			History: []HistoryEntry{
				{
					Level: 0,
					Time:  time.Now(),
				},
			},
		}
		added := data.AddEntry(entry)
		if added {
			p("added SentenceEntry %s\n", audioFile)
		} else {
			p("skip %s\n", audioFile)
		}
	}
}
