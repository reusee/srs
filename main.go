package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/reusee/gobfile"
)

var (
	p = fmt.Printf
	s = fmt.Sprintf
)

type Data struct {
	Practices    []PracticeEntry
	SignatureSet map[string]struct{}
	Words        []*Word
	save         func()
}

type HistoryEntry struct {
	Level int
	Time  time.Time
}

type PracticeResult int

const (
	LEVEL_UP PracticeResult = iota
	LEVEL_RESET
	EXIT
	NONE
)

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

type PracticeEntry interface {
	Signature() string
	Init(*Data)
	Lesson() string
	PracticeOrder() int
	Practice(UI, Input) PracticeResult
	Weight() int

	LastHistory() HistoryEntry
	LevelUp()
	LevelReset()
	AddHistory(HistoryEntry)
	GetHistory() []HistoryEntry
}

type HistoryImpl struct {
	History []HistoryEntry
}

func (h HistoryImpl) LastHistory() HistoryEntry {
	return h.History[len(h.History)-1]
}

func (h *HistoryImpl) LevelUp() {
	h.History = append(h.History, HistoryEntry{
		Level: h.LastHistory().Level + 1,
		Time:  time.Now(),
	})
}

func (h *HistoryImpl) LevelReset() {
	h.History = append(h.History, HistoryEntry{
		Level: 0,
		Time:  time.Now(),
	})
}

func (h *HistoryImpl) AddHistory(entry HistoryEntry) {
	h.History = append(h.History, entry)
}

func (h HistoryImpl) GetHistory() []HistoryEntry {
	return h.History
}

var (
	rootPath string

	LevelTime = []time.Duration{
		0,
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())

	_, rootPath, _, _ = runtime.Caller(0)
	rootPath, _ = filepath.Abs(rootPath)
	rootPath = filepath.Dir(rootPath)

	base := 2.2
	for i := 0.0; i < 12; i++ {
		t := time.Duration(float64(time.Hour*24) * math.Pow(base, i))
		LevelTime = append(LevelTime, t)
	}
}

var commandHandlers = map[string]func(*Data, []string){}

func main() {
	data := Data{
		SignatureSet: make(map[string]struct{}),
	}

	db, err := gobfile.New(&data, filepath.Join(rootPath, "db.gob"), 47213)
	if err != nil {
		panic(err)
	}
	data.save = func() {
		err := db.Save()
		if err != nil {
			panic(err)
		}
	}
	for _, e := range data.Practices {
		e.Init(&data)
	}

	// stat
	fmt.Printf("%d practice entries, %d words\n", len(data.Practices), len(data.Words))

	cmd := "practice"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	// commands
	switch cmd {
	case "complete":
		data.Complete(args)
	case "history":
		data.PrintHistory(args)
	case "practice":
		data.Practice(args)
	case "words":
		data.ListWords(args)
	case "edit-word":
		data.EditWord(args)
	case "fix":

	default:
		if handler, ok := commandHandlers[cmd]; ok {
			handler(&data, args)
		} else {
			log.Fatalf("unknown command %s", cmd)
		}
	}

	data.save()
	db.Close()

}

func playAudio(f string) {
	exec.Command("mplayer", filepath.Join(rootPath, "files", f)).Run()
}

func (d *Data) AddEntry(entry PracticeEntry) (added bool) {
	sig := entry.Signature()
	if _, has := d.SignatureSet[sig]; has {
		return
	}
	d.Practices = append(d.Practices, entry)
	d.SignatureSet[sig] = struct{}{}
	added = true
	return
}

func (d *Data) Complete([]string) {
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

func (data *Data) PrintHistory([]string) {
	counter := make(map[string]int)
	total := 0
	for _, entry := range data.Practices {
		for _, h := range entry.GetHistory() {
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

func (d *Data) ListWords([]string) {
	for i, w := range d.Words {
		p("%-6d %-20s %s\n", i, w.AudioFile, w.Text)
	}
}

func (d *Data) EditWord(args []string) {
	n, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("expected word index, not %v", args[0])
	}
	word := d.Words[n]
	fmt.Printf("editing %-20s %s\n", word.AudioFile, word.Text)
	var text string
	fmt.Scanf("%s", &text)
	word.Text = text
}
