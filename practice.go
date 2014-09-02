package main

import (
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/reusee/lgtk"
)

func (data *Data) Practice([]string) {
	var entries []*Entry
	now := time.Now()
	// filter
	for _, e := range data.Entries {
		lastHistory := e.History[len(e.History)-1]
		if lastHistory.Time.Add(LevelTime[lastHistory.Level]).Before(now) {
			entries = append(entries, e)
			if lastHistory.Level > 0 {
			}
		}
	}
	// sort
	sort.Sort(EntrySorter(entries))
	// select
	max := 25
	maxReview := 20
	maxNew := 8
	nReview := 0
	nNew := 0
	var selected []*Entry
	for _, entry := range entries {
		if len(selected) >= max {
			break
		}
		lastLevel := entry.History[len(entry.History)-1].Level
		if lastLevel == 0 && nNew >= maxNew { // new
			continue
		} else if lastLevel > 0 && nReview >= maxReview { // review
			continue
		}
		selected = append(selected, entry)
		if lastLevel == 0 {
			nNew++
		} else if lastLevel > 0 {
			nReview++
		}
	}
	// practice
	ui_gtk(selected, data)
}

type UI func(what string, args ...interface{})

type Input func() rune

func ui_gtk(entries []*Entry, data *Data) {
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
		lastHistory := e.History[len(e.History)-1]
		g.ExecEval(`win.child.info:set_label(T)`, "T",
			s("level %d lesson %s", lastHistory.Level, e.Lesson()))
		res := e.Practice(ui, input)
		switch res {
		case LEVEL_UP:
			e.History = append(e.History, HistoryEntry{Level: lastHistory.Level + 1, Time: time.Now()})
			save()
		case LEVEL_RESET:
			e.History = append(e.History, HistoryEntry{Level: 0, Time: time.Now()})
			save()
		case EXIT:
			break loop
		}
	}

	wg.Wait()
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
	if lastHistory.Level > 0 {
		return 1
	} else {
		return 2
	}
}
