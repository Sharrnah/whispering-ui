package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"strings"
	"sync"
)

var _ fyne.Widget = (*EntryWithPopupMenu)(nil)
var _ fyne.Tabbable = (*EntryWithPopupMenu)(nil)
var _ fyne.Focusable = (*EntryWithPopupMenu)(nil)

type HotKeyEntryShortcuts struct {
	Name    string
	Handler func()
}

type HotKeyEntry struct {
	widget.Entry
	dirty        bool
	text         *widget.RichText
	impl         fyne.Widget
	propertyLock sync.RWMutex

	popUp                   *widget.PopUpMenu
	additionalMenuItems     []*fyne.MenuItem
	shortcutSubmitFunctions []HotKeyEntryShortcuts

	OnFocusChanged func(bool)

	runeKey     string
	modifierKey string
}

func (e *HotKeyEntry) superWidget() fyne.Widget {
	return e
}

func NewHotKeyEntry() *HotKeyEntry {
	e := &HotKeyEntry{}
	e.MultiLine = false
	e.ExtendBaseWidget(e)
	return e
}

func (e *HotKeyEntry) requestFocus() {
	impl := e.superWidget()
	if c := fyne.CurrentApp().Driver().CanvasForObject(impl); c != nil {
		c.Focus(impl.(fyne.Focusable))
	}
}

func (e *HotKeyEntry) AddAdditionalMenuItem(menuItem *fyne.MenuItem) {
	e.additionalMenuItems = append(e.additionalMenuItems, menuItem)
}

func (e *HotKeyEntry) TappedSecondary(pe *fyne.PointEvent) {
	if e.Disabled() && e.Password {
		return // no popup options for a disabled concealed field
	}

	e.requestFocus()
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	super := e.superWidget()

	cutItem := fyne.NewMenuItem(lang.L("Cut"), func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
	})
	copyItem := fyne.NewMenuItem(lang.L("Copy"), func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
	})
	pasteItem := fyne.NewMenuItem(lang.L("Paste"), func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
	})
	selectAllItem := fyne.NewMenuItem(lang.L("Select all"), func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutSelectAll{})
	})

	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(e)
	popUpPos := entryPos.Add(fyne.NewPos(pe.Position.X, pe.Position.Y))
	c := fyne.CurrentApp().Driver().CanvasForObject(e)

	var menu *fyne.Menu
	if e.Disabled() {
		menu = fyne.NewMenu("", copyItem, selectAllItem)
	} else if e.Password {
		menu = fyne.NewMenu("", pasteItem, selectAllItem)
	} else {
		var menuItems []*fyne.MenuItem
		menuItems = append(menuItems, cutItem, copyItem, pasteItem, selectAllItem)
		menu = fyne.NewMenu("", menuItems...)
	}

	// add additional menu items
	if len(e.additionalMenuItems) > 0 {
		menu.Items = append(menu.Items, fyne.NewMenuItemSeparator())
		menu.Items = append(menu.Items, e.additionalMenuItems...)
	}
	e.popUp = widget.NewPopUpMenu(menu, c)
	e.popUp.ShowAtPosition(popUpPos)
}

func (e *HotKeyEntry) TypedRune(r rune) {
	runeString := string(r)
	if r == '+' {
		runeString = "plus"
	}
	if r == ' ' {
		runeString = "space"
	}
	e.runeKey = runeString
	e.UpdateHotkeyValue()
	return
}

func (e *HotKeyEntry) TypedKey(key *fyne.KeyEvent) {
	if e.Disabled() {
		return
	}

	switch key.Name {
	case fyne.KeyEscape:
		e.SetText("")
		e.runeKey = ""
		e.modifierKey = ""
		return
	case "LeftControl", "RightControl":
		e.modifierKey = "ctrl"
	case "LeftShift", "RightShift":
		e.modifierKey = "shift"
	case "LeftAlt", "RightAlt":
		e.modifierKey = "alt"
	case "LeftSuper", "RightSuper":
		e.modifierKey = "windows"
	}

	e.UpdateHotkeyValue()
}

func (e *HotKeyEntry) UpdateHotkeyValue() {
	entryText := &strings.Builder{}

	if e.modifierKey != "" {
		entryText.WriteString(e.modifierKey)
	}
	if e.runeKey != "" && e.modifierKey != "" {
		entryText.WriteString("+")
	}
	if e.runeKey != "" {
		entryText.WriteString(e.runeKey)
	}
	e.SetText(entryText.String())
}

func (e *HotKeyEntry) FocusLost() {
	if e.OnFocusChanged != nil {
		e.OnFocusChanged(false)
	}
	e.Entry.FocusLost()
}

func (e *HotKeyEntry) FocusGained() {
	if e.OnFocusChanged != nil {
		e.OnFocusChanged(true)
	}
	e.Entry.FocusGained()
}
