package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"sync"
)

var _ fyne.Widget = (*EntryWithPopupMenu)(nil)
var _ fyne.Tabbable = (*EntryWithPopupMenu)(nil)

type EntryWithPopupMenu struct {
	widget.Entry
	dirty        bool
	text         *widget.RichText
	impl         fyne.Widget
	propertyLock sync.RWMutex

	popUp               *widget.PopUpMenu
	additionalMenuItems []*fyne.MenuItem
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

func (e *EntryWithPopupMenu) requestFocus() {
	impl := e.superWidget()
	if c := fyne.CurrentApp().Driver().CanvasForObject(impl); c != nil {
		c.Focus(impl.(fyne.Focusable))
	}
}

func (e *EntryWithPopupMenu) AddAdditionalMenuItem(menuItem *fyne.MenuItem) {
	e.additionalMenuItems = append(e.additionalMenuItems, menuItem)
}

func (e *EntryWithPopupMenu) TappedSecondary(pe *fyne.PointEvent) {
	if e.Disabled() && e.Password {
		return // no popup options for a disabled concealed field
	}

	e.requestFocus()
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	super := e.superWidget()

	cutItem := fyne.NewMenuItem("Cut", func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
	})
	selectAllItem := fyne.NewMenuItem("Select all", func() {
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
		menu = fyne.NewMenu("", cutItem, copyItem, pasteItem, selectAllItem)
	}

	// add additional menu items
	if len(e.additionalMenuItems) > 0 {
		menu.Items = append(menu.Items, fyne.NewMenuItemSeparator())
		menu.Items = append(menu.Items, e.additionalMenuItems...)
	}
	e.popUp = widget.NewPopUpMenu(menu, c)
	e.popUp.ShowAtPosition(popUpPos)
}
