package main

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/reusee/lgtk"
)

func (data *Data) Practice([]string) {
	var entries []*Entry
	now := time.Now()
	for _, e := range data.Entries {
		lastHistory := e.History[len(e.History)-1]
		if lastHistory.Time.Add(LevelTime[lastHistory.Level]).Before(now) {
			entries = append(entries, e)
		}
	}
	sort.Sort(EntrySorter(entries))
	max := 25
	if len(entries) > max {
		entries = entries[:max]
	}
	ui_gtk(entries, data)
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
