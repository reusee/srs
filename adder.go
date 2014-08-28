package main

import (
	"path/filepath"
	"strings"
	"time"
)

func init() {
	commandHandlers["add-words"] = AddWords
	commandHandlers["add-sentences"] = AddSentences
	commandHandlers["add-dialogs"] = AddDialogs
}

func AddWords(data *Data, args []string) {
	for _, audioFile := range args {
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
	data.Complete(nil)
}

func AddSentences(data *Data, args []string) {
	for _, audioFile := range args {
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

func AddDialogs(data *Data, args []string) {
	for _, audioFile := range args {
		audioFile, err := filepath.Abs(audioFile)
		if err != nil {
			panic(err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		// add sentence
		entry := &Entry{
			IsEntry: &DialogEntry{
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
			p("added DialogEntry %s\n", audioFile)
		} else {
			p("skip %s\n", audioFile)
		}
	}
}
