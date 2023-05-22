package ui

import (
	"github.com/rivo/tview"
)

type LogController interface {
	SelectCallback(message string)
	GetChannel() chan string
}

type DCUI struct {
	application   *tview.Application
	flex          *tview.Flex
	LogSelector   *LogSelector
	LogBox        *LogBox
	logController LogController
	LogStream     chan string
}

func CreateUI(logController LogController) *DCUI {

	dcui := new(DCUI)
	dcui.logController = logController

	// current log stream
	dcui.LogStream = dcui.logController.GetChannel()

	// New application
	dcui.application = tview.NewApplication()

	// Flex box
	dcui.flex = tview.NewFlex()
	dcui.flex.Box = tview.NewBox().
		SetBorder(true).
		SetTitle("Docker Compose Logs")

	// Node Selector
	logSelector := NewLogSelector()
	dcui.LogSelector = logSelector
	dcui.LogSelector.SetController(dcui.logController)
	dcui.flex.AddItem(dcui.LogSelector.tree, 0, 1, true)

	// Log View
	dcui.LogBox = NewLogBox()
	dcui.LogSelector.AddSelectHandler(dcui.LogBox.Update)
	dcui.flex.AddItem(dcui.LogBox.Log, 0, 6, false)

	dcui.application.SetRoot(dcui.flex, true)

	return dcui

}

func (ui *DCUI) consumeLogStream() {

	for msg := range ui.LogStream {
		ui.LogBox.WriteText(msg)
		ui.application.Draw()

	}
}

func (ui *DCUI) Start() {
	go ui.consumeLogStream()

	if err := ui.application.Run(); err != nil {
		panic(err)
	}
}
