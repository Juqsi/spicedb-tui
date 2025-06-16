package main

import (
	"log"
	"os"
	"os/signal"
	"spicedb-tui/internal/tui"
	"syscall"

	"github.com/rivo/tview"
	"spicedb-tui/internal/config"
	"spicedb-tui/internal/i18n"
)

var app *tview.Application

func main() {
	app = tview.NewApplication()

	if err := config.LoadOrAskForConfig(app, tui.StartTUI); err != nil {
		log.Fatalf(i18n.T("error_loading_config"), err)
	}

	if config.Current.Language != "" {
		i18n.SetLanguage(config.Current.Language)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		app.Stop()
		os.Exit(0)
	}()

	if err := app.Run(); err != nil {
		log.Fatalf(i18n.T("error_running_tui"), err)
	}
}
