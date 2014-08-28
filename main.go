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
	Init(*Data)
	Lesson() string
	PracticeOrder() int
	Practice(UI, Input) PracticeResult
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
		e.Init(&data)
	}

	// stat
	fmt.Printf("%d practice entries, %d words\n", len(data.Entries), len(data.Words))

	cmd := "practice"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	// commands
	args := os.Args[2:]
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
	default:
		if handler, ok := commandHandlers[cmd]; ok {
			handler(&data, args)
		} else {
			log.Fatalf("unknown command %s", cmd)
		}
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

type EntrySorter []*Entry

func (s EntrySorter) Len() int { return len(s) }

func (s EntrySorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (self EntrySorter) Less(i, j int) bool {
	left, right := self[i], self[j]
	leftLastHistory := left.History[len(left.History)-1]
	rightLastHistory := right.History[len(right.History)-1]
	leftLesson := left.Lesson()
	rightLesson := right.Lesson()
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
			leftTypeOrder := left.PracticeOrder()
			rightTypeOrder := right.PracticeOrder()
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
