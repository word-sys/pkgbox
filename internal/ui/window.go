package ui

import (
	"context"
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/detector"
	"pkgbox/internal/downloader"
	"pkgbox/internal/installer"
)

type AppWindow struct {
	Window       *gtk.Window
	RootBox      *gtk.Box
	Stack        *gtk.Stack
	DropZone     *DropZone
	InfoCard     *InfoCard
	ProgressView *ProgressView
	DownloadView *DownloadView
	InfoLabel    *gtk.Label
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

	infoLabel, err := gtk.LabelNew("Ready. Drop a package file or URL above.")
	if err != nil {
		return nil, err
	}

	appWin := &AppWindow{
		Window:    win,
		RootBox:   box,
		Stack:     stack,
		InfoLabel: infoLabel,
	}

	dz, err := NewDropZone(func(item DropItem) {
		appWin.onItemDropped(item)
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

	progressView, err := NewProgressView(func() {
		appWin.onDoneRequested()
	})
	if err != nil {
		return nil, err
	}
	appWin.ProgressView = progressView

	downloadView, err := NewDownloadView(func() {
		appWin.onDownloadCancelled()
	})
	if err != nil {
		return nil, err
	}
	appWin.DownloadView = downloadView

	stack.AddNamed(dz.Widget, "drop")
	stack.AddNamed(infoCard.Widget, "info")
	stack.AddNamed(progressView.Widget, "progress")
	stack.AddNamed(downloadView.Widget, "download")

	box.PackStart(stack, true, true, 8)
	box.PackStart(infoLabel, false, false, 4)

	win.Add(box)
	return appWin, nil
}

func (w *AppWindow) onItemDropped(item DropItem) {
	if item.Kind == DropKindFile {
		w.processLocalFile(item.Value)
	} else if item.Kind == DropKindURL {
		w.startRemoteDownload(item.Value)
	}
}

func (w *AppWindow) processLocalFile(filePath string) {
	info, err := detector.InspectFile(filePath)
	if err != nil {
		w.InfoLabel.SetText(fmt.Sprintf("Error inspecting file: %v", err))
		return
	}

	w.InfoCard.Update(info)
	w.Stack.SetVisibleChildName("info")
	w.InfoLabel.SetText(fmt.Sprintf("Inspected: %s (%s)", info.AppName, info.Type))
}

func (w *AppWindow) startRemoteDownload(rawURL string) {
	ctx, cancel := context.WithCancel(context.Background())
	w.DownloadView.Reset(rawURL, cancel)
	w.Stack.SetVisibleChildName("download")
	w.InfoLabel.SetText("Downloading package...")

	go func() {
		localPath, err := downloader.DownloadPackage(ctx, rawURL, func(downloaded, total int64, fraction float64) {
			w.DownloadView.UpdateProgress(downloaded, total, fraction)
		})
		if err != nil {
			glib.IdleAdd(func() bool {
				if ctx.Err() == context.Canceled {
					w.InfoLabel.SetText("Download cancelled.")
				} else {
					w.InfoLabel.SetText(fmt.Sprintf("Download failed: %v", err))
				}
				w.Stack.SetVisibleChildName("drop")
				return false
			})
			return
		}

		glib.IdleAdd(func() bool {
			w.processLocalFile(localPath)
			return false
		})
	}()
}

func (w *AppWindow) onDownloadCancelled() {
	w.Stack.SetVisibleChildName("drop")
	w.InfoLabel.SetText("Download cancelled. Ready for new package.")
}

func (w *AppWindow) onCancelRequested() {
	w.Stack.SetVisibleChildName("drop")
	w.InfoLabel.SetText("Ready. Drop a package file or URL above.")
}

func (w *AppWindow) onDoneRequested() {
	w.Stack.SetVisibleChildName("drop")
	w.InfoLabel.SetText("Ready. Drop a package file or URL above.")
}

func (w *AppWindow) onInstallRequested(info *detector.FileInfo) {
	w.ProgressView.Reset(info.AppName)
	w.Stack.SetVisibleChildName("progress")
	w.InfoLabel.SetText(fmt.Sprintf("Installing %s...", info.AppName))

	go func() {
		result, err := installer.InstallUserSpaceApp(info, func(stage string, fraction float64) {
			w.ProgressView.UpdateProgress(stage, fraction)
		})
		if err != nil {
			w.ProgressView.ShowError(err)
			return
		}
		w.ProgressView.ShowSuccess(result)
	}()
}

func (w *AppWindow) onPermissionsRequested(info *detector.FileInfo) {
	w.InfoLabel.SetText(fmt.Sprintf("Configuring permissions for %s", info.AppName))
}

func (w *AppWindow) Show() {
	w.Window.ShowAll()
}
