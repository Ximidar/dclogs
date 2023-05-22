package ui

import (
	"github.com/rivo/tview"
)

func CreateUI() {
	box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
}

func list_services() {
	// This will create a box that lists available services to
	// view the log from
}

func createLogBox() {
	// This will create the log box that will show the currently selected logs
}
