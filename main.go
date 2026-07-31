package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// ⚡ TODOS LOS IMPORTS CORREGIDOS A V2
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ValenRomero24/P3-MP3-Widget/internal/audio"
	"github.com/ValenRomero24/P3-MP3-Widget/internal/domain"
	"github.com/ValenRomero24/P3-MP3-Widget/internal/ui"
)

func formatTrackTitle(title string) string {
	// 1. Quitar la extensión (.flac, .mp3, etc.)
	newTitle := strings.TrimSuffix(title, filepath.Ext(title))

	// 2. Eliminar corchetes/paréntesis y su contenido al final (ej: "[01_The_End]")
	reBrackets := regexp.MustCompile(`\[.*?\]|\(.*?\)`)
	newTitle = reBrackets.ReplaceAllString(newTitle, "")

	// 3. Eliminar números iniciales y guiones bajos (ej: "01_" o "01 - ")
	rePrefix := regexp.MustCompile(`^\d+[\s\._\-]+`)
	newTitle = rePrefix.ReplaceAllString(newTitle, "")

	// 4. Reemplazar guiones bajos restantes por espacios
	newTitle = strings.ReplaceAll(newTitle, "_", " ")
	newTitle = strings.ReplaceAll(newTitle, "-", " ")

	// 5. Limpiar espacios sobrantes a los lados
	newTitle = strings.TrimSpace(newTitle)

	// 6. Truncar si es demasiado largo para que el widget mantenga su tamaño
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
	engine	:= audio.NewBeepEngine()

	currentTrack, _ := manager.CurrentTrack()
	_ = engine.Play(currentTrack.Path)

	a := app.NewWithID("com.valenromero.p3widget")
	a.Settings().SetTheme(&ui.P3Theme{})
	w := a.NewWindow("Persona 3 MP3 Player")

// --- CONFIGURACIÓN DEL WIDGET CORREGIDA ---
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(300, 250))
	
	// Nota: Si Pop!_OS (GNOME) te sigue mostrando el borde superior, 
	// Fyne nos permite sugerirle al sistema operativo que la ventana es de tipo "Splash" 
	// (ventanas de carga que no tienen bordes ni botones de cerrar nativos).
	// Descomentá la línea de abajo si querés forzar que no tenga bordes:
	// w.SetMainMenu(nil)

	lblTitle	:= widget.NewLabelWithStyle("Reproduciendo: " + formatTrackTitle(currentTrack.Title), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	lblTime		:= widget.NewLabel("00:00 / 00:00")

	slider 		:= widget.NewSlider(0, 100)
	isDragging	:= false

	slider.OnChanged = func(val float64){
		isDragging = true
	}  

	slider.OnChangeEnded = func(val float64) {
		_, tot := engine.GetProgress()
		if tot > 0 {
			target := time.Duration(val/100.0 * float64(tot))
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

	playPrevTrack := func(){
		if manager.Prev(){
			t, _ := manager.CurrentTrack()
			_ = engine.Play(t.Path)
			lblTitle.SetText("Reproduciendo: " + formatTrackTitle(t.Title))
		}
	}

	btnPlayPause	:= widget.NewButton("Play/Pause", func(){
		engine.TogglePause()
	})

	btnNext			:= widget.NewButton(">>", func() {
		playNextTrack()
	})

	btnPrevious			:= widget.NewButton("<<", func(){
		playPrevTrack()
	})

	btnSeekBack 	:= widget.NewButton("-5s", func() {
		engine.Seek(-5 * time.Second)
	})
	btnSeekForward	:= widget.NewButton("+5s", func() {
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

	btnPlaylist := widget.NewButton("📋 Lista", func(){
		listContainer := container.NewScroll(trackList)
		listContainer.SetMinSize(fyne.NewSize(300, 180))

		popup = widget.NewModalPopUp(
			container.NewVBox(
				widget.NewLabelWithStyle("Canciones", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				listContainer,
				widget.NewButton("Cerrar", func(){ popup.Hide() }),
			),
			w.Canvas(),
		)
		popup.Show()
	})

	controlsRow := container.NewHBox(btnPrevious, btnSeekBack, btnPlayPause, btnSeekForward, btnNext)

	content := container.NewVBox(
		lblTitle,
		slider,
		lblTime,
		container.NewHBox(controlsRow, btnPlaylist),
		widget.NewButton("Cerrar widget", func(){ a.Quit() }),
	)

	w.SetContent(content)
	w.Show()

	// Goroutine encargada de actualizar el progreso de tiempo de forma segura
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			pos, tot 	:= engine.GetProgress()
			if tot > 0 {
				minP, secP	:= int(pos.Minutes()), int(pos.Seconds())%60
				minT, secT	:= int(tot.Minutes()), int(tot.Seconds())%60
				nuevoTexto	:= fmt.Sprintf("%02d:%02d / %02d:%02d", minP, secP, minT, secT)

				if pos >= tot-500*time.Millisecond && pos > 0 {
					fyne.Do(func(){
						playNextTrack()
					})
					continue
				}

				fyne.Do(func(){
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