package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/detector"
)

type AppWindow struct {
	Window    *gtk.Window
	RootBox   *gtk.Box
	DropZone  *DropZone
	InfoLabel *gtk.Label
}

func NewAppWindow() (*AppWindow, error) {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, err
	}

	win.SetTitle("PkgBox")
	win.SetDefaultSize(520, 420)
	win.SetPosition(gtk.WIN_POS_CENTER)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	if err != nil {
		return nil, err
	}
	box.SetMarginStart(20)
	box.SetMarginEnd(20)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)

	titleLabel, err := gtk.LabelNew("PkgBox - Universal Linux App Installer")
	if err != nil {
		return nil, err
	}
	box.PackStart(titleLabel, false, false, 0)

	appWin := &AppWindow{
		Window:  win,
		RootBox: box,
	}

	infoLabel, err := gtk.LabelNew("Ready. Drop a package file above.")
	if err != nil {
		return nil, err
	}
	appWin.InfoLabel = infoLabel

	dz, err := NewDropZone(func(filePath string) {
		appWin.onFileDropped(filePath)
	})
	if err != nil {
		return nil, err
	}
	appWin.DropZone = dz

	box.PackStart(dz.Widget, true, true, 8)
	box.PackStart(infoLabel, false, false, 4)

	win.Add(box)
	return appWin, nil
}

func (w *AppWindow) onFileDropped(filePath string) {
	info, err := detector.InspectFile(filePath)
	if err != nil {
		w.InfoLabel.SetText(fmt.Sprintf("Error inspecting file: %v", err))
		return
	}

	w.InfoLabel.SetText(fmt.Sprintf("Detected: %s | Type: %s | Arch: %s | Size: %s",
		info.FileName, info.Type, info.Arch, info.FormattedSize))
}

func (w *AppWindow) Show() {
	w.Window.ShowAll()
}
