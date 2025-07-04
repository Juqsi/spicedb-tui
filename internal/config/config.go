package config

import (
	"encoding/json"
	"fmt"
	"github.com/rivo/tview"
	"os"
	"spicedb-tui/internal/i18n"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
	Language string `json:"language"`
}

var (
	Current    Config
	configPath = "config.json"
)

func LoadOrAskForConfig(app *tview.Application, start func(*tview.Application)) error {
	Current = Config{
		Endpoint: "localhost:50051",
		Token:    "foobar",
		Language: "en",
	}

	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(data, &Current)
			if Current.Language != "" {
				i18n.SetLanguage(Current.Language)
			}
		}
		start(app)
		return nil
	}

	defaultConfigBytes, _ := json.MarshalIndent(Current, "", "  ")
	if err := os.WriteFile(configPath, defaultConfigBytes, 0644); err != nil {
		return fmt.Errorf("could not write config.json: %v", err)
	}
	ShowConfigPage(app, nil, start)
	return nil
}

func ShowConfigPage(app *tview.Application, appPages *tview.Pages, onSave func(*tview.Application)) {
	langs := []string{"en", "de"}
	var defaultLangIndex int
	for i, l := range langs {
		if l == Current.Language {
			defaultLangIndex = i
		}
	}

	form := tview.NewForm().
		AddInputField(i18n.T("endpoint"), Current.Endpoint, 40, nil, func(text string) {
			Current.Endpoint = text
		}).
		AddInputField(i18n.T("token"), Current.Token, 40, nil, func(text string) {
			Current.Token = text
		}).
		AddDropDown(i18n.T("language"), langs, defaultLangIndex, func(option string, idx int) {
			Current.Language = option
			i18n.SetLanguage(option)
		}).
		AddButton(i18n.T("continue"), func() {
			data, _ := json.MarshalIndent(Current, "", "  ")
			os.WriteFile(configPath, data, 0644)
			onSave(app)
		}).
		AddButton(i18n.T("exit"), func() {
			app.Stop()
			os.Exit(0)
		})

	form.SetBorder(true).SetTitle(i18n.T("config_title")).SetTitleAlign(tview.AlignLeft)
	if appPages != nil {
		appPages.AddAndSwitchToPage("config", form, true)
	} else {
		app.SetRoot(form, true)
	}
}
