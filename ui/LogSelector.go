package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LogSelector struct {
	Services          []string
	SelectedService   string
	SelectedContainer string
	tree              *tview.TreeView
	root              *tview.TreeNode
	selected          *tview.TreeNode
	controller        LogController
	selectHandlers    []func(node string)
}

func NewLogSelector() *LogSelector {
	ls := new(LogSelector)

	// Build the tree
	ls.tree = tview.NewTreeView()
	ls.tree.SetBorder(true)
	ls.tree.SetSelectedFunc(ls.SelectEventHandler)

	ls.controller = nil

	return ls
}

func (ls *LogSelector) SetRoot(title string) {
	ls.root = tview.NewTreeNode(title)
	ls.tree.
		SetRoot(ls.root).
		SetCurrentNode(ls.root)
}

func (ls *LogSelector) SetController(controller LogController) {
	ls.controller = controller
}

func (ls *LogSelector) AddNode(title string, children ...string) {

	node := tview.NewTreeNode(title)
	node.SetReference(title)

	if len(children) != 0 {
		node.SetColor(tcell.ColorGreen)
		node.SetExpanded(false)
	}

	for _, child := range children {
		child_node := tview.NewTreeNode(child)
		path := fmt.Sprintf("%s/%s", title, child)
		child_node.SetReference(path)
		node.AddChild(child_node)
	}

	ls.root.AddChild(node)

}

func (ls *LogSelector) AddSelectHandler(handler func(node string)) {
	ls.selectHandlers = append(ls.selectHandlers, handler)
}

func (ls *LogSelector) SelectEventHandler(node *tview.TreeNode) {
	reference := node.GetReference().(string)
	if reference == "" {
		return
	}

	children := node.GetChildren()
	if len(children) != 0 {
		node.SetExpanded(!node.IsExpanded())
	}

	ls.selected = node
	ls.SelectedContainer = reference

	if ls.controller != nil {
		ls.controller.SelectCallback(reference)
	}

	for _, handler := range ls.selectHandlers {
		handler(reference)
	}
}
