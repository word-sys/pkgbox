package ui

import (
	"os/exec"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/installer"
)

type ProgressView struct {
	Widget      *gtk.Box
	TitleLabel  *gtk.Label
	StatusLabel *gtk.Label
	ProgressBar *gtk.ProgressBar
	ActionBox   *gtk.Box
	LaunchBtn   *gtk.Button
	DoneBtn     *gtk.Button
	LastResult  *installer.InstallResult
	OnDone      func()
}

func NewProgressView(onDone func()) (*ProgressView, error) {
	root, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 16)
	if err != nil {
		return nil, err
	}
	root.SetMarginStart(24)
	root.SetMarginEnd(24)
	root.SetMarginTop(32)
	root.SetMarginBottom(24)

	titleLabel, err := gtk.LabelNew("Installing Application...")
	if err != nil {
		return nil, err
	}

	progressBar, err := gtk.ProgressBarNew()
	if err != nil {
		return nil, err
	}
	progressBar.SetFraction(0.0)

	statusLabel, err := gtk.LabelNew("Starting installation...")
	if err != nil {
		return nil, err
	}

	actionBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	if err != nil {
		return nil, err
	}
	actionBox.SetHAlign(gtk.ALIGN_CENTER)

	launchBtn, err := gtk.ButtonNewWithLabel("Launch Application")
	if err != nil {
		return nil, err
	}

	doneBtn, err := gtk.ButtonNewWithLabel("Done")
	if err != nil {
		return nil, err
	}

	actionBox.PackStart(launchBtn, false, false, 0)
	actionBox.PackStart(doneBtn, false, false, 0)

	root.PackStart(titleLabel, false, false, 0)
	root.PackStart(progressBar, false, false, 8)
	root.PackStart(statusLabel, false, false, 0)
	root.PackStart(actionBox, false, false, 16)

	view := &ProgressView{
		Widget:      root,
		TitleLabel:  titleLabel,
		StatusLabel: statusLabel,
		ProgressBar: progressBar,
		ActionBox:   actionBox,
		LaunchBtn:   launchBtn,
		DoneBtn:     doneBtn,
		OnDone:      onDone,
	}

	actionBox.SetVisible(false)

	doneBtn.Connect("clicked", func() {
		if view.OnDone != nil {
			view.OnDone()
		}
	})

	launchBtn.Connect("clicked", func() {
		if view.LastResult != nil && view.LastResult.InstallPath != "" {
			cmd := exec.Command(view.LastResult.InstallPath)
			_ = cmd.Start()
		}
	})

	return view, nil
}

func (pv *ProgressView) Reset(appName string) {
	pv.TitleLabel.SetText("Installing " + appName + "...")
	pv.StatusLabel.SetText("Initializing...")
	pv.ProgressBar.SetFraction(0.0)
	pv.ActionBox.SetVisible(false)
	pv.LastResult = nil
}

func (pv *ProgressView) UpdateProgress(stage string, fraction float64) {
	glib.IdleAdd(func() bool {
		pv.StatusLabel.SetText(stage)
		pv.ProgressBar.SetFraction(fraction)
		return false
	})
}

func (pv *ProgressView) ShowSuccess(result *installer.InstallResult) {
	glib.IdleAdd(func() bool {
		pv.LastResult = result
		pv.TitleLabel.SetText(result.AppName + " Installed Successfully")
		pv.StatusLabel.SetText("Desktop shortcut created in your application menu.")
		pv.ProgressBar.SetFraction(1.0)
		pv.ActionBox.SetVisible(true)
		return false
	})
}

func (pv *ProgressView) ShowError(err error) {
	glib.IdleAdd(func() bool {
		pv.TitleLabel.SetText("Installation Failed")
		pv.StatusLabel.SetText(err.Error())
		pv.ProgressBar.SetFraction(1.0)
		pv.ActionBox.SetVisible(true)
		pv.LaunchBtn.SetVisible(false)
		return false
	})
}
