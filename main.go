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
	Entries      []*Entry
	SignatureSet map[string]struct{}
	Words        []*Word
	save         func()
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
	Signature() string
	Init(*Data, *Entry)
	Lesson() string
	PracticeOrder() int
	Practice(UI, Input) PracticeResult
	Weight() int
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

var rootPath string

var LevelTime = []time.Duration{
	0,
}

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
	for _, e := range data.Entries {
		e.Init(&data, e)
	}

	// stat
	fmt.Printf("%d practice entries, %d words\n", len(data.Entries), len(data.Words))

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
		for _, e := range data.Entries {
			data.SignatureSet[e.Signature()] = struct{}{}
		}
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
	exec.Command("mpg123", filepath.Join(rootPath, "files", f)).Run()
}

func (d *Data) AddEntry(entry *Entry) (added bool) {
	sig := entry.Signature()
	if _, has := d.SignatureSet[sig]; has {
		return
	}
	d.Entries = append(d.Entries, entry)
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
