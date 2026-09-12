package ui

import (
	"context"
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type DownloadView struct {
	Widget      *gtk.Box
	TitleLabel  *gtk.Label
	URLLabel    *gtk.Label
	ProgressBar *gtk.ProgressBar
	StatusLabel *gtk.Label
	CancelBtn   *gtk.Button
	CancelFunc  context.CancelFunc
	OnCancel    func()
}

func NewDownloadView(onCancel func()) (*DownloadView, error) {
	root, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 16)
	if err != nil {
		return nil, err
	}
	root.SetMarginStart(24)
	root.SetMarginEnd(24)
	root.SetMarginTop(36)
	root.SetMarginBottom(24)

	titleLabel, err := gtk.LabelNew("Downloading Remote Package...")
	if err != nil {
		return nil, err
	}

	urlLabel, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	urlLabel.SetEllipsize(3)
	urlLabel.SetSelectable(true)

	progressBar, err := gtk.ProgressBarNew()
	if err != nil {
		return nil, err
	}
	progressBar.SetFraction(0.0)

	statusLabel, err := gtk.LabelNew("Connecting to server...")
	if err != nil {
		return nil, err
	}

	cancelBtn, err := gtk.ButtonNewWithLabel("Cancel Download")
	if err != nil {
		return nil, err
	}
	cancelBtn.SetHAlign(gtk.ALIGN_CENTER)

	root.PackStart(titleLabel, false, false, 0)
	root.PackStart(urlLabel, false, false, 0)
	root.PackStart(progressBar, false, false, 8)
	root.PackStart(statusLabel, false, false, 0)
	root.PackStart(cancelBtn, false, false, 12)

	view := &DownloadView{
		Widget:      root,
		TitleLabel:  titleLabel,
		URLLabel:    urlLabel,
		ProgressBar: progressBar,
		StatusLabel: statusLabel,
		CancelBtn:   cancelBtn,
		OnCancel:    onCancel,
	}

	cancelBtn.Connect("clicked", func() {
		if view.CancelFunc != nil {
			view.CancelFunc()
		}
		if view.OnCancel != nil {
			view.OnCancel()
		}
	})

	return view, nil
}

func (dv *DownloadView) Reset(targetURL string, cancelFunc context.CancelFunc) {
	dv.URLLabel.SetText(targetURL)
	dv.StatusLabel.SetText("Connecting to server...")
	dv.ProgressBar.SetFraction(0.0)
	dv.CancelFunc = cancelFunc
}

func (dv *DownloadView) UpdateProgress(downloaded, total int64, fraction float64) {
	glib.IdleAdd(func() bool {
		dv.ProgressBar.SetFraction(fraction)
		if total > 0 {
			pct := int(fraction * 100)
			dv.StatusLabel.SetText(fmt.Sprintf("Downloading: %s / %s (%d%%)", formatBytes(downloaded), formatBytes(total), pct))
		} else {
			dv.StatusLabel.SetText(fmt.Sprintf("Downloaded: %s", formatBytes(downloaded)))
		}
		return false
	})
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
