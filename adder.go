package main

import (
	"log"
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
			log.Fatalf("AddWords: wrong audio file path %v", err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		index := data.GetWordIndex(audioFile, "")
		// audio to word entry
		entry := &AudioToWordEntry{
			WordIndex: index,
			HistoryImpl: &HistoryImpl{
				History: []HistoryEntry{
					{
						Level: 0,
						Time:  time.Now(),
					},
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
		entry2 := &WordToAudioEntry{
			WordIndex: index,
			HistoryImpl: &HistoryImpl{
				History: []HistoryEntry{
					{
						Level: 0,
						Time:  time.Now(),
					},
				},
			},
		}
		added = data.AddEntry(entry2)
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
			log.Fatalf("AddSentences: wrong audio file path %v", err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		// add sentence
		entry := &SentenceEntry{
			AudioFile:      audioFile,
			sentenceCommon: sentenceCommon(audioFile),
			HistoryImpl: &HistoryImpl{
				History: []HistoryEntry{
					{
						Level: 0,
						Time:  time.Now(),
					},
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
			log.Fatalf("AddDialogs: wrong audio file path %v", err)
		}
		audioFile = strings.TrimPrefix(audioFile, filepath.Join(rootPath, "files"))
		// add sentence
		entry := &DialogEntry{
			AudioFile:      audioFile,
			sentenceCommon: sentenceCommon(audioFile),
			HistoryImpl: &HistoryImpl{
				History: []HistoryEntry{
					{
						Level: 0,
						Time:  time.Now(),
					},
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
