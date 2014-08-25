package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/reusee/lgtk"
)

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
#level {
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
			id = 'level',
			name = 'level',
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
		g.ExecEval(`win.child.level:set_label(T)`, "T", strconv.Itoa(lastHistory.Level))
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

func (e *AudioToWordEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press any key to show answer")
	input()
	ui("set-text", e.word.Text)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		ui("set-hint", "playing...")
		playAudio(e.word.AudioFile)
		ui("set-hint", "")
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}

func (e *WordToAudioEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-text", e.word.Text)
	ui("set-hint", "press any key to play audio")
	input()
repeat:
	ui("set-hint", "playing...")
	playAudio(e.word.AudioFile)
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}

func (e *SentenceEntry) Practice(ui UI, input Input) PracticeResult {
	ui("set-hint", "playing...")
	playAudio(e.AudioFile)
repeat:
	ui("set-hint", "press G to levelup, T to reset level, Space to repeat")
read_key:
	key := input()
	switch key {
	case 'g':
		return LEVEL_UP
	case 't':
		return LEVEL_RESET
	case ' ':
		ui("set-hint", "playing...")
		playAudio(e.AudioFile)
		ui("set-hint", "")
		goto repeat
	case 'q':
		ui("set-hint", "exit...")
		return EXIT
	default:
		goto read_key
	}
	return NONE
}
