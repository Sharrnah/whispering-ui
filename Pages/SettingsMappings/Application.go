package SettingsMappings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/Pages/Advanced"
	"whispering-tiger-ui/UpdateUtility"
)

var ApplicationSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Start downloads using the UI. (Recommended)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("DisableUiDownloads", !b)
				})
				widgetCheckbox.Checked = !fyne.CurrentApp().Preferences().BoolWithFallback("DisableUiDownloads", false)

				return widgetCheckbox
			},
		},
		{
			SettingsName:         "Run Python backend with UTF-8 encoding. (Recommended)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("RunWithUTF8", b)
				})
				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("RunWithUTF8", true)

				return widgetCheckbox
			},
		},
		{
			SettingsName:         "Focus window on message receive. (Can improve speed in VR)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("AutoRefocusWindow", b)
				})
				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("AutoRefocusWindow", false)

				return widgetCheckbox
			},
		},
		{
			SettingsName:         "Check for App updates at startup",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", nil)
				widgetCheckbox.OnChanged = func(b bool) {
					if b {
						fyne.CurrentApp().Preferences().SetBool("CheckForUpdateAtStartup", true)
					} else {
						dialog.ShowConfirm("Disable update check", "Are you sure you want to disable App update checks at startup?", func(b bool) {
							if b {
								fyne.CurrentApp().Preferences().SetBool("CheckForUpdateAtStartup", false)
							} else {
								widgetCheckbox.SetChecked(true)
							}
						}, fyne.CurrentApp().Driver().AllWindows()[0])
					}
				}

				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("CheckForUpdateAtStartup", true)

				checkForUpdatesButton := widget.NewButton("Check for App updates now", func() {
					if !UpdateUtility.VersionCheck(fyne.CurrentApp().Driver().AllWindows()[0], true) {
						dialog.ShowInformation("No update available", "You are running the latest version of Whispering Tiger.", fyne.CurrentApp().Driver().AllWindows()[0])
					}
				})

				return container.NewHBox(widgetCheckbox, checkForUpdatesButton)
			},
		},
		{
			SettingsName:         "Check for Plugin updates at startup",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", nil)
				widgetCheckbox.OnChanged = func(b bool) {
					if b {
						fyne.CurrentApp().Preferences().SetBool("CheckForPluginUpdatesAtStartup", true)
					} else {
						dialog.ShowConfirm("Disable update check", "Are you sure you want to disable Plugin update checks at startup?", func(b bool) {
							if b {
								fyne.CurrentApp().Preferences().SetBool("CheckForPluginUpdatesAtStartup", false)
							} else {
								widgetCheckbox.SetChecked(true)
							}
						}, fyne.CurrentApp().Driver().AllWindows()[0])
					}
				}

				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("CheckForPluginUpdatesAtStartup", true)

				checkForUpdatesButton := widget.NewButton("Check for Plugin updates now", func() {
					if UpdateUtility.PluginsUpdateAvailable() {
						dialog.ShowConfirm("New Plugin updates available", "Whispering Tiger has new Plugin updates available. Go to Plugin List now?", func(b bool) {
							if b {
								Advanced.CreatePluginListWindow(nil, true)
							}
						}, fyne.CurrentApp().Driver().AllWindows()[0])
					}
				})

				return container.NewHBox(widgetCheckbox, checkForUpdatesButton)
			},
		},
	},
}
