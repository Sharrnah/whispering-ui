package CustomWidget

// Widget from https://github.com/fyne-io/fyne-x/blob/master/widget/completionentry.go

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
)

var _ fyne.Tappable = (*CompletionEntry)(nil)
var _ fyne.DoubleTappable = (*CompletionEntry)(nil)

// CompletionEntry is an Entry with options displayed in a PopUpMenu.
type CompletionEntry struct {
	widget.Entry
	popupMenu        *widget.PopUp
	navigableList    *navigableList
	Options          []string
	OptionsTextValue []TextValueOption
	FilteredOptions  []string
	pause            bool
	itemHeight       float32
	ShowAllEntryText string

	CustomCreate func() fyne.CanvasObject
	CustomUpdate func(id widget.ListItemID, object fyne.CanvasObject)
}

// NewCompletionEntry creates a new CompletionEntry which creates a popup menu that responds to keystrokes to navigate through the items without losing the editing ability of the text input.
func NewCompletionEntry(options []string) *CompletionEntry {
	c := &CompletionEntry{Options: options, FilteredOptions: options}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CompletionEntry) SetValueOptions(valueOptions []TextValueOption) {
	for _, option := range valueOptions {
		c.OptionsTextValue = append(c.OptionsTextValue, option)
		c.Options = append(c.Options, option.Text)
	}
}

func (c *CompletionEntry) GetCurrentValueOptionEntry() *TextValueOption {
	var bestMatch *TextValueOption = nil
	for i := 0; i < len(c.OptionsTextValue); i++ {
		if strings.Contains(strings.ToLower(c.OptionsTextValue[i].Text), strings.ToLower(c.Text)) && bestMatch == nil {
			bestMatch = &c.OptionsTextValue[i]
		}
		if c.OptionsTextValue[i].Text == c.Text {
			return &c.OptionsTextValue[i]
		}
	}
	// nothing found. return best match if it exists
	if bestMatch != nil {
		return bestMatch
	}
	return nil
}

func (c *CompletionEntry) GetValueOptionEntryByText(entry string) TextValueOption {
	var bestMatch TextValueOption
	for i := 0; i < len(c.OptionsTextValue); i++ {
		if strings.Contains(strings.ToLower(c.OptionsTextValue[i].Text), strings.ToLower(entry)) && bestMatch == (TextValueOption{}) {
			bestMatch = c.OptionsTextValue[i]
		}
		if c.OptionsTextValue[i].Text == entry {
			return c.OptionsTextValue[i]
		}
	}
	// nothing found. return best match if it exists
	if bestMatch != (TextValueOption{}) {
		return bestMatch
	}
	return TextValueOption{}
}

func (c *CompletionEntry) GetValueOptionEntryByValue(entry string) TextValueOption {
	var bestMatch TextValueOption
	for i := 0; i < len(c.OptionsTextValue); i++ {
		if strings.Contains(strings.ToLower(c.OptionsTextValue[i].Value), strings.ToLower(entry)) && bestMatch == (TextValueOption{}) {
			bestMatch = c.OptionsTextValue[i]
		}
		if c.OptionsTextValue[i].Value == entry {
			return c.OptionsTextValue[i]
		}
	}
	// nothing found. return best match if it exists
	if bestMatch != (TextValueOption{}) {
		return bestMatch
	}
	return TextValueOption{}
}

// HideCompletion hides the completion menu.
func (c *CompletionEntry) HideCompletion() {
	if c.popupMenu != nil {
		c.popupMenu.Hide()
	}
}

func (c *CompletionEntry) Tapped(ev *fyne.PointEvent) {
	c.Entry.Tapped(ev)

	if c.Disabled() {
		return
	}

	if c.OnChanged != nil {
		c.OnChanged(c.Entry.Text)
	} else {
		c.ShowCompletion()
	}

	c.selectCurrentItem()

	// select all text on initial tap
	c.Entry.TypedShortcut(&fyne.ShortcutSelectAll{})
}

func (c *CompletionEntry) DoubleTapped(ev *fyne.PointEvent) {
	c.Entry.DoubleTapped(ev)

	if c.Disabled() {
		return
	}

	if c.OnChanged != nil {
		c.OnChanged(c.Entry.Text)
	} else {
		c.ShowCompletion()
	}

	c.selectCurrentItem()
}

// Move changes the relative position of the select entry.
//
// Implements: fyne.Widget
func (c *CompletionEntry) Move(pos fyne.Position) {
	c.Entry.Move(pos)
	if c.popupMenu != nil {
		c.popupMenu.Resize(c.maxSize())
		c.popupMenu.Move(c.popUpPos())
	}
}

// Refresh the list to update the options to display.
func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
	if c.navigableList != nil {
		c.navigableList.SetOptions(c.FilteredOptions)
	}
}

// SetOptions set the completion list with itemList and update the view.
func (c *CompletionEntry) SetOptions(itemList []string) {
	c.Options = itemList
	c.FilteredOptions = itemList
	c.Refresh()
}

func (c *CompletionEntry) SetOptionsFilter(itemList []string) {
	c.FilteredOptions = itemList
	if c.ShowAllEntryText != "" && len(c.FilteredOptions) < len(c.Options) {
		c.FilteredOptions = append(c.FilteredOptions, c.ShowAllEntryText)
	}
	c.Refresh()
}

func (c *CompletionEntry) ResetOptionsFilter() {
	c.FilteredOptions = c.Options

	c.Refresh()
}

// ShowCompletion displays the completion menu
func (c *CompletionEntry) ShowCompletion() {
	if c.pause {
		return
	}
	if len(c.FilteredOptions) == 0 {
		c.HideCompletion()
		return
	}

	if c.navigableList == nil {
		c.navigableList = newNavigableList(c.FilteredOptions, &c.Entry, c.setTextFromMenu, c.HideCompletion,
			c.CustomCreate, c.CustomUpdate)
	} else {
		c.navigableList.UnselectAll()
		c.navigableList.selected = -1
	}
	holder := fyne.CurrentApp().Driver().CanvasForObject(c)

	if c.popupMenu == nil {
		c.popupMenu = widget.NewPopUp(c.navigableList, holder)
	}
	c.popupMenu.Resize(c.maxSize())
	c.popupMenu.ShowAtPosition(c.popUpPos())
	holder.Focus(c.navigableList)
}

// calculate the max size to make the popup to cover everything below the entry
func (c *CompletionEntry) maxSize() fyne.Size {
	cnv := fyne.CurrentApp().Driver().CanvasForObject(c)

	if c.itemHeight == 0 {
		// set item height to cache
		c.itemHeight = c.navigableList.CreateItem().MinSize().Height
	}

	listheight := float32(len(c.FilteredOptions))*(c.itemHeight+2*theme.Padding()+theme.SeparatorThicknessSize()) + 2*theme.Padding()
	canvasSize := fyne.Size{
		Width:  0,
		Height: 200,
	}
	// try to fall back if cnv is nil for some reason...
	if cnv != nil {
		canvasSize = cnv.Size()
	} else if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
		canvasSize = fyne.CurrentApp().Driver().AllWindows()[0].Content().Size()
	}
	entrySize := c.Size()
	if canvasSize.Height > listheight {
		return fyne.NewSize(entrySize.Width, listheight)
	}

	return fyne.NewSize(
		entrySize.Width,
		canvasSize.Height-c.Position().Y-entrySize.Height-theme.InputBorderSize()-theme.Padding())
}

// calculate where the popup should appear
func (c *CompletionEntry) popUpPos() fyne.Position {
	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	return entryPos.Add(fyne.NewPos(0, c.Size().Height))
}

func (c *CompletionEntry) selectCurrentItem() {
	if c.navigableList == nil {
		return
	}
	c.navigableList.navigating = true
	c.navigableList.UnselectAll()
	c.navigableList.selected = -1

	for i := 0; i < len(c.navigableList.items); i++ {
		if c.navigableList.items[i] == c.Entry.Text {
			c.navigableList.ScrollToBottom()
			c.navigableList.Select(i)
			break
		}
	}
	c.navigableList.navigating = false
}

func (c *CompletionEntry) SelectItemByValue(s string) {
	c.navigableList.navigating = true
	c.navigableList.UnselectAll()
	c.navigableList.selected = -1

	for i := 0; i < len(c.navigableList.items); i++ {
		if c.navigableList.items[i] == s {
			c.navigableList.ScrollToBottom()
			c.navigableList.ScrollTo(i)
			break
		}
	}
}

// Prevent the menu to open when the user validate value from the menu.
func (c *CompletionEntry) setTextFromMenu(s string) {
	c.pause = true
	// reset filter on selection of ShowAllEntryText
	if c.ShowAllEntryText != "" && (s == "" || s == c.ShowAllEntryText) {
		c.ResetOptionsFilter()

		c.popupMenu.Resize(c.maxSize())
		c.selectCurrentItem()

		c.pause = false
		return
	}
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.pause = false

	c.popupMenu.Hide()

	if c.Entry.OnSubmitted != nil {
		c.Entry.OnSubmitted(s)
	}
}

type navigableList struct {
	widget.List
	entry           *widget.Entry
	selected        int
	setTextFromMenu func(string)
	hide            func()
	navigating      bool
	items           []string

	customCreate func() fyne.CanvasObject
	customUpdate func(id widget.ListItemID, object fyne.CanvasObject)
}

func newNavigableList(items []string, entry *widget.Entry, setTextFromMenu func(string), hide func(),
	create func() fyne.CanvasObject, update func(id widget.ListItemID, object fyne.CanvasObject)) *navigableList {
	n := &navigableList{
		entry:           entry,
		selected:        -1,
		setTextFromMenu: setTextFromMenu,
		hide:            hide,
		items:           items,
		customCreate:    create,
		customUpdate:    update,
	}

	n.List = widget.List{
		Length: func() int {
			return len(n.items)
		},
		CreateItem: func() fyne.CanvasObject {
			if fn := n.customCreate; fn != nil {
				return fn()
			}
			return widget.NewLabel("")
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			if fn := n.customUpdate; fn != nil {
				fn(i, o)
				return
			}
			if len(n.items) > i {
				o.(*widget.Label).SetText(n.items[i])
			}
		},
		OnSelected: func(id widget.ListItemID) {
			if !n.navigating && id > -1 {
				setTextFromMenu(n.items[id])
			}
			n.navigating = false
		},
	}
	n.ExtendBaseWidget(n)
	return n
}

// Implements: fyne.Focusable
func (n *navigableList) FocusGained() {
}

// Implements: fyne.Focusable
func (n *navigableList) FocusLost() {
}

func (n *navigableList) SetOptions(items []string) {
	n.Unselect(n.selected)
	n.items = items
	n.Refresh()
	n.selected = -1
}

func (n *navigableList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.items)-1 {
			n.selected++
		} else {
			n.selected = 0
		}
		n.navigating = true
		n.Select(n.selected)

	case fyne.KeyUp:
		if n.selected > 0 {
			n.selected--
		} else {
			n.selected = len(n.items) - 1
		}
		n.navigating = true
		n.Select(n.selected)
	case fyne.KeyReturn, fyne.KeyEnter:
		if n.selected == -1 { // so the user want to submit the entry
			n.hide()
			n.entry.TypedKey(event)
		} else {
			n.navigating = false
			n.OnSelected(n.selected)
		}
	case fyne.KeyEscape:
		n.hide()
	default:
		n.entry.TypedKey(event)

	}
}

func (n *navigableList) TypedRune(r rune) {
	n.entry.TypedRune(r)
}
