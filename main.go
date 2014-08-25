package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	m "./memory"
	"github.com/reusee/gobfile"
)

var (
	p = fmt.Printf
)

type Data struct {
	Entries []*Entry
	Words   []*Word
	save    func()
}

type Entry struct {
	History []HistoryEntry
	IsEntry
}

type HistoryEntry struct {
	Level int
	Time  time.Time
}

type IsEntry interface {
	IsTheSame(IsEntry) bool
}

type AudioToWordEntry struct {
	WordIndex int
}

func (e *AudioToWordEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*AudioToWordEntry); ok && t.WordIndex == e.WordIndex {
		return true
	}
	return false
}

type WordToAudioEntry struct {
	WordIndex int
}

func (e *WordToAudioEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*WordToAudioEntry); ok && t.WordIndex == e.WordIndex {
		return true
	}
	return false
}

type SentenceEntry struct {
	AudioFile string
}

func (e *SentenceEntry) IsTheSame(entry IsEntry) bool {
	if t, ok := entry.(*SentenceEntry); ok && t.AudioFile == e.AudioFile {
		return true
	}
	return false
}

type Word struct {
	AudioFile string
	Text      string
}

func (d *Data) GetWordIndex(audioFile string, text string) int {
	for i, w := range d.Words {
		if w.AudioFile == audioFile {
			return i
		}
	}
	d.Words = append(d.Words, &Word{
		AudioFile: audioFile,
		Text:      text,
	})
	return len(d.Words) - 1
}

var rootPath string

func init() {
	gob.Register(new(AudioToWordEntry))
	gob.Register(new(WordToAudioEntry))
	gob.Register(new(SentenceEntry))

	_, rootPath, _, _ = runtime.Caller(0)
	rootPath, _ = filepath.Abs(rootPath)
	rootPath = filepath.Dir(rootPath)
}

func main() {
	var data Data

	db, err := gobfile.New(&data, filepath.Join(rootPath, "db.gob"), 47213)
	if err != nil {
		panic(err)
	}
	defer func() {
		data.save()
		db.Close()
	}()
	data.save = func() {
		err := db.Save()
		if err != nil {
			panic(err)
		}
	}

	// stat
	fmt.Printf("%d practice entries, %d words\n", len(data.Entries), len(data.Words))

	// commands
	switch os.Args[1] {
	case "migrate":
		data.Migrate()
	case "complete":
		data.Complete()
	case "history":
		data.PrintHistory()
	}

}

func playAudio(f string) {
	exec.Command("mpg123", filepath.Join(rootPath, "files", f)).Run()
}

func (d *Data) AddEntry(entry *Entry) (added bool) {
	for _, e := range d.Entries {
		if entry.IsEntry.IsTheSame(e.IsEntry) {
			return
		}
	}
	d.Entries = append(d.Entries, entry)
	added = true
	return
}

func (d *Data) Complete() {
	var text string
	for _, word := range d.Words {
		if strings.TrimSpace(word.Text) == "" {
			p("%s\n", word.AudioFile)
			playAudio(word.AudioFile)
			fmt.Scanf("%s", &text)
			word.Text = text
			d.save()
		}
	}
}

func (data *Data) Migrate() {
	mem := &m.Memory{
		Concepts: make(map[string]*m.Concept),
		Connects: make(map[string]*m.Connect),
	}
	mem.Load()
	for _, connect := range mem.Connects {
		from := mem.Concepts[connect.From]
		to := mem.Concepts[connect.To]
		history := connect.Histories
		var entry *Entry
		if from.What == m.WORD && to.What == m.AUDIO {
			entry = &Entry{
				IsEntry: &WordToAudioEntry{
					WordIndex: data.GetWordIndex(to.File, from.Text),
				},
			}
		} else if from.What == m.AUDIO && to.What == m.WORD {
			entry = &Entry{
				IsEntry: &AudioToWordEntry{
					WordIndex: data.GetWordIndex(from.File, to.Text),
				},
			}
		} else if from.What == m.AUDIO && to.What == m.SENTENCE {
			entry = &Entry{
				IsEntry: &SentenceEntry{
					AudioFile: from.File,
				},
			}
		} else {
			panic("not here")
		}
		if data.AddEntry(entry) {
			for _, e := range history {
				entry.History = append(entry.History, HistoryEntry{
					Level: e.Level,
					Time:  e.Time,
				})
			}
		}
	}
}

func (data *Data) PrintHistory() {
	counter := make(map[string]int)
	total := 0
	for _, entry := range data.Entries {
		for _, h := range entry.History {
			if h.Level == 0 {
				continue
			}
			counter[h.Time.Format("2006-01-02")]++
			total++
		}
	}
	var dates []string
	for date := range counter {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		p("%s %d\n", date, counter[date])
	}

	per := time.Second * 5
	p("total %d, %v / %v\n", total, per, time.Duration(total)*per)
}
