package ui

type LogSelector struct {
	Services        []string
	SelectedService string

	// add a callback we can call to change logs
	// add a variable for holding the ui screen
}

func NewLogSelector() *LogSelector {
	ls := new(LogSelector)

	return ls
}

func (ls *LogSelector) drawBox() {

}

func (ls *LogSelector) ClickEventHandler() {
	// Is it in the box?

	// does it select a log

	// change selected log
}
