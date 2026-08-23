package main

import (
	"log"
	"os"

	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/ui"
)

func main() {
	gtk.Init(&os.Args)

	appWin, err := ui.NewAppWindow()
	if err != nil {
		log.Fatalf("failed to initialize application window: %v", err)
	}

	appWin.Show()
	gtk.Main()
}
