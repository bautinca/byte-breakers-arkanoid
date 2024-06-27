package main

import (
	"runtime"
	"strconv"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

const anchoVentana = 600
const altoVentana = 800

type estadoJuego int

const (
	start estadoJuego = iota
	play
	loose
	win
)

var state estadoJuego = start

//-------------------------------------------------
// ---------------------STRUCTS--------------------
//-------------------------------------------------

type color struct {
	r byte // ROJO
	g byte // VERDE
	b byte // AZUL
	a byte // TRANSPARENCIA (ALPHA)
}

// Coordenadas
type pos struct {
	x float32
	y float32
}

// Barra (jugador)
type barra struct {
	pos     pos
	ancho   int
	alto    int
	vel_x   float32
	color   color
	vida    int
	score   int
	pelotas []pelota
	teclado []uint8
}

// Pelota
type pelota struct {
	pos     pos
	radio   float32
	vel_x   float32
	vel_y   float32
	color   color
	jugador *barra
}

// Ladrillo
type ladrillo struct {
	pos      pos
	ancho    int
	alto     int
	color    color
	resist   int
	extScore int
}

// ------------------------------------------------------------------------------------
// ----------------------------------INTERFAZ-------------------------------------------
// ------------------------------------------------------------------------------------

// Interfaz structs con metodos Dibujar() -> ladrillo, pelota y barra
type Dibujable interface {
	Dibujar(ventana []byte)
}

// Interfaz structs con metodos Movimiento() -> pelota y barra
type Movible interface {
	Movimiento()
}

// El 'elem' seria algun ladrillo, pelota o barra y dentro de la funcion aplicamos el metodo Dibujar() al respectivo elem
func llamarDibujar(elem Dibujable, ventana []byte) {
	elem.Dibujar(ventana)
}

// El 'elem' seria algun pelota o barra y dentro de la funcion llamamos al metodo Movimiento() para dicho struct
func llamarMovimiento(elem Movible) {
	elem.Movimiento()
}

// ------------------------------------------------------------------------------------
// ----------------------------------METODOS-------------------------------------------
// ------------------------------------------------------------------------------------

// Metodo para dibujar los ladrillo
func (bloque *ladrillo) Dibujar(ventana []byte) {

	startX := bloque.pos.x - float32(bloque.ancho)/2
	startY := bloque.pos.y - float32(bloque.alto)/2

	var wg sync.WaitGroup
	numCPUs := runtime.NumCPU()
	wg.Add(numCPUs)
	arrLadrillo := make([]byte, bloque.alto*bloque.ancho)
	pedazoLadrillo := len(arrLadrillo) / numCPUs

	for i := 0; i < numCPUs; i++ {
		go func(i int) {
			defer wg.Done()
			inicio := i * pedazoLadrillo
			fin := inicio + pedazoLadrillo + 4

			for j := inicio; j < fin; j++ {
				if j < len(arrLadrillo) {
					x := j % bloque.ancho
					y := j / bloque.ancho
					colorear(pos{startX + float32(x), startY + float32(y)}, bloque.color, ventana)
				}

			}
		}(i)
	}
	wg.Wait()
}

// Metodo para dibujar barra
func (barra *barra) Dibujar(ventana []byte) {

	startX := barra.pos.x - float32(barra.ancho)/2
	startY := barra.pos.y - float32(barra.alto)/2

	var wg sync.WaitGroup

	numCPUs := runtime.NumCPU()

	wg.Add(numCPUs)

	arrBarra := make([]byte, barra.ancho*barra.alto)

	pedazoBarra := len(arrBarra) / numCPUs

	for i := 0; i < numCPUs; i++ {

		go func(i int) {

			defer wg.Done()

			inicio := i * pedazoBarra
			fin := inicio + pedazoBarra + 4

			for j := inicio; j < fin; j++ {

				x := j % barra.ancho
				y := j / barra.ancho

				colorear(pos{startX + float32(x), startY + float32(y)}, barra.color, ventana)
			}
		}(i)
	}
	wg.Wait()

	graficarVida(*barra, ventana)
	graficarPuntaje(*barra, ventana, 3, 3, pos{float32(anchoVentana) + 225, float32(altoVentana) - 20}, color{255, 255, 255, 255})
}

func (pelota *pelota) Dibujar(ventana []byte) {
	startX := pelota.pos.x
	startY := pelota.pos.y

	for y := -pelota.radio; y < pelota.radio; y++ {
		for x := -pelota.radio; x < pelota.radio; x++ {
			if x*x+y*y < pelota.radio*pelota.radio {
				colorear(pos{startX + float32(x), startY + float32(y)}, pelota.color, ventana)
			}
		}
	}
}

// Metodo que mueve la barra a los laterales
func (barra *barra) Movimiento() {
	if barra.teclado[sdl.SCANCODE_LEFT] != 0 {
		if barra.pos.x-float32(barra.ancho)/2 > 0 {
			barra.pos.x -= barra.vel_x
		}

	} else if barra.teclado[sdl.SCANCODE_RIGHT] != 0 {
		if barra.pos.x+float32(barra.ancho)/2 < float32(anchoVentana) {
			barra.pos.x += barra.vel_x
		}
	}
}

// Metodo movimiento pelotita
func (pelota *pelota) Movimiento() {

	pelota.pos.x += pelota.vel_x
	pelota.pos.y += pelota.vel_y

	if pelota.pos.y-pelota.radio <= 0 {
		pelota.vel_y = -pelota.vel_y
	}

	if pelota.pos.x-pelota.radio <= 0 || pelota.pos.x+pelota.radio >= float32(anchoVentana) {
		pelota.vel_x = -pelota.vel_x
	}

	if pelota.pos.y >= float32(altoVentana) {

		if len(pelota.jugador.pelotas) > 1 {
			for i, v := range pelota.jugador.pelotas {
				if v.pos == *&pelota.pos {
					if i > 0 && i < len(pelota.jugador.pelotas) {
						pelota.jugador.pelotas = append(pelota.jugador.pelotas[:i], pelota.jugador.pelotas[i+1:]...)
					} else if i == 0 {
						pelota.jugador.pelotas = append(pelota.jugador.pelotas[1:])
					}

				}
			}
		} else {
			pelota.pos.x = float32(anchoVentana) / 2
			pelota.pos.y = float32(altoVentana)/2 + 100
			pelota.vel_x = 0
			pelota.vel_y = 10
			state = start
			pelota.jugador.pos.x = float32(anchoVentana) / 2
			pelota.jugador.pos.y = float32(altoVentana) - 50
			pelota.jugador.vida--

			if pelota.jugador.vida == 0 {
				state = loose
			}
		}

	}

	if pelota.pos.y+pelota.radio >= pelota.jugador.pos.y-float32(pelota.jugador.alto)/2 && pelota.pos.y+pelota.radio <= pelota.jugador.pos.y+float32(pelota.jugador.alto)/2 {
		if pelota.pos.x >= pelota.jugador.pos.x-float32(pelota.jugador.ancho)/2 && pelota.pos.x <= pelota.jugador.pos.x+float32(pelota.jugador.ancho)/2 {
			velocidades_x := []int{-11, -9, -7, -5, -3, -1, 1, 3, 5, 7, 9, 11}
			configuracion_velocidad(pelota, pelota.jugador, velocidades_x)
		}
	}
}

// Metodo pelota para romper ladrillo al impactar la pelota
func (bola *pelota) impactoLadrillo(ladrillo *ladrillo, ventana []byte, resistenciaColor map[int]color) {

	var refinadoImpacto float32 = 5.0

	if ladrillo.resist > 0 {
		// Si la pelota golpea la cara inferior o superior del ladrillo
		if bola.pos.x >= ladrillo.pos.x-float32(ladrillo.ancho)/2 && bola.pos.x <= ladrillo.pos.x+float32(ladrillo.ancho)/2 {
			// Si golpea la cara inferior
			if bola.pos.y-bola.radio-refinadoImpacto <= ladrillo.pos.y+float32(ladrillo.alto)/2 && bola.pos.y-bola.radio >= ladrillo.pos.y {
				bola.vel_y = -bola.vel_y
				bola.pos.y = ladrillo.pos.y + float32(ladrillo.alto)/2 + bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				efecto_puntaje(bola, ladrillo)
				if ladrillo.resist == 0 {
					bola.jugador.score += ladrillo.extScore
				}

			}
			// Si golpea la cara superior
			if bola.pos.y+bola.radio+refinadoImpacto >= ladrillo.pos.y-float32(ladrillo.alto)/2 && bola.pos.y+bola.radio <= ladrillo.pos.y {
				bola.vel_y = -bola.vel_y
				bola.pos.y = ladrillo.pos.y - float32(ladrillo.alto)/2 - bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				efecto_puntaje(bola, ladrillo)
				if ladrillo.resist == 0 {
					bola.jugador.score += ladrillo.extScore
				}
			}

		}
		// Si la pelota golpea la cara izquierda o derecha del ladrillo
		if bola.pos.y >= ladrillo.pos.y-float32(ladrillo.alto)/2 && bola.pos.y <= ladrillo.pos.y+float32(ladrillo.alto)/2 {
			// Si golpea la cara izquierda
			if bola.pos.x+bola.radio+refinadoImpacto >= ladrillo.pos.x-float32(ladrillo.ancho)/2 && bola.pos.x+bola.radio <= ladrillo.pos.x {
				bola.vel_x = -bola.vel_x
				bola.pos.x = ladrillo.pos.x - float32(ladrillo.ancho)/2 - bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				efecto_puntaje(bola, ladrillo)
				if ladrillo.resist == 0 {
					bola.jugador.score += ladrillo.extScore
				}
			}
			// Si golpea la cara derecha
			if bola.pos.x-bola.radio-refinadoImpacto <= ladrillo.pos.x+float32(ladrillo.ancho)/2 && bola.pos.x-bola.radio >= ladrillo.pos.x {
				bola.vel_x = -bola.vel_x
				bola.pos.x = ladrillo.pos.x + float32(ladrillo.ancho)/2 + bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				efecto_puntaje(bola, ladrillo)
				if ladrillo.resist == 0 {
					bola.jugador.score += ladrillo.extScore
				}
			}
		}

	}
}

// ---------------------------------------------------------------------------------------------
// -------------------------------------FUNCIONES-----------------------------------------------
// ---------------------------------------------------------------------------------------------

func efecto_puntaje(bola *pelota, bloque *ladrillo) {
	score_newball := 100
	canal_Pelota := make(chan int)
	nueva_pelota := pelota{
		pos{float32(anchoVentana) / 2, float32(altoVentana)/2 + 100},
		bola.radio,
		0,
		-10,
		color{0, 255, 255, 255}, // BLANCO
		bola.jugador,
	}

	go func() {
		if bola.jugador.score != 0 && bola.jugador.score%score_newball == 0 && bloque.resist == 0 {
			bola.jugador.pelotas = append(bola.jugador.pelotas, nueva_pelota)
			canal_Pelota <- 1
		}
		canal_Pelota <- 0
		close(canal_Pelota)
	}()

	go func() {
		var reduccionBarra int = 5
		aux := <-canal_Pelota
		if aux == 1 {
			bola.jugador.ancho -= reduccionBarra
		}
	}()
}

// Funcion para graficar el score del jugador
func graficarPuntaje(barra barra, ventana []byte, ancho, alto int, coordenada pos, color color) {

	var simbolosNumeros = [][]byte{
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
		{
			0, 0, 1, 0, 0,
			0, 1, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 1, 1, 1, 0,
		},
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			0, 0, 0, 0, 1,
			0, 0, 0, 1, 0,
			0, 0, 1, 0, 0,
			0, 1, 0, 0, 0,
			1, 1, 1, 1, 1,
		},
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			0, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
			0, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
		{
			0, 0, 1, 1, 0,
			0, 1, 0, 1, 0,
			1, 0, 0, 1, 0,
			1, 1, 1, 1, 1,
			0, 0, 0, 1, 0,
			0, 0, 0, 1, 0,
			0, 0, 0, 1, 0,
		},
		{
			1, 1, 1, 1, 1,
			1, 0, 0, 0, 0,
			1, 1, 1, 1, 0,
			0, 0, 0, 0, 1,
			0, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 0,
			1, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
		{
			1, 1, 1, 1, 1,
			0, 0, 0, 0, 1,
			0, 0, 0, 1, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
			0, 0, 1, 0, 0,
		},
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
		{
			0, 1, 1, 1, 0,
			1, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 1,
			0, 0, 0, 0, 1,
			1, 0, 0, 0, 1,
			0, 1, 1, 1, 0,
		},
	}
	// Convertimos el puntaje del jugador a una cadena
	strScore := strconv.Itoa(barra.score)

	// Creamos slice para almacenar los digitos
	digitos := make([]int, len(strScore))

	// Iteramos sobre la cadena, a cada caracter lo pasamos a entero y lo almacenamos en el slice
	for i, v := range strScore {
		digitos[i] = int(v - '0')
	}

	// Para dibujar cada digito crearemos gorrutinas
	var wg sync.WaitGroup
	wg.Add(len(digitos))

	for i, v := range digitos {

		go func(i, v int) {
			defer wg.Done()
			numeroMatriz := simbolosNumeros[v]

			startX := coordenada.x - float32(5*ancho)/2 + float32(i*ancho*6) // Le sumamos float32(i*ancho*6) para q cada digito se dibuje uno separado del otro
			startY := coordenada.y - float32(7*alto)/2

			for index, value := range numeroMatriz {
				if value == 1 {
					for y := startY; y < startY+float32(alto); y++ {
						for x := startX; x < startX+float32(ancho); x++ {
							colorear(pos{x, y}, color, ventana)
						}
					}
				}
				startX += float32(ancho)

				if (index+1)%5 == 0 {
					startY += float32(alto)
					startX -= float32(ancho) * 5
				}
			}
		}(i, v)

	}

	wg.Wait()
}

// Funcion para graficar la vida de la barra
func graficarVida(barra barra, ventana []byte) {

	var vida_grafico = []byte{
		0, 1, 1, 0, 1, 1, 0,
		1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1,
		0, 1, 1, 1, 1, 1, 0,
		0, 0, 1, 1, 1, 0, 0,
		0, 0, 0, 1, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0,
	}

	var wg sync.WaitGroup
	wg.Add(barra.vida)

	for vida := 0; vida < barra.vida; vida++ {

		go func(vida int) {
			defer wg.Done()
			startX := vida*25 + 10 // Sumamos +10 para separarnos del borde izquierdo de la ventana un margen
			startY := altoVentana - 30

			for i, valor := range vida_grafico {
				if valor == 1 {
					for y := startY; y < startY+3; y++ {
						for x := startX; x < startX+3; x++ {
							colorear(pos{float32(x), float32(y)}, color{0, 0, 0, 255}, ventana)
						}
					}
				}

				startX += 3

				if (i+1)%7 == 0 {
					startY += 3
					startX -= 3 * 7
				}
			}
		}(vida)

	}
	wg.Wait()

}

// Configuracion pelota velocidad al impactar con la barra
func configuracion_velocidad(pelota *pelota, jugador *barra, velocidades_x []int) {
	var segmento float32 = float32(jugador.ancho) / float32(len(velocidades_x))

	for indice, velocidad := range velocidades_x {
		if pelota.pos.x <= jugador.pos.x-float32(jugador.ancho)/2+segmento*(float32(indice+1)) {
			pelota.vel_x = float32(velocidad)
			pelota.vel_y = -pelota.vel_y
			pelota.pos.y = jugador.pos.y - float32(jugador.alto)/2 - pelota.radio
			return
		}
	}
}

// Colorear pixel ventana
func colorear(pos pos, c color, ventana []byte) {
	index := (pos.y*float32(anchoVentana) + pos.x) * 4

	if int(index) < len(ventana)-4 && index >= 0 {
		ventana[int(index)] = c.r
		ventana[int(index+1)] = c.g
		ventana[int(index+2)] = c.b
		ventana[int(index+3)] = c.a
	}
}

// Limpieza ventana en negro
func limpieza(ventana []byte) {

	numGorrutinas := runtime.NumCPU()

	var wg sync.WaitGroup

	wg.Add(numGorrutinas)

	porcionVentana := len(ventana) / numGorrutinas

	for i := 0; i < numGorrutinas; i++ {

		go func(i int) {
			defer wg.Done()
			start := i * porcionVentana
			end := start + porcionVentana - 1

			for j := start; j < end; j++ {
				x := j % int(anchoVentana)
				y := (j - x) / int(altoVentana)
				colorear(pos{float32(x), float32(y)}, color{0, 0, 0, 0}, ventana)
			}
		}(i)
	}
	wg.Wait()
}

// Diagramamos muro con todos los ladrillos, sus coordenadas y sus resistencias
func diagramar_mapa(coordenada pos, ancho int, alto int, ventana []byte) ([]ladrillo, map[int]color) {

	var ladrillos = []byte{
		0, 1, 2, 3, 4, 5, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 5, 5, 0,
		1, 1, 1, 1, 1, 5, 5, 0, 5,
		1, 1, 1, 1, 1, 5, 5, 5, 5,
		1, 1, 1, 1, 1, 5, 5, 5, 5,
		1, 1, 1, 1, 1, 5, 5, 5, 5,
		1, 1, 1, 1, 1, 0, 0, 0, 0,
	}

	resistenciaColor := make(map[int]color)
	resistenciaColor[0] = color{0, 0, 0, 0}       // NEGRO
	resistenciaColor[1] = color{0, 152, 152, 255} // ROJO MUY CLARO
	resistenciaColor[2] = color{0, 84, 84, 255}   // ROJO CLARO
	resistenciaColor[3] = color{0, 0, 0, 255}     // ROJO PURO
	resistenciaColor[4] = color{0, 0, 0, 190}     // ROJO OSCURO
	resistenciaColor[5] = color{0, 0, 0, 120}     // ROJO MUY OSCURO

	muro := make([]ladrillo, 9*17) // Ancho*alto muro ladrillos
	startX := int(coordenada.x) - (ancho*9)/2 + ancho/2
	startY := int(coordenada.y) - (alto*17)/2 + alto/2

	// Cada ladrillo lo construimos con gorrutinas distintas
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	for indice, value := range ladrillos {
		// Por cada gorrutina creada creamos 1 gorrutina adicional q debe finalizarse para seguir con el hilo principal
		wg.Add(1)
		go func(indice int, value byte) {
			defer wg.Done()

			x := startX + (indice%9)*(ancho+1)
			y := startY + (indice/9)*(alto+1)

			ladrillo := ladrillo{pos{float32(x), float32(y)}, ancho, alto, resistenciaColor[int(value)], int(value), 10}

			mutex.Lock()
			muro = append(muro, ladrillo)
			mutex.Unlock()
		}(indice, value)
	}
	wg.Wait()

	return muro, resistenciaColor
}

// Copia del muro
func replicaMuro(muro []ladrillo) []ladrillo {
	copiaMuro := make([]ladrillo, len(muro))
	for index, value := range muro {
		copiaMuro[index] = value
	}

	return copiaMuro
}

// Grafica de los ladrillos del muro
func graficarLadrillos(muro []ladrillo, pixelesVentana []byte) {
	var dl sync.WaitGroup
	numCPUs := runtime.NumCPU()
	dl.Add(numCPUs)
	pedazoMuro := len(muro)/numCPUs + 1

	for cpu := 0; cpu < numCPUs; cpu++ {
		go func(cpu int) {
			defer dl.Done()
			start := cpu * pedazoMuro
			fin := start + pedazoMuro

			for i := start; i < fin; i++ {
				if i < len(muro) {
					llamarDibujar(&muro[i], pixelesVentana)
				}
			}
		}(cpu)
	}
	dl.Wait()

	contador := 0
	for _, ladrillo := range muro {
		negro := color{0, 0, 0, 0}
		if ladrillo.color == negro {
			contador++
			if contador == len(muro) {
				state = win
			}
		}
	}
}

// Grafica de las pelotas
func graficarPelotas(jugador barra, pixelesVentana []byte) {
	var dp sync.WaitGroup
	dp.Add(len(jugador.pelotas))
	for i := 0; i < len(jugador.pelotas); i++ {
		go func(i int) {
			defer dp.Done()
			llamarDibujar(&jugador.pelotas[i], pixelesVentana)
		}(i)
	}
	dp.Wait()
}

// Movimiento de las pelotas
func movimientoPelotas(jugador barra) {
	for i := 0; i < len(jugador.pelotas); i++ {
		go llamarMovimiento(&jugador.pelotas[i])
	}
}

// Verifica el estado de todos los ladrillos del muro individualmente si c/u de las pelotas las golpeo o no
func estadoLadrillos(jugador barra, muro []ladrillo, pixelesVentana []byte, resistenciaColor map[int]color) {
	for i, _ := range muro {
		for j := 0; j < len(jugador.pelotas); j++ {
			jugador.pelotas[j].impactoLadrillo(&muro[i], pixelesVentana, resistenciaColor)
		}
	}
}
