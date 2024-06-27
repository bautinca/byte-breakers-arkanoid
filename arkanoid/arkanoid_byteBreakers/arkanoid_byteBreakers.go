package main

import (
	"fmt"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// ----------------------------------------------------------------------------
// -----------------------------------MAIN-------------------------------------
// ----------------------------------------------------------------------------

func main() {

	// Init Texto
	if err := ttf.Init(); err != nil {
		fmt.Println("Error creacion texto:", err)
	}
	defer ttf.Quit()

	// Ventana
	ventana, err := sdl.CreateWindow("Arkanoid ByteBreakers", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, anchoVentana, altoVentana, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println("Error creacion ventana:", err)
		return
	}
	defer ventana.Destroy()

	// Renderizador
	renderizador, err := sdl.CreateRenderer(ventana, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println("Error creacion render:", err)
	}
	defer renderizador.Destroy()

	// Fuente Texto
	font, err := ttf.OpenFont("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 24)
	if err != nil {
		fmt.Println("Error creacion texto:", err)
	}
	defer font.Close()

	// Texturizador
	texturizador, err := renderizador.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, anchoVentana, altoVentana)
	if err != nil {
		fmt.Println("Error creacion texturizador:", err)
	}
	defer texturizador.Destroy()

	// Color Fuente
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	// Ventana donde dibujamos
	pixelesVentana := make([]byte, anchoVentana*altoVentana*4)

	// Teclado
	teclado := sdl.GetKeyboardState()

	var jugador barra

	// Pelota inicial jugador
	pelota1 := pelota{
		pos:     pos{float32(anchoVentana) / 2, float32(altoVentana)/2 + 100},
		radio:   5,
		vel_x:   0,
		vel_y:   10,
		color:   color{255, 255, 255, 255},
		jugador: &jugador,
	}

	// Jugador
	jugador = barra{
		pos{float32(anchoVentana) / 2, float32(altoVentana) - 50},
		100,
		10,
		15,
		color{255, 255, 255, 255},
		3,
		0,
		[]pelota{pelota1},
		teclado,
	}

	// Copiamos los atributos del jugador por si el usuario pierde para resetearlo
	copiaJugador := jugador

	// Diagramacion mapa y resistencias todos los ladrillos
	muro, resistenciaColor := diagramar_mapa(pos{300, 200}, 50, 20, pixelesVentana)

	// Copia mapa para restaurarlo cuando el usuario pierde 3 vidas
	copiaMuro := replicaMuro(muro)

	// -----------------------FOTOGRAMAS-------------------------------
	for {

		for evento := sdl.PollEvent(); evento != nil; evento = sdl.PollEvent() {
			switch evento.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		limpieza(pixelesVentana)

		// Movimiento jugador
		llamarMovimiento(&jugador)

		// Grafica ladrillos (Y verificamos si el usuario gano)
		graficarLadrillos(muro, pixelesVentana)

		//Graficar pelotas
		graficarPelotas(jugador, pixelesVentana)

		// Dibujar jugador
		llamarDibujar(&jugador, pixelesVentana)

		renderizador.Copy(texturizador, nil, nil)
		if err != nil {
			fmt.Println("Error copia textura en renderizador:", err)
		}

		// Switch STATE juego
		switch state {
		// Juego en pausa
		case start:
			if teclado[sdl.SCANCODE_SPACE] != 0 {
				state = play
			}

		// Si el usuario esta jugando
		case play:
			estadoLadrillos(jugador, muro, pixelesVentana, resistenciaColor)

			movimientoPelotas(jugador)

		// Si el usuario gano
		case win:
			limpieza(pixelesVentana)
			textoVictoria := fmt.Sprintf("YOU WIN! SCORE: %d", jugador.score)
			surface, err := font.RenderUTF8Solid(textoVictoria, textColor)
			if err != nil {
				fmt.Println("Error creacion texto:", err)
			}
			defer surface.Free()

			texturaTexto, err := renderizador.CreateTextureFromSurface(surface)
			if err != nil {
				fmt.Println("Error textura texto:", err)
			}
			defer texturaTexto.Destroy()

			renderizador.Copy(texturaTexto, nil, &sdl.Rect{X: (anchoVentana / 2) - 150, Y: altoVentana / 2, W: surface.W, H: surface.H})

			if teclado[sdl.SCANCODE_SPACE] != 0 {
				return
			}

		// Si el usuario perdio
		case loose:
			limpieza(pixelesVentana)
			for index, value := range copiaMuro {
				muro[index] = value
			}

			textoDerrota := fmt.Sprintf("SCORE: %d", jugador.score)

			surface, err := font.RenderUTF8Solid(textoDerrota, textColor)
			if err != nil {
				fmt.Println("Error creacion texto:", err)
			}
			defer surface.Free()

			texturaTexto, err := renderizador.CreateTextureFromSurface(surface)
			if err != nil {
				fmt.Println("Error textura texto:", err)
			}
			defer texturaTexto.Destroy()

			renderizador.Copy(texturaTexto, nil, &sdl.Rect{X: (anchoVentana / 2) - 70, Y: altoVentana / 2, W: surface.W, H: surface.H})

			if teclado[sdl.SCANCODE_SPACE] != 0 {
				jugador = copiaJugador

				for index, value := range copiaMuro {
					muro[index] = value
				}

				state = start
			}

		}

		pixelsPointer := unsafe.Pointer(&pixelesVentana[0])

		texturizador.Update(nil, pixelsPointer, int(anchoVentana)*4)
		if err != nil {
			fmt.Println("Error actualizacion texturizador:", err)
		}

		renderizador.Present()

		sdl.Delay(16)
	}

}
