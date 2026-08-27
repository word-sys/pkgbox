package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/detector"
)

type AppWindow struct {
	Window    *gtk.Window
	RootBox   *gtk.Box
	Stack     *gtk.Stack
	DropZone  *DropZone
	InfoCard  *InfoCard
	InfoLabel *gtk.Label
}

func NewAppWindow() (*AppWindow, error) {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, err
	}

	win.SetTitle("PkgBox")
	win.SetDefaultSize(560, 440)
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

	stack, err := gtk.StackNew()
	if err != nil {
		return nil, err
	}
	stack.SetTransitionType(gtk.STACK_TRANSITION_TYPE_CROSSFADE)
	stack.SetTransitionDuration(200)

	infoLabel, err := gtk.LabelNew("Ready. Drop a package file above.")
	if err != nil {
		return nil, err
	}

	appWin := &AppWindow{
		Window:    win,
		RootBox:   box,
		Stack:     stack,
		InfoLabel: infoLabel,
	}

	dz, err := NewDropZone(func(filePath string) {
		appWin.onFileSelected(filePath)
	})
	if err != nil {
		return nil, err
	}
	appWin.DropZone = dz

	infoCard, err := NewInfoCard(
		func(info *detector.FileInfo) {
			appWin.onInstallRequested(info)
		},
		func(info *detector.FileInfo) {
			appWin.onPermissionsRequested(info)
		},
		func() {
			appWin.onCancelRequested()
		},
	)
	if err != nil {
		return nil, err
	}
	appWin.InfoCard = infoCard

	stack.AddNamed(dz.Widget, "drop")
	stack.AddNamed(infoCard.Widget, "info")

	box.PackStart(stack, true, true, 8)
	box.PackStart(infoLabel, false, false, 4)

	win.Add(box)
	return appWin, nil
}

func (w *AppWindow) onFileSelected(filePath string) {
	info, err := detector.InspectFile(filePath)
	if err != nil {
		w.InfoLabel.SetText(fmt.Sprintf("Error inspecting file: %v", err))
		return
	}

	w.InfoCard.Update(info)
	w.Stack.SetVisibleChildName("info")
	w.InfoLabel.SetText(fmt.Sprintf("Inspected: %s (%s)", info.AppName, info.Type))
}

func (w *AppWindow) onCancelRequested() {
	w.Stack.SetVisibleChildName("drop")
	w.InfoLabel.SetText("Ready. Drop a package file above.")
}

func (w *AppWindow) onInstallRequested(info *detector.FileInfo) {
	w.InfoLabel.SetText(fmt.Sprintf("Ready to install %s (%s)", info.AppName, info.Type))
}

func (w *AppWindow) onPermissionsRequested(info *detector.FileInfo) {
	w.InfoLabel.SetText(fmt.Sprintf("Configuring permissions for %s", info.AppName))
}

func (w *AppWindow) Show() {
	w.Window.ShowAll()
}
