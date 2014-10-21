package main

import (
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/reusee/lgtk"
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

func (data *Data) Practice([]string) {
	var entries []EntryInfo
	now := time.Now()
	// filter
	nReview := 0
	levelStat := make(map[int]int)
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
			if lastHistory.Level > 0 {
				nReview++
			}
			levelStat[lastHistory.Level]++
		}
	}
	p("%d entries to review\n", nReview)
	for i := 1; i < 16; i++ {
		if n := levelStat[i]; n > 0 {
			p("%d %d\n", i, n)
		}
	}

	// sort
	sort.Sort(EntrySorter(entries))

	// select
	maxWeight := 250
	maxReviewWeight := 200
	maxNewWeight := 80
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
	ui_gtk(selected, data)
}

type UI func(what string, args ...interface{})

type Input func() rune

func ui_gtk(entries []EntryInfo, data *Data) {
	keys := make(chan rune)
	g, err := lgtk.New(`
Gdk = lgi.Gdk

css = Gtk.CssProvider.get_default()
css:load_from_data([[
GtkWindow {
	background-color: black;
	color: white;
}
#hint {
	font-size: 16px;
}
#text {
	font-size: 48px;
	color: #0099CC;
}
#info {
	color: grey;
}
]])
Gtk.StyleContext.add_provider_for_screen(Gdk.Screen.get_default(), css, 999)

win = Gtk.Window{
	title = 'Spaced Repetition System',
	Gtk.Grid{
		orientation = 'VERTICAL',
		Gtk.Label{
			expand = true,
		},
		Gtk.Label{
			id = 'hint',
			name = 'hint',
		},
		Gtk.Label{
			id = 'text',
			name = 'text',
		},
		Gtk.Label{
			expand = true,
		},
		Gtk.Label{
			id = 'info',
			name = 'info',
		},
	},
}

function win:on_key_press_event(ev)
	Key(ev.keyval)
	return true
end
function win.on_destroy()
	Exit(0)
end
win:show_all()

	`,
		"Key", func(val rune) {
			select {
			case keys <- val:
			default:
			}
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	var ui UI = func(what string, args ...interface{}) {
		switch what {
		case "set-hint":
			g.ExecEval(`win.child.hint:set_label(T)`, "T", args[0].(string))
		case "set-text":
			g.ExecEval(`win.child.text:set_label(T)`, "T", args[0].(string))
		default:
			log.Fatalf("unknown ui action %s", what)
		}
	}

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
		g.ExecEval(`win.child.info:set_label(T)`, "T",
			s("level %d lesson %s%s", lastHistory.Level, e.Lesson(), lateStr))
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
