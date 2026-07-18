package main

import (
	"fmt"
	"log"
	"os"
	"time"

	// ⚡ TODOS LOS IMPORTS CORREGIDOS A V2
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ValenRomero24/P3-MP3-Widget/internal/audio"
	"github.com/ValenRomero24/P3-MP3-Widget/internal/domain"
)

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
	w := a.NewWindow("Persona 3 MP3 Player")

// --- CONFIGURACIÓN DEL WIDGET CORREGIDA ---
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(350, 150))
	
	// Nota: Si Pop!_OS (GNOME) te sigue mostrando el borde superior, 
	// Fyne nos permite sugerirle al sistema operativo que la ventana es de tipo "Splash" 
	// (ventanas de carga que no tienen bordes ni botones de cerrar nativos).
	// Descomentá la línea de abajo si querés forzar que no tenga bordes:
	// w.SetMainMenu(nil)

	lblTitle	:= widget.NewLabel("Reproduciendo: " + currentTrack.Title)
	lblTime		:= widget.NewLabel("00:00 / 00:00")

	btnPlayPause	:= widget.NewButton("Play/Pause", func(){
		engine.TogglePause()
	})

	btnNext			:= widget.NewButton(">>", func() {
		if manager.Next() {
			t, _ := manager.CurrentTrack()
			_ = engine.Play(t.Path)
			lblTitle.SetText("Reproduciendo: " + t.Title)
		}
	})

	btnSeekBack		:= widget.NewButton("-5s", func() {
		engine.Seek(-5 * time.Second)
	})
	btnSeekForward	:= widget.NewButton("+5s", func() {
		engine.Seek(5 * time.Second)
	})

	content := container.NewVBox(
		lblTitle,
		lblTime,
		container.NewHBox(btnSeekBack, btnPlayPause, btnSeekForward, btnNext),
		widget.NewButton("Cerrar widget", func(){ a.Quit() }),
	)

	w.SetContent(content)

	// Goroutine encargada de actualizar el progreso de tiempo de forma segura
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			pos, tot := engine.GetProgress()
			minP, secP := int(pos.Minutes()), int(pos.Seconds())%60
			minT, secT := int(tot.Minutes()), int(tot.Seconds())%60
			nuevoTexto := fmt.Sprintf("%02d:%02d / %02d:%02d", minP, secP, minT, secT)
			
			// ⚡ CORRECCIÓN ABSOLUTA: Despachamos la actualización al hilo principal de la UI
			fyne.DoAndWait(func() {
				lblTime.SetText(nuevoTexto)
			})
		}
	}()

	w.ShowAndRun()
}