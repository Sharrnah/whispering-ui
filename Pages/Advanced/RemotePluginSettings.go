package Advanced

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"reflect"
	"sort"
	"whispering-tiger-ui/LocalPlugins"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

// Each renderer keeps its own settings and destination. Local integrations must
// never mutate the AI profile or send their controls to its WebSocket.
type pluginScope struct {
	config *Settings.Conf
	send   func(SendMessageChannel.SendMessageStruct)
}

func localPluginScope(config *Settings.Conf, send func(interface{})) *pluginScope {
	return &pluginScope{config: config, send: func(message SendMessageChannel.SendMessageStruct) {
		if message.Type == "setting_change" {
			if message.Name == "plugins" {
				config.Plugins = map[string]bool{}
				for name, active := range message.Value.(map[string]bool) {
					if active {
						config.Plugins[name] = true
					}
				}
			}
			if message.Name == "plugin_settings" {
				config.Plugin_settings = message.Value
			}
			enabled := []string{}
			for name, active := range config.Plugins {
				if active {
					enabled = append(enabled, name)
				}
			}
			sort.Strings(enabled)
			fyne.CurrentApp().Preferences().SetBool("local.plugins.enabled", len(enabled) > 0)
			send(map[string]interface{}{"type": "configure", "enabled": enabled, "settings": config.Plugin_settings})
		} else {
			send(message)
		}
	}}
}

var remotePluginTabIndex int

func CreatePluginSettingsPage() fyne.CanvasObject {
	if Settings.Config.Run_backend {
		return createBackendPluginSettingsPage()
	}
	local := createLocalPluginSettingsPage(LocalPlugins.Observe)
	tabs := container.NewAppTabs(container.NewTabItem(lang.L("This PC"), local), container.NewTabItem(lang.L("AI PC"), createBackendPluginSettingsPage()))
	tabs.SelectIndex(remotePluginTabIndex)
	tabs.OnSelected = func(item *container.TabItem) {
		remotePluginTabIndex = tabs.SelectedIndex()
		if item == tabs.Items[0] {
			LocalPlugins.Start()
		}
	}
	return tabs
}

func createLocalPluginSettingsPage(observe func(func(LocalPlugins.Snapshot), func(string))) fyne.CanvasObject {
	status := widget.NewLabel(lang.L("Starting local plugins"))
	status.Wrapping = fyne.TextWrapWord
	description := widget.NewLabel(lang.L("Plugins here run on this PC and receive transcripts and translations from the AI PC. Use AI PC for speech recognition, translation and voice generation plugins."))
	description.Wrapping = fyne.TextWrapWord
	content := container.NewVScroll(widget.NewLabel(lang.L("Loading plugins")))
	config := Settings.Conf{Plugins: map[string]bool{}, Plugin_settings: map[string]interface{}{}}
	scope := localPluginScope(&config, LocalPlugins.Send)
	rendered := false
	var available []string
	render := func(value LocalPlugins.Snapshot) {
		nextPlugins := map[string]bool{}
		for _, name := range value.Enabled {
			nextPlugins[name] = true
		}
		nextSettings := map[string]interface{}{}
		for name, values := range value.Settings {
			nextSettings[name] = values
		}
		if rendered && reflect.DeepEqual(available, value.Available) && reflect.DeepEqual(config.Plugins, nextPlugins) && reflect.DeepEqual(config.Plugin_settings, nextSettings) {
			return
		}
		rendered = true
		available = append([]string(nil), value.Available...)
		open := map[string]bool{}
		if previous, ok := content.Content.(*widget.Accordion); ok {
			for _, item := range previous.Items {
				if item.Open {
					open[item.Title] = true
				}
			}
		}
		config.Plugins = map[string]bool{}
		for _, name := range value.Enabled {
			config.Plugins[name] = true
		}
		settings := map[string]interface{}{}
		for name, values := range value.Settings {
			settings[name] = values
		}
		config.Plugin_settings = settings
		accordion := widget.NewAccordion()
		for _, name := range value.Available {
			item := widget.NewAccordionItem(name, nil)
			item.Detail = BuildSinglePluginSettings(name, item, accordion, nil, nil, scope)
			item.Open = open[name] || open[name+" (\u2705)"] || open[name+" (\u274c)"]
			accordion.Append(item)
		}
		content.Content = accordion
		if len(value.Available) == 0 {
			content.Content = widget.NewLabel(lang.L("No Plugins found. Download Plugins using the button below."))
		}
		content.Refresh()
	}
	observe(render, status.SetText)
	download := widget.NewButton(lang.L("Download / Update Plugins"), func() {
		CreatePluginListWindow(func() { LocalPlugins.Send(map[string]interface{}{"type": "refresh"}) }, false, func(_, _ string) { LocalPlugins.Restart() })
	})
	reload := widget.NewButton(lang.L("Reload"), func() { LocalPlugins.Start(); LocalPlugins.Send(map[string]interface{}{"type": "refresh"}) })
	return container.NewBorder(container.NewVBox(description, container.NewBorder(nil, nil, download, reload, status)), nil, nil, nil, content)
}
