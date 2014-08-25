package main

import (
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	m "./memory"
	"github.com/reusee/gobfile"
)

var (
	p = fmt.Printf

	lessonPattern = regexp.MustCompile("[0-9]+")
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
	Load(*Data)
	GetLesson() string
}

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

func (e *AudioToWordEntry) GetLesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
}

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

func (e *WordToAudioEntry) GetLesson() string {
	return lessonPattern.FindStringSubmatch(e.word.AudioFile)[0]
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

func (e *SentenceEntry) Load(*Data) {}

func (e *SentenceEntry) GetLesson() string {
	return lessonPattern.FindStringSubmatch(e.AudioFile)[0]
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

var LevelTime = []time.Duration{
	0,
}

func init() {
	gob.Register(new(AudioToWordEntry))
	gob.Register(new(WordToAudioEntry))
	gob.Register(new(SentenceEntry))

	_, rootPath, _, _ = runtime.Caller(0)
	rootPath, _ = filepath.Abs(rootPath)
	rootPath = filepath.Dir(rootPath)

	base := 2.2
	for i := 0.0; i < 12; i++ {
		t := time.Duration(float64(time.Hour*24) * math.Pow(base, i))
		LevelTime = append(LevelTime, t)
	}
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
	for _, e := range data.Entries {
		e.Load(&data)
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
	case "practice":
		data.Practice()
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

func (data *Data) Practice() {
	var entries []*Entry
	now := time.Now()
	for _, e := range data.Entries {
		lastHistory := e.History[len(e.History)-1]
		if lastHistory.Time.Add(LevelTime[lastHistory.Level]).Before(now) {
			entries = append(entries, e)
		}
	}
	sort.Sort(EntrySorter{entries, data})
	max := 25
	if len(entries) > max {
		entries = entries[:max]
	}
	ui_gtk(entries, data)
}

type EntrySorter struct {
	l []*Entry
	d *Data
}

func (s EntrySorter) Len() int { return len(s.l) }

func (s EntrySorter) Swap(i, j int) { s.l[i], s.l[j] = s.l[j], s.l[i] }

func (self EntrySorter) Less(i, j int) bool {
	left, right := self.l[i], self.l[j]
	leftLastHistory := left.History[len(left.History)-1]
	rightLastHistory := right.History[len(right.History)-1]
	leftLesson := left.GetLesson()
	rightLesson := right.GetLesson()
	leftLevelOrder := self.getLevelOrder(left)
	rightLevelOrder := self.getLevelOrder(right)
	if leftLevelOrder < rightLevelOrder {
		return true
	} else if leftLevelOrder > rightLevelOrder {
		return false
	} else if leftLevelOrder == rightLevelOrder && (leftLevelOrder == 1 || leftLevelOrder == 3) { // old connect
		if leftLastHistory.Level < rightLastHistory.Level { // review low level first
			return true
		} else if leftLastHistory.Level > rightLastHistory.Level {
			return false
		} else if leftLastHistory.Level == rightLastHistory.Level { // same level
			if leftLesson < rightLesson { // review earlier lesson first
				return true
			} else if leftLesson > rightLesson {
				return false
			} else { // randomize
				if rand.Intn(2) == 1 { // randomize
					return true
				}
				return false
			}
		}
	} else if leftLevelOrder == rightLevelOrder && leftLevelOrder == 2 { // new connect
		if leftLesson < rightLesson { // learn earlier lesson first
			return true
		} else if leftLesson > rightLesson {
			return false
		} else { // same lesson
			leftTypeOrder := self.getTypeOrder(left)
			rightTypeOrder := self.getTypeOrder(right)
			if leftTypeOrder < rightTypeOrder {
				return true
			} else if leftTypeOrder > rightTypeOrder {
				return false
			} else {
				return leftLastHistory.Time.Before(rightLastHistory.Time)
			}
			return true
		}
		return true
	}
	return false
	return true
}

func (s EntrySorter) getLevelOrder(e *Entry) int {
	lastHistory := e.History[len(e.History)-1]
	if lastHistory.Level >= 7 { // 7 -
		return 3
	}
	if lastHistory.Level > 0 { // 1 - 6
		return 1
	}
	return 2 // 0
}

func (s EntrySorter) getTypeOrder(e *Entry) (ret int) {
	switch e.IsEntry.(type) {
	case *AudioToWordEntry:
		ret = 1
	case *SentenceEntry:
		ret = 2
	case *WordToAudioEntry:
		ret = 3
	default:
		panic("fixme")
	}
	return
}
