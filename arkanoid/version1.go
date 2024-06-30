package main

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var anchoVentana int32 = 600
var altoVentana int32 = 800

type estadoJuego int

const (
	start estadoJuego = iota
	play
)

var state estadoJuego = start

// ---------------------STRUCTS--------------------

// Color: Sabemos que el color de cada pixel es formato [RGBA] y por cada parametro pesa 8 bits = 1byte
// por lo tanto los tipos de dato son bytes, solo tomaremos los colores RGB, el Alpha (transparencia) no nos interesa
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

// Barra
type barra struct {
	pos   pos
	ancho int
	alto  int
	vel_x float32
	color color
	vida  int
	score int
}

// Pelota
type pelota struct {
	pos   pos
	radio float32
	vel_x float32
	vel_y float32
	color color
}

// Ladrillo
type ladrillo struct {
	pos    pos
	ancho  int
	alto   int
	color  color
	resist int
}

// -------------------------------------------------------------------------------------

// Metodo para dibujar los ladrillos
func (ladrillo *ladrillo) dibujar_ladrillo(ventana []byte) {
	startX := ladrillo.pos.x - float32(ladrillo.ancho)/2
	startY := ladrillo.pos.y - float32(ladrillo.alto)/2

	for y := 0; y < ladrillo.alto; y++ {
		for x := 0; x < ladrillo.ancho; x++ {
			colorear(pos{startX + float32(x), startY + float32(y)}, ladrillo.color, ventana)
		}
	}
}

// Metodo para dibujar barra
func (barra *barra) dibujar_barra(ventana []byte) {
	// Si el usuario ingresa las coordenadas para dibujar la barra esa coordenada es el pixel central de la barra, por lo tanto
	// debemos movernos hacia el pixel superior izquierdo de la barra para empezar a dibujar la barra de izquierda a derecha de arriba a abajo
	startX := barra.pos.x - float32(barra.ancho)/2
	startY := barra.pos.y - float32(barra.alto)/2

	// Ahora doble for iterando por toda la barra para pintarla
	for y := 0; y < barra.alto; y++ {
		for x := 0; x < barra.ancho; x++ {
			colorear(pos{startX + float32(x), startY + float32(y)}, barra.color, ventana)
		}
	}

	graficarVida(*barra, ventana)
	graficarPuntaje(*barra, ventana, 3, 3, pos{float32(anchoVentana) + 225, float32(altoVentana) - 20}, color{255, 255, 255, 255})
}

// Metodo para dibujar pelota
func (pelota *pelota) dibujar_pelota(ventana []byte) {
	// Para dibujar la pelota imaginamos un cuadrado que lo encierra, por tanto dibujaremos este cuadrado
	// pero para cada pixel colocamos un condicional para saber si este pixel es menor o igual al radio de la pelota para pintarlo
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
func (barra *barra) movimientoBarra(teclado []uint8) {
	if teclado[sdl.SCANCODE_LEFT] != 0 { // Presionada flechita ← teclado
		if barra.pos.x-float32(barra.ancho)/2 > 0 {
			barra.pos.x -= barra.vel_x
		}

	} else if teclado[sdl.SCANCODE_RIGHT] != 0 {
		if barra.pos.x+float32(barra.ancho)/2 < float32(anchoVentana) {
			barra.pos.x += barra.vel_x
		}
	}
}

// Metodo movimiento pelotita
func (pelota *pelota) movimientoPelota(jugador *barra) {
	pelota.pos.x += pelota.vel_x
	pelota.pos.y += pelota.vel_y

	// Si la pelota choca con el borde superior de la ventana entonces rebota, cambia su velocidad en Y en sentido contrario
	if pelota.pos.y-pelota.radio <= 0 {
		pelota.vel_y = -pelota.vel_y
	}

	// Si la pelota choca en cualquiera de los 2 bordes laterales de la ventana entonces cambia su velocidad en X
	if pelota.pos.x-pelota.radio <= 0 || pelota.pos.x+pelota.radio >= float32(anchoVentana) {
		pelota.vel_x = -pelota.vel_x
	}

	// Si la pelota se escaba por el borde inferior de la ventana entonces se restaura en el punto central de la ventana
	if pelota.pos.y >= float32(altoVentana) {
		pelota.pos.x = float32(anchoVentana) / 2
		pelota.pos.y = float32(altoVentana)/2 + 100
		pelota.vel_x = 0
		pelota.vel_y = 10
		state = start
		jugador.pos.x = float32(anchoVentana) / 2
		jugador.pos.y = float32(altoVentana) - 50
		jugador.vida--
	}

	// Si la pelota choca con la barra...
	if pelota.pos.y+pelota.radio >= jugador.pos.y-float32(jugador.alto)/2 && pelota.pos.y+pelota.radio <= jugador.pos.y+float32(jugador.alto)/2 {
		if pelota.pos.x >= jugador.pos.x-float32(jugador.ancho)/2 && pelota.pos.x <= jugador.pos.x+float32(jugador.ancho)/2 {
			velocidades_x := []int{-15, -13 - 11, -9, -7, -5, -3, -1, 1, 3, 5, 7, 9, 11, 13, 15}
			configuracion_velocidad(pelota, jugador, velocidades_x)
		}
	}
}

// -----------------------------------------------------------------------------------------------------------
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

	for i, v := range digitos {
		numeroMatriz := simbolosNumeros[v]

		startX := coordenada.x - float32(5*ancho)/2 + float32(i*ancho*6)
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

	}
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

	for vida := 0; vida < barra.vida; vida++ {
		startX := 10 + vida*25
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
	}

}

// Funcion para romper ladrillo al impactar la pelota
func impactoLadrillo(jugador *barra, ladrillo *ladrillo, pelota *pelota, ventana []byte, resistenciaColor map[int]color) {

	// Constante para que la pelota al impactar con el ladrillo no se 'meta' tanto en el ladrillo ya que todo es frame por frame
	var refinadoImpacto float32 = 5.0

	// Si el ladrillo es distinto a negro quiere decir que se puede romper
	if ladrillo.resist > 0 {

		// Si la pelota golpea la cara inferior o superior del ladrillo
		if pelota.pos.x >= ladrillo.pos.x-float32(ladrillo.ancho)/2 && pelota.pos.x <= ladrillo.pos.x+float32(ladrillo.ancho)/2 {
			// Si golpea la cara inferior
			if pelota.pos.y-pelota.radio-refinadoImpacto <= ladrillo.pos.y+float32(ladrillo.alto)/2 && pelota.pos.y-pelota.radio >= ladrillo.pos.y {
				pelota.vel_y = -pelota.vel_y
				pelota.pos.y = ladrillo.pos.y + float32(ladrillo.alto)/2 + pelota.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				jugador.score += 10
			}
			// Si golpea la cara superior
			if pelota.pos.y+pelota.radio+refinadoImpacto >= ladrillo.pos.y-float32(ladrillo.alto)/2 && pelota.pos.y+pelota.radio <= ladrillo.pos.y {
				pelota.vel_y = -pelota.vel_y
				pelota.pos.y = ladrillo.pos.y - float32(ladrillo.alto)/2 - pelota.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				jugador.score += 10
			}
		}
		// Si la pelota golpea la cara izquierda o derecha del ladrillo
		if pelota.pos.y >= ladrillo.pos.y-float32(ladrillo.alto)/2 && pelota.pos.y <= ladrillo.pos.y+float32(ladrillo.alto)/2 {
			// Si golpea la cara izquierda
			if pelota.pos.x+pelota.radio+refinadoImpacto >= ladrillo.pos.x-float32(ladrillo.ancho)/2 && pelota.pos.x+pelota.radio <= ladrillo.pos.x {
				pelota.vel_x = -pelota.vel_x
				pelota.pos.x = ladrillo.pos.x - float32(ladrillo.ancho)/2 - pelota.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				jugador.score += 10
			}
			// Si golpea la cara derecha
			if pelota.pos.x-pelota.radio-refinadoImpacto <= ladrillo.pos.x+float32(ladrillo.ancho)/2 && pelota.pos.x-pelota.radio >= ladrillo.pos.x {
				pelota.vel_x = -pelota.vel_x
				pelota.pos.x = ladrillo.pos.x + float32(ladrillo.ancho)/2 + pelota.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				jugador.score += 10
			}
		}
	}
}

// Funcion que modifica la velocidad de la pelota segun en que sector choque de la barra
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

// Funcion que colorea un pixel de la ventana
func colorear(pos pos, c color, ventana []byte) {
	// Vamos a ir al indice en la matriz de bytes 'ventana', para ello a la coordenada en Y la multiplicamos por el ancho de la ventana
	// esto es para bajar en filas de la matriz, y luego le sumamos la coordenada X para avanzar en columnas en la matriz hasta llegar al indice indicado
	// Recordar que la matriz es al fin y al cabo un slice fijo, osea un arreglo. Y al final de todo lo multiplicamos por 4 porque cada pixel tiene 4 parametros, [RGBA]
	index := (pos.y*float32(anchoVentana) + pos.x) * 4

	// Consultamos que el indice de la matriz (arreglo) sea valido, para ello el indice no debe ser menor a 0, ni superar el largo total de la matriz
	if int(index) < len(ventana)-4 && index >= 0 {
		ventana[int(index)] = c.r
		ventana[int(index+1)] = c.g // Pintamos el pixel, para ello tener en cuenta que cada pixel esta representado por 4 parametros [RGBA] por tanto
		// cada pixel en el arreglo (matriz) sera 4 indices seguidos
		ventana[int(index+2)] = c.b
		ventana[int(index+3)] = c.a
	}
}

// Funcion que limpia toda la ventana
func limpieza(ventana []byte) {
	for x := 0; x < int(anchoVentana); x++ {
		for y := 0; y < int(altoVentana); y++ {
			colorear(pos{float32(x), float32(y)}, color{0, 0, 0, 0}, ventana)
		}
	}
}

// Funcion generadora Mapa ladrillos, cada ladrillo tiene su valor de resistencia
var ladrillos = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 2, 1, 1, 1, 1, 1,
	1, 1, 3, 1, 2, 1, 1, 1, 1,
	1, 3, 1, 1, 1, 2, 1, 1, 1,
	3, 1, 1, 1, 1, 1, 2, 1, 1,
	1, 3, 1, 1, 1, 1, 1, 2, 1,
	1, 1, 3, 1, 1, 1, 1, 1, 2,
	1, 1, 1, 3, 1, 1, 1, 2, 1,
	1, 1, 1, 1, 3, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 1,
}

func dibujar_mapa(coordenada pos, ancho int, alto int, ventana []byte, resistenciaColor map[int]color) []ladrillo {

	muro := make([]ladrillo, 9*17) // Ancho*alto muro ladrillos
	startX := int(coordenada.x) - (ancho*9)/2 + ancho/2
	startY := int(coordenada.y) - (alto*17)/2 + alto/2

	for indice, value := range ladrillos {
		ladrillo := ladrillo{pos{float32(startX), float32(startY)}, ancho, alto, resistenciaColor[int(value)], int(value)}
		muro = append(muro, ladrillo)
		startX += ancho + 1

		if (indice+1)%9 == 0 {
			startY += alto + 1
			startX -= (ancho + 1) * 9
		}
	}

	return muro
}

// ----------------------------------------------------------------------------

func main() {
	// Creacion Ventana
	ventana, err := sdl.CreateWindow("Arkanoid ByteBreakers", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, anchoVentana, altoVentana, sdl.WINDOW_SHOWN)

	// Verificacion error creacion ventana
	if err != nil {
		fmt.Println("Error creacion ventana:", err)
		return
	}
	// Antes del final del main destruimos la ventana con defer
	defer ventana.Destroy()

	// Creamos un renderizador
	renderizador, err := sdl.CreateRenderer(ventana, -1, sdl.RENDERER_ACCELERATED)

	// Verificacion error creaccion render
	if err != nil {
		fmt.Println("Error creacion render:", err)
	}
	defer renderizador.Destroy()

	// Creacion texturizador, los formatos de pixeles son [RGBA] = [Red, Green, Blue, Alpha] donde c/u pesa 8 bits
	texturizador, err := renderizador.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, anchoVentana, altoVentana)

	// Verificamos si el texturizador dio error
	if err != nil {
		fmt.Println("Error creacion texturizador:", err)
	}
	defer texturizador.Destroy()

	// Representamos la ventana en una matriz de anchoVentana*altoVentana donde cada dato es tipo byte=uint8=8 bits = 1byt, al fin y al cabo
	// la matriz es un arreglo, o slice pero con solo capacidad inicial fija, osea no se reallocara nunca
	// Ademas a la dimensionalidad de la matriz la multiplicamos por 4 ya que cada pixel tiene un largo 4, donde cada pixel tiene los componentes [RGBA] donde cada componente
	// pesa 8 bits, por lo tanto se puede decir que cada pixel pesa 32 bits = 4 bytes
	pixelesVentana := make([]byte, anchoVentana*altoVentana*4)

	// Definimos la barra color blanca [RGBA] = [255, 255, 255, 255]
	jugador := barra{pos{300, 750}, 100, 10, 15, color{255, 255, 255, 255}, 3, 0}

	// Definimos la pelota color blanca tambien
	pelota1 := pelota{
		pos:   pos{float32(anchoVentana) / 2, float32(altoVentana)/2 + 100},
		radio: 5,
		vel_x: 0,
		vel_y: 10,
		color: color{255, 255, 255, 255},
	}

	// Resistencias de ladrillos asociadas a colores
	resistenciaColor := make(map[int]color)
	resistenciaColor[0] = color{0, 0, 0, 0}       // NEGRO
	resistenciaColor[1] = color{0, 0, 0, 255}     // ROJO
	resistenciaColor[2] = color{0, 128, 128, 128} // GRIS CLARO
	resistenciaColor[3] = color{0, 30, 30, 30}    // GRIS OSCURO

	// Definimos el muro de ladrillos
	var muro []ladrillo = dibujar_mapa(pos{300, 200}, 50, 20, pixelesVentana, resistenciaColor)
	// Copia del muro de ladrillos para q cada vez q pierda el usuario las 3 vidas reiniciar el mapa de 0
	copiaMuro := make([]ladrillo, len(muro))
	for index, value := range muro {
		copiaMuro[index] = value
	}

	// Estado del teclado (arreglo para ver que teclas son presionadas)
	teclado := sdl.GetKeyboardState()
	// Iteracion en fotogramas
	for {

		// Solicitamos los eventos que hace el usuario, como presionar teclas del teclado o botones del mouse
		for evento := sdl.PollEvent(); evento != nil; evento = sdl.PollEvent() {
			// switch case para evaluar cada evento
			switch evento.(type) {
			// Si el evento es de tipo exit entonces finalizamos main
			case *sdl.QuitEvent:
				return
			}
		}

		// Si el estado del juego esta jugando
		if state == play {

			// Dibujamos el muro de ladrillos
			for i, _ := range muro {
				impactoLadrillo(&jugador, &muro[i], &pelota1, pixelesVentana, resistenciaColor)
			}
			// Actualizamos el mov. de la pelota
			pelota1.movimientoPelota(&jugador)

			// Si el estado del juego esta en start (pausa)
		} else if state == start {
			// Al momento que el usuario presiona la tecla ESPACIO entramos a este if
			if teclado[sdl.SCANCODE_SPACE] != 0 {
				// Si la vida del juegador esta en 0 quiere decir que perdio, por tanto reseteamos los ladrillos
				if jugador.vida == 0 {
					for index, value := range copiaMuro {
						muro[index] = value
					}
					// Restauramos la vida del jugador
					jugador.vida = 3
					// Y restauramos el puntaje
					jugador.score = 0
				}
				// Finalmente cuando el usuario presiona ESPACIO el estado del juego pasara de start a play, por lo que ahora la pelota se movera y los ladrillos sentiran el impacto
				state = play
			}

		}
		// Limpieza del fotograma nuevo antes de dibujar todo de 0 (da la sensacion de movimiento los objetos que se desplazan)
		limpieza(pixelesVentana)
		// Actualizamos la barra para que se desplaze a los laterales
		jugador.movimientoBarra(teclado)

		for i, _ := range muro {
			muro[i].dibujar_ladrillo(pixelesVentana)
		}
		// Dibujamos la pelota
		pelota1.dibujar_pelota(pixelesVentana)

		// La dibujamos la barra
		jugador.dibujar_barra(pixelesVentana)

		// Este paso lo tenemos que hacer si o si
		pixelsPointer := unsafe.Pointer(&pixelesVentana[0])

		// Actualizamos el texturizador en la ventana de pixeles
		texturizador.Update(nil, pixelsPointer, int(anchoVentana)*4)
		// Si sucedio algun error...
		if err != nil {
			fmt.Println("Error actualizacion texturizador:", err)
		}

		// Copiamos el texturizador en el renderizador
		renderizador.Copy(texturizador, nil, nil)
		// Si fallo...
		if err != nil {
			fmt.Println("Error copia textura en renderizador:", err)
		}

		// Presentamos el render de la textura en la ventana
		renderizador.Present()

		// Colocamos un delay largo para que aguante la ventana abierta
		sdl.Delay(16)
	}

}
