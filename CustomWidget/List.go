package CustomWidget

// You could always make a real lightweight custom widget that uses SimpleRenderer that returns the container - this wrapping it all in a widget that could be tappable

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"log"
)

type ListItemID = int
type List struct {
	widget.List

	popUp *widget.PopUpMenu
	impl  fyne.Widget
}

func NewList(length func() int, createItem func() fyne.CanvasObject, updateItem func(ListItemID, fyne.CanvasObject)) *List {
	list := &List{}
	list.BaseWidget = widget.BaseWidget{}
	list.Length = length
	list.CreateItem = createItem
	list.UpdateItem = updateItem
	list.ExtendBaseWidget(list)

	return list
}
func NewListWithData(data binding.DataList, createItem func() fyne.CanvasObject, updateItem func(binding.DataItem, fyne.CanvasObject)) *List {
	l := NewList(
		data.Length,
		createItem,
		func(i ListItemID, o fyne.CanvasObject) {
			item, err := data.GetItem(i)
			if err != nil {
				fyne.LogError(fmt.Sprintf("Error getting data item %d", i), err)
				return
			}
			updateItem(item, o)
		})

	data.AddListener(binding.NewDataListener(l.Refresh))
	return l
}
func (t *List) TappedSecondary(pe *fyne.PointEvent) {
	cutItem := fyne.NewMenuItem(lang.L("Cut"), func() {
		log.Println("CUTTED")
	})

	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(t.impl)
	popUpPos := entryPos.Add(fyne.NewPos(pe.Position.X, pe.Position.Y))

	c := fyne.CurrentApp().Driver().CanvasForObject(t.impl)

	var menu *fyne.Menu
	menu = fyne.NewMenu("", cutItem)

	t.popUp = widget.NewPopUpMenu(menu, c)
	t.popUp.ShowAtPosition(popUpPos)
}
