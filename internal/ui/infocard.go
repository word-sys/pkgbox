package ui

import (
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"pkgbox/internal/detector"
)

type InfoCard struct {
	Widget         *gtk.Box
	NameLabel      *gtk.Label
	TypeLabel      *gtk.Label
	ArchLabel      *gtk.Label
	SizeLabel      *gtk.Label
	ShaLabel       *gtk.Label
	CopyShaBtn     *gtk.Button
	PathLabel      *gtk.Label
	ModeLabel      *gtk.Label
	InstallBtn     *gtk.Button
	PermsBtn       *gtk.Button
	CancelBtn      *gtk.Button
	CurrentInfo    *detector.FileInfo
	OnInstall      func(info *detector.FileInfo)
	OnPermissions  func(info *detector.FileInfo)
	OnCancel       func()
}

func NewInfoCard(onInstall func(info *detector.FileInfo), onPerms func(info *detector.FileInfo), onCancel func()) (*InfoCard, error) {
	root, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	if err != nil {
		return nil, err
	}

	frame, err := gtk.FrameNew("Package Details")
	if err != nil {
		return nil, err
	}

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid.SetRowSpacing(8)
	grid.SetColumnSpacing(16)
	grid.SetMarginStart(16)
	grid.SetMarginEnd(16)
	grid.SetMarginTop(16)
	grid.SetMarginBottom(16)

	nameTitle, _ := gtk.LabelNew("Application:")
	nameTitle.SetHAlign(gtk.ALIGN_START)
	nameLabel, _ := gtk.LabelNew("-")
	nameLabel.SetHAlign(gtk.ALIGN_START)
	nameLabel.SetSelectable(true)

	typeTitle, _ := gtk.LabelNew("Package Type:")
	typeTitle.SetHAlign(gtk.ALIGN_START)
	typeLabel, _ := gtk.LabelNew("-")
	typeLabel.SetHAlign(gtk.ALIGN_START)

	archTitle, _ := gtk.LabelNew("Architecture:")
	archTitle.SetHAlign(gtk.ALIGN_START)
	archLabel, _ := gtk.LabelNew("-")
	archLabel.SetHAlign(gtk.ALIGN_START)

	sizeTitle, _ := gtk.LabelNew("File Size:")
	sizeTitle.SetHAlign(gtk.ALIGN_START)
	sizeLabel, _ := gtk.LabelNew("-")
	sizeLabel.SetHAlign(gtk.ALIGN_START)

	shaTitle, _ := gtk.LabelNew("SHA256:")
	shaTitle.SetHAlign(gtk.ALIGN_START)

	shaBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	shaLabel, _ := gtk.LabelNew("-")
	shaLabel.SetHAlign(gtk.ALIGN_START)
	shaLabel.SetEllipsize(3) // PANGO_ELLIPSIZE_END
	shaLabel.SetSelectable(true)

	copyShaBtn, _ := gtk.ButtonNewWithLabel("Copy")
	copyShaBtn.SetHAlign(gtk.ALIGN_START)
	shaBox.PackStart(shaLabel, true, true, 0)
	shaBox.PackStart(copyShaBtn, false, false, 0)

	modeTitle, _ := gtk.LabelNew("Install Scope:")
	modeTitle.SetHAlign(gtk.ALIGN_START)
	modeLabel, _ := gtk.LabelNew("-")
	modeLabel.SetHAlign(gtk.ALIGN_START)

	pathTitle, _ := gtk.LabelNew("Source File:")
	pathTitle.SetHAlign(gtk.ALIGN_START)
	pathLabel, _ := gtk.LabelNew("-")
	pathLabel.SetHAlign(gtk.ALIGN_START)
	pathLabel.SetEllipsize(3)
	pathLabel.SetSelectable(true)

	grid.Attach(nameTitle, 0, 0, 1, 1)
	grid.Attach(nameLabel, 1, 0, 1, 1)
	grid.Attach(typeTitle, 0, 1, 1, 1)
	grid.Attach(typeLabel, 1, 1, 1, 1)
	grid.Attach(archTitle, 0, 2, 1, 1)
	grid.Attach(archLabel, 1, 2, 1, 1)
	grid.Attach(sizeTitle, 0, 3, 1, 1)
	grid.Attach(sizeLabel, 1, 3, 1, 1)
	grid.Attach(shaTitle, 0, 4, 1, 1)
	grid.Attach(shaBox, 1, 4, 1, 1)
	grid.Attach(modeTitle, 0, 5, 1, 1)
	grid.Attach(modeLabel, 1, 5, 1, 1)
	grid.Attach(pathTitle, 0, 6, 1, 1)
	grid.Attach(pathLabel, 1, 6, 1, 1)

	frame.Add(grid)
	root.PackStart(frame, true, true, 0)

	btnBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
	if err != nil {
		return nil, err
	}
	btnBox.SetHAlign(gtk.ALIGN_END)

	cancelBtn, err := gtk.ButtonNewWithLabel("Cancel")
	if err != nil {
		return nil, err
	}

	permsBtn, err := gtk.ButtonNewWithLabel("Permissions")
	if err != nil {
		return nil, err
	}

	installBtn, err := gtk.ButtonNewWithLabel("Install")
	if err != nil {
		return nil, err
	}

	btnBox.PackStart(cancelBtn, false, false, 0)
	btnBox.PackStart(permsBtn, false, false, 0)
	btnBox.PackStart(installBtn, false, false, 0)

	root.PackStart(btnBox, false, false, 0)

	card := &InfoCard{
		Widget:        root,
		NameLabel:     nameLabel,
		TypeLabel:     typeLabel,
		ArchLabel:     archLabel,
		SizeLabel:     sizeLabel,
		ShaLabel:      shaLabel,
		CopyShaBtn:    copyShaBtn,
		PathLabel:     pathLabel,
		ModeLabel:     modeLabel,
		InstallBtn:    installBtn,
		PermsBtn:      permsBtn,
		CancelBtn:     cancelBtn,
		OnInstall:     onInstall,
		OnPermissions: onPerms,
		OnCancel:      onCancel,
	}

	copyShaBtn.Connect("clicked", func() {
		if card.CurrentInfo != nil && card.CurrentInfo.SHA256 != "" {
			CopyToClipboard(card.CurrentInfo.SHA256)
			card.CopyShaBtn.SetLabel("Copied!")
			time.AfterFunc(1500*time.Millisecond, func() {
				glib.IdleAdd(func() bool {
					card.CopyShaBtn.SetLabel("Copy")
					return false
				})
			})
		}
	})

	cancelBtn.Connect("clicked", func() {
		if card.OnCancel != nil {
			card.OnCancel()
		}
	})

	permsBtn.Connect("clicked", func() {
		if card.OnPermissions != nil && card.CurrentInfo != nil {
			card.OnPermissions(card.CurrentInfo)
		}
	})

	installBtn.Connect("clicked", func() {
		if card.OnInstall != nil && card.CurrentInfo != nil {
			card.OnInstall(card.CurrentInfo)
		}
	})

	return card, nil
}

func CopyToClipboard(text string) {
	clip, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err == nil && clip != nil {
		clip.SetText(text)
		clip.Store()
	}
}

func (c *InfoCard) Update(info *detector.FileInfo) {
	c.CurrentInfo = info
	if info == nil {
		return
	}

	c.NameLabel.SetText(info.AppName)
	c.TypeLabel.SetText(string(info.Type))
	c.ArchLabel.SetText(info.Arch)
	c.SizeLabel.SetText(info.FormattedSize)
	c.ShaLabel.SetText(info.SHA256)
	c.PathLabel.SetText(info.Path)

	scope := "User-space (No root required)"
	if info.Type == detector.TypeDeb || info.Type == detector.TypeRPM {
		scope = "System-wide (Polkit authentication required)"
	}
	c.ModeLabel.SetText(scope)
}
