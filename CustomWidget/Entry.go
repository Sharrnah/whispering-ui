package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"runtime"
	"strings"
)

var _ fyne.Widget = (*EntryWithPopupMenu)(nil)
var _ fyne.Tabbable = (*EntryWithPopupMenu)(nil)

type EntryWithPopupMenuShortcuts struct {
	Name    string
	Handler func()
}

type EntryWithPopupMenu struct {
	widget.Entry
	impl fyne.Widget

	popUp                   *widget.PopUpMenu
	additionalMenuItems     []*fyne.MenuItem
	shortcutSubmitFunctions []EntryWithPopupMenuShortcuts
}

func (e *EntryWithPopupMenu) superWidget() fyne.Widget {
	return e
}

func NewMultiLineEntry() *EntryWithPopupMenu {
	e := &EntryWithPopupMenu{}
	e.MultiLine = true
	e.Wrapping = fyne.TextTruncate
	e.ExtendBaseWidget(e)
	return e
}
func NewMultiLineEntryWithData(data binding.String) *EntryWithPopupMenu {
	entry := NewMultiLineEntry()
	entry.Bind(data)
	return entry
}
func (e *EntryWithPopupMenu) Bind(data binding.String) {
	e.Entry.Bind(data)
}
func (e *EntryWithPopupMenu) Unbind() {
	e.Entry.Unbind()
}

func (e *EntryWithPopupMenu) requestFocus() {
	impl := e.superWidget()
	if c := fyne.CurrentApp().Driver().CanvasForObject(impl); c != nil {
		c.Focus(impl.(fyne.Focusable))
	}
}

func (e *EntryWithPopupMenu) AddAdditionalMenuItem(menuItem *fyne.MenuItem) {
	e.additionalMenuItems = append(e.additionalMenuItems, menuItem)
}
func (e *EntryWithPopupMenu) RemoveAdditionalMenuItem(menuItem *fyne.MenuItem) {
	for i, item := range e.additionalMenuItems {
		if item == menuItem {
			e.additionalMenuItems = append(e.additionalMenuItems[:i], e.additionalMenuItems[i+1:]...)
			return
		}
	}
}
func (e *EntryWithPopupMenu) FindAdditionalMenuItemByLabel(label string) *fyne.MenuItem {
	for _, item := range e.additionalMenuItems {
		if item.Label == label {
			return item
		}
	}
	return nil
}

func (e *EntryWithPopupMenu) TappedSecondary(pe *fyne.PointEvent) {
	if e.Disabled() && e.Password {
		return // no popup options for a disabled concealed field
	}

	e.requestFocus()
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	super := e.superWidget()

	//undoItem := fyne.NewMenuItem(lang.L("Undo"), e.Undo)
	//redoItem := fyne.NewMenuItem(lang.L("Redo"), e.Redo)
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

type ShortcutEntrySubmit struct {
	fyne.KeyName
	Modifier fyne.KeyModifier
	Handler  func()
}

var _ fyne.Shortcut = (*ShortcutEntrySubmit)(nil)
var _ fyne.KeyboardShortcut = (*ShortcutEntrySubmit)(nil)

func modifierToString(mods fyne.KeyModifier) string {
	s := []string{}
	if (mods & fyne.KeyModifierShift) != 0 {
		s = append(s, string("Shift"))
	}
	if (mods & fyne.KeyModifierControl) != 0 {
		s = append(s, string("Control"))
	}
	if (mods & fyne.KeyModifierAlt) != 0 {
		s = append(s, string("Alt"))
	}
	if (mods & fyne.KeyModifierSuper) != 0 {
		if runtime.GOOS == "darwin" {
			s = append(s, string("Command"))
		} else {
			s = append(s, string("Super"))
		}
	}
	return strings.Join(s, "+")
}

func (s ShortcutEntrySubmit) ShortcutName() string {
	id := &strings.Builder{}
	id.WriteString("CustomDesktop:")
	id.WriteString(modifierToString(s.Modifier))
	id.WriteString("+")
	id.WriteString(string(s.KeyName))
	return id.String()
}

func (s ShortcutEntrySubmit) Key() fyne.KeyName {
	return s.KeyName
}

func (s ShortcutEntrySubmit) Mod() fyne.KeyModifier {
	return s.Modifier
}

func (e *EntryWithPopupMenu) AddCustomShortcut(shortcutEntrySubmit ShortcutEntrySubmit) {
	addToSlice := true
	for _, shortcut := range e.shortcutSubmitFunctions {
		if shortcut.Name == shortcutEntrySubmit.ShortcutName() {
			addToSlice = false
			break
		}
	}
	if addToSlice {
		e.shortcutSubmitFunctions = append(e.shortcutSubmitFunctions, EntryWithPopupMenuShortcuts{Name: shortcutEntrySubmit.ShortcutName(), Handler: shortcutEntrySubmit.Handler})
		e.TypedShortcut(shortcutEntrySubmit)
	}
}

func (e *EntryWithPopupMenu) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		e.Entry.TypedShortcut(s)
		return
	}

	// if the shortcut is a submit shortcut, call the submit function
	if _, ok := s.(*ShortcutEntrySubmit); !ok {
		for _, shortcut := range e.shortcutSubmitFunctions {
			if shortcut.Name == s.ShortcutName() && shortcut.Handler != nil {
				println("shortcut triggered:", s.ShortcutName())
				shortcut.Handler()
				return
			}
		}
	}
}

func (e *EntryWithPopupMenu) GetPopup() *widget.PopUpMenu {
	return e.popUp
}
