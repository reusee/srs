package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/reusee/lgtk"
)

func ui_gtk(entries []*Entry, data *Data) {
	// ui
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

	setHint := func(s string) {
		g.ExecEval(`win.child.hint:set_label(T)`, "T", s)
	}
	setText := func(s string) {
		g.ExecEval(`win.child.text:set_label(T)`, "T", s)
	}
	setLevel := func(s string) {
		g.ExecEval(`win.child.level:set_label(T)`, "T", s)
	}

	setHint("press f to start")
	for {
		key := <-keys
		if key == 'f' {
			break
		}
	}
	setHint("")

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
		setHint("")
		setText("")

		lastHistory := e.History[len(e.History)-1]
		setLevel(strconv.Itoa(lastHistory.Level))

		switch entry := e.IsEntry.(type) {
		case *AudioToWordEntry:
			setHint("playing...")
			playAudio(entry.word.AudioFile)
			setHint("press any key to show answer")
			<-keys
			setText(entry.word.Text)
		repeat:
			setHint("press G to levelup, T to reset level, Space to repeat")
		read_key:
			key := <-keys
			switch key {
			case 'g':
				e.History = append(e.History, HistoryEntry{Level: lastHistory.Level + 1, Time: time.Now()})
				save()
			case 't':
				e.History = append(e.History, HistoryEntry{Level: 0, Time: time.Now()})
				save()
			case ' ':
				setHint("playing...")
				playAudio(entry.word.AudioFile)
				setHint("")
				goto repeat
			case 'q':
				setText("")
				setHint("exit...")
				break loop
			default:
				goto read_key
			}

		case *WordToAudioEntry:
			setText(entry.word.Text)
			setHint("press any key to play audio")
			<-keys
		repeat2:
			setHint("playing...")
			playAudio(entry.word.AudioFile)
			setHint("press G to levelup, T to reset level, Space to repeat")
		read_key2:
			key := <-keys
			switch key {
			case 'g':
				e.History = append(e.History, HistoryEntry{Level: lastHistory.Level + 1, Time: time.Now()})
				save()
			case 't':
				e.History = append(e.History, HistoryEntry{Level: 0, Time: time.Now()})
				save()
			case ' ':
				goto repeat2
			case 'q':
				setText("")
				setHint("exit...")
				break loop
			default:
				goto read_key2
			}

		case *SentenceEntry:
			setHint("playing...")
			playAudio(entry.AudioFile)
		repeat3:
			setHint("press G to levelup, T to reset level, Space to repeat")
		read_key3:
			key := <-keys
			switch key {
			case 'g':
				e.History = append(e.History, HistoryEntry{Level: lastHistory.Level + 1, Time: time.Now()})
				save()
			case 't':
				e.History = append(e.History, HistoryEntry{Level: 0, Time: time.Now()})
				save()
			case ' ':
				setHint("playing...")
				playAudio(entry.AudioFile)
				setHint("")
				goto repeat3
			case 'q':
				setText("")
				setHint("exit...")
				break loop
			default:
				goto read_key3
			}

		default:
			panic("fixme")
		}

	}

	wg.Wait()
}
