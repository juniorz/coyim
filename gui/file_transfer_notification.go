package gui

import (
	"fmt"
	"log"

	"github.com/coyim/gotk3adapter/gtki"
)

type fileNotification struct {
	area     gtki.Box
	label    gtki.Label
	image    gtki.Image
	success  bool
	canceled bool
	progress float64
}

type fileTransferNotification struct {
	area        gtki.Box
	image       gtki.Image
	label       gtki.Label
	box         gtki.Box
	progressBar gtki.ProgressBar
	button      gtki.Button
	labelButton gtki.Label
	canceled    bool
	files       []*fileNotification
	progress    float64
}

func resizeFileName(name string) string {
	var fileName string

	if len(name) > 20 {
		fileName = name[:21] + "..."
		return fileName
	}

	return name

}

func (file *fileNotification) destroy() {
	file.canceled = true
}

func (file *fileNotification) update(fileName string) {
	file.label.SetLabel(fileName)
	file.image.Hide()
}

// TODO: on close, destroy canceled and failed as well
func (b *builder) fileTransferNotifInit() *fileTransferNotification {
	fileTransferNotif := &fileTransferNotification{}

	b.getItems(
		"file-transfer", &fileTransferNotif.area,
		"image-file-transfer", &fileTransferNotif.image,
		"label-file-transfer", &fileTransferNotif.label,
		"info-file-transfer", &fileTransferNotif.box,
		"bar-file-transfer", &fileTransferNotif.progressBar,
		"button-file-transfer", &fileTransferNotif.button,
		"button-label-file-transfer", &fileTransferNotif.labelButton,
	)

	return fileTransferNotif
}

func (conv *conversationPane) showFileTransferInfo(fileName string) *fileNotification {
	b := newBuilder("FileTransferNotification")

	file := &fileNotification{}

	b.getItems(
		"area-file-transfer-info", &file.area,
		"name-file-transfer-info", &file.label,
		"image-file-transfer-info", &file.image,
	)

	b.ConnectSignals(map[string]interface{}{
		"on_destroy_single_file_transfer": file.destroy,
	})

	// TODO: set progress to zero everytime?
	label := "File transfer started"
	conv.updateFileTransferNotification(label, "Cancel", "filetransfer.svg")
	conv.fileTransferNotif.canceled = false

	file.label.SetLabel(fileName)

	conv.fileTransferNotif.box.Add(file.area)
	file.area.ShowAll()

	conv.fileTransferNotif.files = append(conv.fileTransferNotif.files, file)

	return file
}

func (conv *conversationPane) showFileTransferNotification(fileName string) *fileNotification {
	prov, _ := g.gtk.CssProviderNew()

	css := fmt.Sprintf(`
	box { background-color: #fff9f3;
	      color: #000000;
	      border: 3px;
	     }
	`)
	_ = prov.LoadFromData(css)

	styleContext, _ := conv.fileTransferNotif.area.GetStyleContext()
	styleContext.AddProvider(prov, 9999)

	label := "File transfer started"
	conv.updateFileTransferNotification(label, "Cancel", "filetransfer.svg")
	conv.fileTransferNotif.progressBar.SetFraction(0.0)
	conv.fileTransferNotif.canceled = false

	info := conv.showFileTransferInfo(fileName)

	conv.fileTransferNotif.area.SetVisible(true)

	return info
}

func (conv *conversationPane) getFileTransferNotification() bool {
	if !conv.fileTransferNotif.area.IsVisible() {
		return false
	}
	return true
}

func (conv *conversationPane) updateFileTransferNotification(label, buttonLabel, image string) {
	if buttonLabel == "Close" {
		prov, _ := g.gtk.CssProviderNew()

		css := fmt.Sprintf(`
	                  label { margin-right: 3px;
	                          margin-left: 3px;
	                  }
	                `)
		_ = prov.LoadFromData(css)

		styleContext, _ := conv.fileTransferNotif.labelButton.GetStyleContext()
		styleContext.AddProvider(prov, 9999)
	}

	log.Printf(label)
	conv.fileTransferNotif.label.SetLabel(label)
	conv.fileTransferNotif.labelButton.SetLabel(buttonLabel)
	setImageFromFile(conv.fileTransferNotif.image, image)
}

func (conv *conversationPane) startFileTransfer(file *fileNotification) {
	conv.fileTransferNotif.progress = 0
	for i := range conv.fileTransferNotif.files {
		conv.fileTransferNotif.progress += conv.fileTransferNotif.files[i].progress
	}

	upd := conv.fileTransferNotif.progress / float64(len(conv.fileTransferNotif.files))

	doInUIThread(func() {
		conv.fileTransferNotif.progressBar.SetFraction(upd)
	})
}

func (conv *conversationPane) successFileTransfer(fileName string, file *fileNotification) {
	prov, _ := g.gtk.CssProviderNew()

	css := fmt.Sprintf(`
	label { color: #89AF8F;
	      }
	`)
	_ = prov.LoadFromData(css)

	styleContext, _ := file.label.GetStyleContext()
	styleContext.AddProvider(prov, 9999)

	fileName = "Received: " + fileName
	file.update(fileName)
	file.success = true

	var label string
	count := float64(len(conv.fileTransferNotif.files))
	if conv.fileTransferNotif.progress == count {
		if count == 1 {
			label = "File successfuly received"
		} else {
			label = "Files successfuly received"
		}

		conv.updateFileTransferNotification(label, "Close", "success.svg")

	}
}

func (conv *conversationPane) failFileTransfer(fileName string, file *fileNotification) {
	prov, _ := g.gtk.CssProviderNew()

	css := fmt.Sprintf(`
	label { color: #cc3636;
	     }
	`)
	_ = prov.LoadFromData(css)

	styleContext, _ := file.label.GetStyleContext()
	styleContext.AddProvider(prov, 9999)

	fileName = "Failed: " + fileName
	file.update(fileName)

	for i, f := range conv.fileTransferNotif.files {
		if f == file {
			conv.fileTransferNotif.files = append(conv.fileTransferNotif.files[:i], conv.fileTransferNotif.files[i+1:]...)
		}
	}

	if len(conv.fileTransferNotif.files) == 0 {
		label := "File transfer failed"
		conv.updateFileTransferNotification(label, "Close", "failure.svg")
	}
}

func (conv *conversationPane) cancelFileTransfer(fileName string, file *fileNotification) {
	prov, _ := g.gtk.CssProviderNew()

	css := fmt.Sprintf(`
	label { color: #cc3636;
	      }
	`)
	_ = prov.LoadFromData(css)

	styleContext, _ := file.label.GetStyleContext()
	styleContext.AddProvider(prov, 9999)

	fileName = "Canceled: " + fileName
	file.update(fileName)
	file.canceled = true

	for i, f := range conv.fileTransferNotif.files {
		if f == file {
			conv.fileTransferNotif.files = append(conv.fileTransferNotif.files[:i], conv.fileTransferNotif.files[i+1:]...)
		}
	}

	if len(conv.fileTransferNotif.files) == 0 {
		conv.fileTransferNotif.canceled = true
		label := "File transfer canceled"
		conv.updateFileTransferNotification(label, "Close", "failure.svg")
	}
}

func (conv *conversationPane) isFileTransferNotifCanceled() bool {
	return conv.fileTransferNotif.canceled
}

// TODO: add name of file
func (conv *conversationPane) onDestroyFileTransferNotif() {
	label := conv.fileTransferNotif.labelButton.GetLabel()
	if label == "Cancel" {
		conv.fileTransferNotif.canceled = true
		label := "File transfer canceled"
		conv.updateFileTransferNotification(label, "Close", "failure.svg")

		files := conv.fileTransferNotif.files
		for i, f := range files {
			if f.success {
				break
			}
			files[i].update("Canceled")
		}
	} else {
		conv.fileTransferNotif.canceled = false
		conv.fileTransferNotif.area.SetVisible(false)
		conv.fileTransferNotif.progress = 0.0
		for i := range conv.fileTransferNotif.files {
			conv.fileTransferNotif.files[i].area.Destroy()
		}
	}
}
