package ui

import (
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type DropZone struct {
	Widget         *gtk.EventBox
	StatusLabel    *gtk.Label
	SubLabel       *gtk.Label
	OnFileSelected func(filePath string)
}

func NewDropZone(onFileSelected func(filePath string)) (*DropZone, error) {
	eventBox, err := gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}

	frame, err := gtk.FrameNew("")
	if err != nil {
		return nil, err
	}
	frame.SetShadowType(gtk.SHADOW_IN)

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 8)
	if err != nil {
		return nil, err
	}
	box.SetMarginStart(30)
	box.SetMarginEnd(30)
	box.SetMarginTop(40)
	box.SetMarginBottom(40)

	statusLabel, err := gtk.LabelNew("Drag and drop package here")
	if err != nil {
		return nil, err
	}

	subLabel, err := gtk.LabelNew("Supports .AppImage, .deb, .rpm, .flatpak, and binaries")
	if err != nil {
		return nil, err
	}

	browseBtn, err := gtk.ButtonNewWithLabel("Browse File...")
	if err != nil {
		return nil, err
	}
	browseBtn.SetHAlign(gtk.ALIGN_CENTER)

	box.PackStart(statusLabel, false, false, 0)
	box.PackStart(subLabel, false, false, 0)
	box.PackStart(browseBtn, false, false, 8)

	frame.Add(box)
	eventBox.Add(frame)

	dz := &DropZone{
		Widget:         eventBox,
		StatusLabel:    statusLabel,
		SubLabel:       subLabel,
		OnFileSelected: onFileSelected,
	}

	targetEntry, err := gtk.TargetEntryNew("text/uri-list", gtk.TARGET_OTHER_APP, 0)
	if err != nil {
		return nil, err
	}

	eventBox.DragDestSet(
		gtk.DEST_DEFAULT_ALL,
		[]gtk.TargetEntry{*targetEntry},
		gdk.ACTION_COPY,
	)

	eventBox.Connect("drag-motion", func(w *gtk.EventBox, ctx *gdk.DragContext, x, y int, time uint) bool {
		dz.StatusLabel.SetText("Release mouse to drop package")
		return true
	})

	eventBox.Connect("drag-leave", func(w *gtk.EventBox, ctx *gdk.DragContext, time uint) {
		dz.StatusLabel.SetText("Drag and drop package here")
	})

	eventBox.Connect("drag-data-received", func(w *gtk.EventBox, ctx *gdk.DragContext, x, y int, data *gtk.SelectionData, info, time uint) {
		dz.StatusLabel.SetText("Drag and drop package here")
		raw := string(data.GetData())
		paths, err := ParseURIList(raw)
		if err != nil || len(paths) == 0 {
			log.Printf("invalid drop: %v", err)
			return
		}

		if dz.OnFileSelected != nil {
			dz.OnFileSelected(paths[0])
		}
	})

	browseBtn.Connect("clicked", func() {
		toplevel, _ := eventBox.GetToplevel()
		win, ok := toplevel.(*gtk.Window)
		if !ok {
			return
		}

		dlg, err := gtk.FileChooserDialogNewWith2Buttons(
			"Select Package to Install",
			win,
			gtk.FILE_CHOOSER_ACTION_OPEN,
			"Cancel",
			gtk.RESPONSE_CANCEL,
			"Open",
			gtk.RESPONSE_ACCEPT,
		)
		if err != nil {
			return
		}
		defer dlg.Destroy()

		if dlg.Run() == gtk.RESPONSE_ACCEPT {
			filename := dlg.GetFilename()
			if filename != "" && dz.OnFileSelected != nil {
				dz.OnFileSelected(filename)
			}
		}
	})

	return dz, nil
}
