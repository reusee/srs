package main

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	LevelTime = []time.Duration{
		0,
	}
)

func init() {
	base := 2.2
	for i := 0.0; i < 12; i++ {
		t := time.Duration(float64(time.Hour*24) * math.Pow(base, i))
		LevelTime = append(LevelTime, t)
		//fmt.Printf("%s\n", formatDuration(t))
	}
}

type EntryInfo struct {
	PracticeEntry
	late float64
}

func (data *Data) getAllPracticeEntries() []EntryInfo {
	var entries []EntryInfo
	now := time.Now()
	// filter
	for _, e := range data.Practices {
		lastHistory := e.LastHistory()
		if lastHistory.Time.Add(LevelTime[lastHistory.Level]).Before(now) {
			var late float64
			if lastHistory.Level > 0 {
				late = float64(now.Sub(
					lastHistory.Time.Add(time.Duration(float64(LevelTime[lastHistory.Level])*1.1)))) /
					float64(LevelTime[lastHistory.Level])
			}
			entries = append(entries, EntryInfo{
				PracticeEntry: e,
				late:          late,
			})
		}
	}
	return entries
}

func (data *Data) PrintStat([]string) {
	entries := data.getAllPracticeEntries()
	nReviews := 0
	nLate := 0
	levelStat := make(map[int]int)
	for _, e := range entries {
		lastHistory := e.LastHistory()
		if lastHistory.Level > 0 {
			nReviews++
		}
		if e.late > 0 {
			nLate++
		}
		levelStat[lastHistory.Level]++
	}
	p("%d entries to review, %d late\n", nReviews, nLate)
	for i := 1; i < 16; i++ {
		if n := levelStat[i]; n > 0 {
			p("%d %d\n", i, n)
		}
	}
}

func (data *Data) Practice([]string) {
	entries := data.getAllPracticeEntries()
	// sort
	sort.Sort(EntrySorter(entries))

	// select
	maxWeight := 500
	maxReviewWeight := 300
	maxNewWeight := 50
	reviewWeight := 0
	newWeight := 0
	weight := 0
	var selected []EntryInfo
	for _, entry := range entries {
		if weight >= maxWeight {
			break
		}
		lastLevel := entry.LastHistory().Level
		if lastLevel == 0 && newWeight >= maxNewWeight { // new
			continue
		} else if lastLevel > 0 && reviewWeight >= maxReviewWeight { // review
			continue
		}
		selected = append(selected, entry)
		if lastLevel == 0 {
			newWeight += entry.Weight()
		} else if lastLevel > 0 {
			reviewWeight += entry.Weight()
		}
		weight += entry.Weight()
	}

	// practice
	p("%d entries to practice\n", len(selected))
	runPractice(selected, data)
}

type UI func(what string, args ...interface{})

type Input func() rune

func runeWidth(r rune) int {
	switch {
	case r >= 0x4e00 && r <= 0x9fff, // CJK综合汉字
		r >= 0x2f00 && r <= 0x2fdf,   // 部首
		r >= 0x2e80 && r <= 0x2eff,   // 部首辅助
		r >= 0x2ff0 && r <= 0x2fff,   // 汉字构成记述文字
		r >= 0x3000 && r <= 0x303f,   // CJK标点
		r >= 0x3040 && r <= 0x309f,   // 平假名
		r >= 0x30A0 && r <= 0x30ff,   // 片假名
		r >= 0x31c0 && r <= 0x31ef,   // 笔画
		r >= 0x31f0 && r <= 0x31ff,   // 片假名扩展
		r >= 0x3400 && r <= 0x4dbf,   // 汉字扩展A
		r >= 0xf900 && r <= 0xfaff,   // CJK 互换汉字
		r >= 0xff00 && r <= 0xffef,   // 全角
		r >= 0x20000 && r <= 0x2ffff, // 汉字扩展
		r >= 0xe0100 && r <= 0xe01ef, // 异体汉字
		r >= 0x1b000 && r <= 0x1b0ff, // 假名辅助
		r >= 0x30000 && r <= 0x3ffff:
		return 2
	default:
		return 1
	}
}

func runPractice(entries []EntryInfo, data *Data) {
	termbox.Init()
	defer termbox.Close()

	width, height := termbox.Size()
	printStr := func(line int, str string) {
		l := 0
		for _, r := range str {
			l += runeWidth(r)
		}
		x := (width - l) / 2
		for _, r := range str {
			termbox.SetCell(x, line, r, termbox.ColorDefault, termbox.ColorDefault)
			x += runeWidth(r)
		}
	}

	var hint, text, info string
	redraw := func() {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		printStr(height/2-2, hint)
		printStr(height/2, text)
		printStr(height-1, info)
		termbox.Flush()
	}

	var ui UI = func(what string, args ...interface{}) {
		switch what {
		case "set-hint":
			hint = args[0].(string)
		case "set-text":
			text = args[0].(string)
		case "set-info":
			info = args[0].(string)
		default:
			panic("unknown ui action")
		}
		redraw()
	}

	keys := make(chan rune)
	go func() {
		for {
			ev := termbox.PollEvent()
			switch ev.Type {
			case termbox.EventKey:
				select {
				case keys <- ev.Ch:
				default:
				}
			case termbox.EventResize:
				width = ev.Width
				height = ev.Height
				redraw()
			}
		}
	}()
	var input Input = func() rune {
		return <-keys
	}

	ui("set-hint", "press f to start")
	for {
		if input() == 'f' {
			break
		}
	}
	ui("set-hint", "")

	wg := new(sync.WaitGroup)
	save := func() {
		wg.Add(1)
		go func() {
			data.save()
			wg.Done()
		}()
	}

	// train
loop:
	for _, e := range entries {
		ui("set-hint", "")
		ui("set-text", "")
		lastHistory := e.LastHistory()
		var lateStr string
		if e.late > 0 {
			lateStr = s(" late %f", e.late)
		}
		ui("set-info", s("level %d lesson %s%s", lastHistory.Level, e.Lesson(), lateStr))
		res := e.Practice(ui, input)
		switch res {
		case LEVEL_UP:
			e.LevelUp()
			save()
		case LEVEL_RESET:
			e.LevelReset()
			save()
		case EXIT:
			break loop
		}
	}

	wg.Wait()
}

type EntrySorter []EntryInfo

func (s EntrySorter) Len() int { return len(s) }

func (s EntrySorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (self EntrySorter) Less(i, j int) bool {
	left, right := self[i], self[j]
	leftLastHistory := left.LastHistory()
	rightLastHistory := right.LastHistory()
	leftLesson := left.Lesson()
	rightLesson := right.Lesson()
	if leftLastHistory.Level == 0 {
		if rightLastHistory.Level == 0 { // new entry
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
			}
		} else {
			return false
		}
	} else {
		if rightLastHistory.Level == 0 {
			return true
		} else { // review entry
			if left.late < 0 && right.late < 0 { // both is not late
				if leftLastHistory.Level < rightLastHistory.Level { // review low level first
					return true
				} else if leftLastHistory.Level > rightLastHistory.Level {
					return false
				} else if leftLastHistory.Level == rightLastHistory.Level { // same level
					if leftLesson < rightLesson { // review earlier lesson first
						return true
					} else if leftLesson > rightLesson {
						return false
					} else { // same lesson randomize
						if rand.Intn(2) == 1 { // randomize
							return true
						}
						return false
					}
				}
			} else if left.late > 0 && right.late > 0 && leftLesson == rightLesson { // randomize same lesson
				if rand.Intn(2) == 1 {
					return true
				}
				return false
			} else {
				return left.late > right.late
			}
		}
	}
	panic("not here")
	return false
}
