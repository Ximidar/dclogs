package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

type LogBox struct {
	Log *tview.TextView
}

func NewLogBox() *LogBox {
	lb := new(LogBox)

	lb.Log = tview.NewTextView().
		SetDynamicColors(true)
	lb.Log.SetBorder(true)

	return lb
}

func (lb *LogBox) Update(title string) {
	lb.Log.SetTitle(title)
	lb.Log.Clear()
}

func (lb *LogBox) WriteText(text string) {
	if text[len(text)-1] != '\n' {
		text += "\n"
	}
	fmt.Fprint(lb.Log, text)
}
