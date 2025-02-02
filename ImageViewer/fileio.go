package main

import (
	"errors"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/disintegration/imageorient"
)

func (a *App) openFileDialog() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.mainWin)
			return
		}
		if err == nil && reader == nil {
			return
		}

		a.file, err = os.Open(reader.URI().String()[7:])
		if err != nil {
			dialog.ShowError(err, a.mainWin)
			return
		}

		err = a.open(a.file, true)
		if err != nil {
			dialog.ShowError(err, a.mainWin)
			return
		}
		defer reader.Close()
	}, a.mainWin)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".gif"}))

    // Not easy to find how to set start location for openFileDialog
    // https://github.com/fyne-io/fyne/pull/1379/files
    startLocation, err1 := storage.ListerForURI(storage.NewFileURI(a.config.GetString("ImagePath")))
    if err1 != nil {
		fmt.Errorf("Error finding startLocation %v", err1)
    }
    dialog.SetLocation(startLocation)

	dialog.Show()
}

func (a *App) open(file *os.File, folder bool) error {
	defer file.Close()

	// decode and update the image + get image path
	var err error
	a.img.OriginalImage, _, err = imageorient.Decode(file)
	if err != nil {
		return fmt.Errorf("Unable to decode image %v", err)
	}
	a.img.Path = file.Name()
	a.image.Image = a.img.OriginalImage
	a.image.Refresh()

	// get and display FileInfo
	a.img.FileData, err = os.Stat(a.img.Path)
	if err != nil {
		return err
	}

	a.imgSize.SetText(fmt.Sprintf("Size: %.2f Mb", float64(a.img.FileData.Size())/1000000))

	a.imgLastMod.SetText(fmt.Sprintf("Last modified: \n%s", a.img.FileData.ModTime().Format("02-01-2006")))

	// save all images from folder for next/back
	if folder {
		a.img.Directory = filepath.Dir(file.Name())
        a.refreshImagesInFolder(file)
	}

	a.widthLabel.SetText(fmt.Sprintf("Width:   %dpx", a.img.OriginalImage.Bounds().Max.X))
	a.heightLabel.SetText(fmt.Sprintf("Height: %dpx", a.img.OriginalImage.Bounds().Max.Y))

    fileName := strings.Split(a.img.Path, "/")[len(strings.Split(a.img.Path, "/"))-1]
	a.mainWin.SetTitle(fmt.Sprintf("Image Tagger - %v", fileName))
    a.renamePreview.SetText(fileName)

    // Save the image path to the config.
    a.config.Set("imagepath", a.img.Directory)
    a.WriteConfig()

	// append to last opened images
	a.lastOpened = append(a.lastOpened, file.Name())
	a.app.Preferences().SetString("lastOpened", strings.Join(a.lastOpened, ","))

	// reset editing history
	a.img.lastFilters = nil
	a.img.lastFiltersUndone = nil

	a.resetZoom()

	// activate widgets
	a.reset()
	a.resetBtn.Enable()
	a.leftArrow.Enable()
	a.rightArrow.Enable()
	a.confirmArrow.Enable()
	a.deleteBtn.Enable()
	a.renameBtn.Enable()
	a.zoomIn.Enable()
	a.zoomOut.Enable()
	a.resetZoomBtn.Enable()

    for i := 0; i < tagBtnTotal; i++ {
        a.tagBtns[i].Enable()
    }

	return nil
}

func (a *App) saveFileDialog() {
	if a.img.OriginalImage == nil {
		dialog.ShowError(errors.New("no image opened"), a.mainWin)
		return
	}
	if a.img.EditedImage == nil {
		a.apply()
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		err = a.save(writer)
		if err != nil {
			dialog.ShowError(err, a.mainWin)
			return
		}
	}, a.mainWin)
}

func (a *App) save(writer fyne.URIWriteCloser) error {
	if writer == nil {
		return nil
	}

	switch writer.URI().Extension() {
	case ".jpeg":
		jpeg.Encode(writer, a.img.EditedImage, nil)
	case ".jpg":
		jpeg.Encode(writer, a.img.EditedImage, nil)
	case ".png":
		png.Encode(writer, a.img.EditedImage)
	case ".gif":
		gif.Encode(writer, a.img.EditedImage, nil)
	default:
		os.Remove(writer.URI().String()[7:])
		return errors.New("unsupported file extension\n supported extensions: .jpg, .png, .gif")
	}
	return nil
}

func (a *App) deleteFile() {
	if err := os.Remove(a.img.Path); err != nil {
		dialog.NewError(err, a.mainWin)
		return
	}
	if a.img.index == len(a.img.ImagesInFolder)-1 {
		a.nextImage(false, true)
	} else if len(a.img.ImagesInFolder) == 1 {
		a.image.Image = nil
		a.img.EditedImage = nil
		a.img.OriginalImage = nil
		a.rightArrow.Disable()
		a.leftArrow.Disable()
		a.deleteBtn.Disable()
		a.image.Refresh()
	} else {
		a.nextImage(true, true)
	}
}

func (a *App) refreshImagesInFolder(file *os.File) {
    openFolder, _ := os.Open(a.img.Directory)
    a.img.ImagesInFolder, _ = openFolder.Readdirnames(0)
    // filter image files
    imgList := []string{}
    for _, v := range a.img.ImagesInFolder {
        if strings.HasSuffix(strings.ToLower(v), ".png") || strings.HasSuffix(strings.ToLower(v), ".jpg") || strings.HasSuffix(strings.ToLower(v), ".jpeg") || strings.HasSuffix(strings.ToLower(v), ".gif") {
            imgList = append(imgList, v)
        }
    }
    a.img.ImagesInFolder = imgList
    sort.Strings(a.img.ImagesInFolder) // sort array alphabetically

    // get first index value
    for i, v := range a.img.ImagesInFolder {
        if filepath.Base(file.Name()) == v {
            a.img.index = i
        }
    }
}

func (a *App) renameImage(s string) {
    newPath := strings.TrimSuffix(a.img.Path, filepath.Base(a.img.Path)) + s
    if err := os.Rename(a.img.Path, newPath); err != nil {
        dialog.ShowError(fmt.Errorf("failed to rename file: %v", err), a.mainWin)
        return
    }
    a.img.Path = newPath
    a.refreshImagesInFolder(a.file)
    a.mainWin.SetTitle("Image Tagger - " + s)
    //a.mainWin.Canvas().Overlays().Top().Hide()
}

func (a *App) renameDialog() {
	entry := newEnterEntry()
	entry.enterFunc = func(s string) {
        a.renameImage(s)
        a.mainWin.Canvas().Overlays().Top().Hide()
	}

	entry.SetPlaceHolder(filepath.Base(a.img.Path))
	dialog.ShowCustomConfirm("Rename Image", "Ok", "Cancel", container.NewVBox(entry), func(b bool) {
		if b {
			newPath := strings.TrimSuffix(a.img.Path, filepath.Base(a.img.Path)) + entry.Text
			if err := os.Rename(a.img.Path, newPath); err != nil {
				dialog.ShowError(fmt.Errorf("failed to rename file: %v", err), a.mainWin)
				return
			}
			a.img.Path = newPath
			a.mainWin.SetTitle("Image Tagger - " + entry.Text)
		}
	}, a.mainWin)
}
