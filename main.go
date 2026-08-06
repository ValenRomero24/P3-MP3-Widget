package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ValenRomero24/P3-MP3-Widget/internal/audio"
	"github.com/ValenRomero24/P3-MP3-Widget/internal/domain"
	"github.com/ValenRomero24/P3-MP3-Widget/internal/ui"
)

func formatTrackTitle(title string) string {
	newTitle := strings.TrimSuffix(title, filepath.Ext(title))

	reBrackets := regexp.MustCompile(`\[.*?\]|\(.*?\)`)
	newTitle = reBrackets.ReplaceAllString(newTitle, "")

	rePrefix := regexp.MustCompile(`^\d+[\s\._\-]+`)
	newTitle = rePrefix.ReplaceAllString(newTitle, "")

	newTitle = strings.ReplaceAll(newTitle, "_", " ")
	newTitle = strings.ReplaceAll(newTitle, "-", " ")
	newTitle = strings.TrimSpace(newTitle)

	if len(newTitle) > 35 {
		newTitle = newTitle[:32] + "..."
	}

	return newTitle
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("USO: p3-mp3-widget <ruta-directorio-musica>")
		return
	}

	tracks, err := audio.ScanDirectory(os.Args[1])
	if err != nil || len(tracks) == 0 {
		log.Fatalf("No se encontraron canciones válidas.")
	}

	manager := domain.NewPlaylistManager(tracks)
	engine := audio.NewBeepEngine()

	currentTrack, _ := manager.CurrentTrack()
	_ = engine.Play(currentTrack.Path)

	a := app.NewWithID("com.valenromero.p3widget")
	a.Settings().SetTheme(&ui.P3Theme{})

	var w fyne.Window
	if drv, ok := a.Driver().(desktop.Driver); ok {
		w = drv.CreateSplashWindow()
	} else {
		w = a.NewWindow("Persona 3 MP3 Player")
	}

	w.SetFixedSize(false)
	w.Resize(fyne.NewSize(450, 250))

	lblTitle := widget.NewLabelWithStyle("Reproduciendo: "+formatTrackTitle(currentTrack.Title), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	lblTime := widget.NewLabel("00:00 / 00:00")

	slider := widget.NewSlider(0, 100)
	isDragging := false

	slider.OnChanged = func(val float64) {
		isDragging = true
	}

	slider.OnChangeEnded = func(val float64) {
		_, tot := engine.GetProgress()
		if tot > 0 {
			target := time.Duration(val / 100.0 * float64(tot))
			pos, _ := engine.GetProgress()
			dif := target - pos
			engine.Seek(dif)
		}
		isDragging = false
	}

	playNextTrack := func() {
		if manager.Next() {
			t, _ := manager.CurrentTrack()
			_ = engine.Play(t.Path)
			lblTitle.SetText("Reproduciendo: " + formatTrackTitle(t.Title))
		}
	}

	playPrevTrack := func() {
		if manager.Prev() {
			t, _ := manager.CurrentTrack()
			_ = engine.Play(t.Path)
			lblTitle.SetText("Reproduciendo: " + formatTrackTitle(t.Title))
		}
	}

	btnPlayPause := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		engine.TogglePause()
	})

	btnNext := widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {
		playNextTrack()
	})

	btnPrevious := widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
		playPrevTrack()
	})

	btnSeekBack := widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), func() {
		engine.Seek(-5 * time.Second)
	})
	btnSeekForward := widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), func() {
		engine.Seek(5 * time.Second)
	})

	var popup *widget.PopUp
	allTracks := manager.GetTracks()

	trackList := widget.NewList(
		func() int {
			return len(allTracks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Titulo de la canción")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(formatTrackTitle(allTracks[id].Title))
		},
	)

	trackList.OnSelected = func(id widget.ListItemID) {
		if manager.SelectIndex(id) {
			t, _ := manager.CurrentTrack()
			_ = engine.Play(t.Path)
			lblTitle.SetText("Reproduciendo: " + formatTrackTitle(t.Title))
		}
		if popup != nil {
			popup.Hide()
		}
	}

	btnPlaylist := widget.NewButtonWithIcon("Lista", theme.ListIcon(), func() {
		listContainer := container.NewScroll(trackList)
		listContainer.SetMinSize(fyne.NewSize(300, 180))

		popup = widget.NewModalPopUp(
			container.NewVBox(
				widget.NewLabelWithStyle("Canciones", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				listContainer,
				widget.NewButton("Cerrar", func() { popup.Hide() }),
			),
			w.Canvas(),
		)
		popup.Show()
	})

	controlsRow := container.NewHBox(
		layout.NewSpacer(),
		btnPrevious, btnSeekBack, btnPlayPause, btnSeekForward, btnNext,
		layout.NewSpacer(),
	)

	bg := canvas.NewRectangle(color.RGBA{R: 11, G: 14, B: 20, A: 255})
	bg.CornerRadius = 20                                     
	bg.StrokeColor = color.RGBA{R: 0, G: 163, B: 255, A: 255} 
	bg.StrokeWidth = 2.0

	innerContent := container.NewVBox(
		lblTitle,
		slider,
		lblTime,
		container.NewHBox(controlsRow, btnPlaylist),
		widget.NewButton("Cerrar widget", func() { a.Quit() }),
	)

	finalWidget := container.NewStack(
		bg,
		container.NewPadded(innerContent),
	)

	w.SetContent(finalWidget)
	w.Show()

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			pos, tot := engine.GetProgress()
			if tot > 0 {
				minP, secP := int(pos.Minutes()), int(pos.Seconds())%60
				minT, secT := int(tot.Minutes()), int(tot.Seconds())%60
				nuevoTexto := fmt.Sprintf("%02d:%02d / %02d:%02d", minP, secP, minT, secT)

				if pos >= tot-500*time.Millisecond && pos > 0 {
					fyne.Do(func() {
						playNextTrack()
					})
					continue
				}

				fyne.Do(func() {
					lblTime.SetText(nuevoTexto)

					if !isDragging {
						porcentaje := (float64(pos) / float64(tot)) * 100
						slider.Value = porcentaje
						slider.Refresh()
					}
				})
			}
		}
	}()

	a.Run()
}