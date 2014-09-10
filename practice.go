package main

import (
	"fmt"
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
	base := 2.25
	for i := 0.0; i < 12; i++ {
		t := time.Duration(float64(time.Hour*24) * math.Pow(base, i))
		LevelTime = append(LevelTime, t)
		//fmt.Printf("%s\n", formatDuration(t))
	}
}

func formatDuration(duration time.Duration) string {
	var ret string
	var m, h, d, y time.Duration
	m = duration / time.Minute
	if m >= 60 {
		h = m / 60
		m = m % 60
	}
	if h >= 24 {
		d = h / 24
		h = h % 24
	}
	if d > 365 {
		y = d / 365
		d = d % 365
	}
	if y > 0 {
		ret += fmt.Sprintf("%dyears.", y)
	}
	if d > 0 {
		ret += fmt.Sprintf("%ddays.", d)
	}
	if h > 0 {
		ret += fmt.Sprintf("%dhours.", h)
	}
	if m > 0 {
		ret += fmt.Sprintf("%dmins.", m)
	}
	return ret
}

func (data *Data) Practice([]string) {
	var entries []PracticeEntry
	now := time.Now()
	// filter
	nReview := 0
	for _, e := range data.Practices {
		lastHistory := e.LastHistory()
		if lastHistory.Time.Add(LevelTime[lastHistory.Level]).Before(now) {
			entries = append(entries, e)
			if lastHistory.Level > 0 {
				nReview++
			}
		}
	}
	p("%d entries to review\n", nReview)
	// sort
	sort.Sort(EntrySorter(entries))
	// select
	maxWeight := 250
	maxReviewWeight := 220
	maxNewWeight := 50
	reviewWeight := 0
	newWeight := 0
	weight := 0
	var selected []PracticeEntry
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
	ui_gtk(selected, data)
}

type UI func(what string, args ...interface{})

type Input func() rune

func ui_gtk(entries []PracticeEntry, data *Data) {
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
		g.ExecEval(`win.child.info:set_label(T)`, "T",
			s("level %d lesson %s", lastHistory.Level, e.Lesson()))
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

type EntrySorter []PracticeEntry

func (s EntrySorter) Len() int { return len(s) }

func (s EntrySorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (self EntrySorter) Less(i, j int) bool {
	left, right := self[i], self[j]
	leftLastHistory := left.LastHistory()
	rightLastHistory := right.LastHistory()
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
		}
	}
	return false
}

func (s EntrySorter) getLevelOrder(e PracticeEntry) int {
	lastHistory := e.LastHistory()
	if lastHistory.Level > 0 {
		return 1
	} else {
		return 2
	}
}
