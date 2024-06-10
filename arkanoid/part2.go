package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
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

//-------------------------------------------------
// ---------------------STRUCTS--------------------
//-------------------------------------------------

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

// Barra (jugador)
type barra struct {
	pos     pos
	ancho   int
	alto    int
	vel_x   float32
	color   color
	vida    int
	score   int
	pelotas []pelota // Cada jugador tendra un conjunto de pelotas
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

// Metodo para dibujar los ladrillos
func (bloque *ladrillo) Dibujar(ventana []byte) {

	startX := bloque.pos.x - float32(bloque.ancho)/2
	startY := bloque.pos.y - float32(bloque.alto)/2

	// Dibujamos ladrillo con gorrutinas igual que para dibujar la barra
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
	// Si el usuario ingresa las coordenadas para dibujar la barra esa coordenada es el pixel central de la barra, por lo tanto
	// debemos movernos hacia el pixel superior izquierdo de la barra para empezar a dibujar la barra de izquierda a derecha de arriba a abajo
	startX := barra.pos.x - float32(barra.ancho)/2
	startY := barra.pos.y - float32(barra.alto)/2

	// Para dibujar la barra ahora usamos gorrutinas, donde dividimos la barra en pedazos donde cada pedazo la pintara cada gorrutina
	// Declaramos variable 'wg' que espera a q finalicen la ejecucion tantas gorrutinas antes de seguir por el hilo principal
	var wg sync.WaitGroup
	// Solicitamos la cant de cpus q tenemos para q cada cpu se encargue de ejecutar cada gorrutina y aplicar paralelismo
	numCPUs := runtime.NumCPU()
	// A la variable 'wg' le agregamos el numero de CPU, esto quiere decir q para q la variable 'wg' se complete se deben completar
	// tantas gorrutinas como CPUs tengamos
	wg.Add(numCPUs)
	// A la barra que dibujaremos le armamos un arreglo matriz q la represente, su cantidad seria el ancho*alto de la barra
	arrBarra := make([]byte, barra.ancho*barra.alto)
	// Ahora al arreglo q representa la barra la dividimos en tantos pedazos iguales como CPUs tengamos
	pedazoBarra := len(arrBarra) / numCPUs
	// Ahora iteramos por las CPUs y en cada iteracion creamos una gorrutina en una funcion anonima
	for i := 0; i < numCPUs; i++ {
		// Funcion anonima
		go func(i int) {
			// Al entrar en una gorrutina automaticamente decimos q al terminar de ejecutarse la gorrutina se marque
			// como gorrutina finalizada en al variable 'wg' asi la tiene en cuenta
			defer wg.Done()
			// Cada gorrutina tomara una porcion del arreglo total q representa la barra, por tanto en cada
			// gorrutina indicamos el indice inicio y fin de dicho arreglo
			inicio := i * pedazoBarra
			fin := inicio + pedazoBarra + 4 // El +4 lo agregamos para terminar de pintar la ultima porcion de barra esos pixeles extra

			// Ahora iteramos por el pedacito de arreglo de la barra para pintar cada pixel
			for j := inicio; j < fin; j++ {
				// Al iterar por el pedacito de arreglo tenemos q ir sacando las coordenadas x e y, para ello las formulas
				// son las sig...
				x := j % barra.ancho
				y := j / barra.ancho
				// Finalmente al tener las coordenadas X e Y coloreamos ese pixel de barra q pertenece a dicho pedacito de arreglo
				colorear(pos{startX + float32(x), startY + float32(y)}, barra.color, ventana)
			}
		}(i)
	}
	// Esperamos a q finalicen tantas gorrutinas como CPUs tengamos
	wg.Wait()

	graficarVida(*barra, ventana)
	graficarPuntaje(*barra, ventana, 3, 3, pos{float32(anchoVentana) + 225, float32(altoVentana) - 20}, color{255, 255, 255, 255})
}

// Metodo para dibujar pelota
func (pelota *pelota) Dibujar(ventana []byte) {
	// Este metodo se esta ejecutando en una gorrutina, por tanto antes q nada completaremos la variable waitGroup para marcar
	// que se completo esta gorrutina al final de todo
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
func (barra *barra) Movimiento() {
	if barra.teclado[sdl.SCANCODE_LEFT] != 0 { // Presionada flechita ← teclado
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

		// Si el largo de las pelotas q tiene q atajar el jugador es >1 quiere decir que si se le escapa una pelota entonces
		// aun no perdio la vida
		if len(pelota.jugador.pelotas) > 1 {
			// Buscamos el indice de la pelota q se fue de la ventana para eliminarlo del slice de pelotas q tiene el jugador
			for i, v := range pelota.jugador.pelotas {
				if v.pos == *&pelota.pos {
					if i > 0 && i < len(pelota.jugador.pelotas) {
						pelota.jugador.pelotas = append(pelota.jugador.pelotas[:i], pelota.jugador.pelotas[i+1:]...)
					} else if i == 0 {
						pelota.jugador.pelotas = append(pelota.jugador.pelotas[1:])
					}

				}
			}
		} else { // Si la pelota q se le escapa al jugador es unica en la ventana entonces aca el jugador si pierde la vida y se resetea
			pelota.pos.x = float32(anchoVentana) / 2
			pelota.pos.y = float32(altoVentana)/2 + 100
			pelota.vel_x = 0
			pelota.vel_y = 10
			state = start
			pelota.jugador.pos.x = float32(anchoVentana) / 2
			pelota.jugador.pos.y = float32(altoVentana) - 50
			pelota.jugador.vida--
		}

	}

	// Si la pelota choca con la barra...
	if pelota.pos.y+pelota.radio >= pelota.jugador.pos.y-float32(pelota.jugador.alto)/2 && pelota.pos.y+pelota.radio <= pelota.jugador.pos.y+float32(pelota.jugador.alto)/2 {
		if pelota.pos.x >= pelota.jugador.pos.x-float32(pelota.jugador.ancho)/2 && pelota.pos.x <= pelota.jugador.pos.x+float32(pelota.jugador.ancho)/2 {
			velocidades_x := []int{-11, -9, -7, -5, -3, -1, 1, 3, 5, 7, 9, 11}
			configuracion_velocidad(pelota, pelota.jugador, velocidades_x)
		}
	}
}

// Metodo pelota para romper ladrillo al impactar la pelota
func (bola *pelota) impactoLadrillo(ladrillo *ladrillo, ventana []byte, resistenciaColor map[int]color) {

	// Constante para que la pelota al impactar con el ladrillo no se 'meta' tanto en el ladrillo ya que todo es frame por frame
	var refinadoImpacto float32 = 5.0

	// Si el ladrillo es distinto a negro quiere decir que se puede romper
	if ladrillo.resist > 0 {
		nueva_pelota := pelota{
			pos{bola.pos.x, bola.pos.y},
			bola.radio,
			bola.vel_x,
			bola.vel_y,
			color{0, 255, 255, 255}, // BLANCO
			bola.jugador,
		}
		// Si la pelota golpea la cara inferior o superior del ladrillo
		if bola.pos.x >= ladrillo.pos.x-float32(ladrillo.ancho)/2 && bola.pos.x <= ladrillo.pos.x+float32(ladrillo.ancho)/2 {
			// Si golpea la cara inferior
			if bola.pos.y-bola.radio-refinadoImpacto <= ladrillo.pos.y+float32(ladrillo.alto)/2 && bola.pos.y-bola.radio >= ladrillo.pos.y {
				bola.vel_y = -bola.vel_y
				bola.pos.y = ladrillo.pos.y + float32(ladrillo.alto)/2 + bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				bola.jugador.score += ladrillo.extScore

				// Si el usuario marca un score multiplo de 200 entonces le agregamos nueva pelota
				if bola.jugador.score != 0 && bola.jugador.score%200 == 0 {
					bola.jugador.pelotas = append(bola.jugador.pelotas, nueva_pelota)
				}
			}
			// Si golpea la cara superior
			if bola.pos.y+bola.radio+refinadoImpacto >= ladrillo.pos.y-float32(ladrillo.alto)/2 && bola.pos.y+bola.radio <= ladrillo.pos.y {
				bola.vel_y = -bola.vel_y
				bola.pos.y = ladrillo.pos.y - float32(ladrillo.alto)/2 - bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				bola.jugador.score += ladrillo.extScore

				// Si el usuario marca un score multiplo de 200 entonces le agregamos nueva pelota
				if bola.jugador.score != 0 && bola.jugador.score%200 == 0 {
					bola.jugador.pelotas = append(bola.jugador.pelotas, nueva_pelota)
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
				bola.jugador.score += ladrillo.extScore

				// Si el usuario marca un score multiplo de 200 entonces le agregamos nueva pelota
				if bola.jugador.score != 0 && bola.jugador.score%200 == 0 {
					bola.jugador.pelotas = append(bola.jugador.pelotas, nueva_pelota)
				}
			}
			// Si golpea la cara derecha
			if bola.pos.x-bola.radio-refinadoImpacto <= ladrillo.pos.x+float32(ladrillo.ancho)/2 && bola.pos.x-bola.radio >= ladrillo.pos.x {
				bola.vel_x = -bola.vel_x
				bola.pos.x = ladrillo.pos.x + float32(ladrillo.ancho)/2 + bola.radio
				ladrillo.resist--
				ladrillo.color = resistenciaColor[ladrillo.resist]
				bola.jugador.score += ladrillo.extScore

				// Si el usuario marca un score multiplo de 200 entonces le agregamos nueva pelota
				if bola.jugador.score != 0 && bola.jugador.score%200 == 0 {
					bola.jugador.pelotas = append(bola.jugador.pelotas, nueva_pelota)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------------------------
// -------------------------------------FUNCIONES-----------------------------------------------
// ---------------------------------------------------------------------------------------------

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

// Funcion que limpia toda la ventana (Aca aplicamos gorrutinas para q la limpieza sea mas rapida)
func limpieza(ventana []byte) {
	// Sacamos el numeros de CPUs
	numGorrutinas := runtime.NumCPU()
	// Variable para esperar la finalizacion de ejecucion de gorrutinas
	var wg sync.WaitGroup
	// La variable wg va a tener q esperar 'numGorrutinas' gorrutinas ya que vamos a generar tantas gorrutinas como CPUs tengamos
	// Para llevar a cabo el paralelismo
	wg.Add(numGorrutinas)

	// Dividimos la ventana en numGorrutinas para q cada porcion de la ventana la limpie una gorrutina distinta
	porcionVentana := len(ventana) / numGorrutinas

	// Iteramos por la cantidad de CPUs para generar una gorrutina distinta
	for i := 0; i < numGorrutinas; i++ {
		// Creamos gorrutina con funcion anonima
		go func(i int) {
			defer wg.Done()
			start := i * porcionVentana
			end := start + porcionVentana - 1

			for j := start; j < end; j++ {
				// Definimos las coordenadas x e y
				x := j % int(anchoVentana)
				y := (j - x) / int(altoVentana)
				colorear(pos{float32(x), float32(y)}, color{0, 0, 0, 0}, ventana)
			}
		}(i)
	}
	wg.Wait()
}

func diagramar_mapa(coordenada pos, ancho int, alto int, ventana []byte, resistenciaColor map[int]color) []ladrillo {
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
		0, 0, 0, 0, 0, 0, 0, 0, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 1,
	}

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

	return muro
}

// ----------------------------------------------------------------------------
// -----------------------------------MAIN-------------------------------------
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

	// Estado del teclado (arreglo para ver que teclas son presionadas)
	teclado := sdl.GetKeyboardState()

	var jugador barra

	// Definimos la pelota color blanca tambien
	pelota1 := pelota{
		pos:     pos{float32(anchoVentana) / 2, float32(altoVentana)/2 + 100},
		radio:   5,
		vel_x:   0,
		vel_y:   10,
		color:   color{255, 255, 255, 255},
		jugador: &jugador,
	}

	// Definimos la barra color blanca [RGBA] = [255, 255, 255, 255]
	jugador = barra{
		pos{300, 750},
		100,
		10,
		15,
		color{255, 255, 255, 255},
		3,
		0,
		[]pelota{pelota1},
		teclado,
	}

	// Resistencias de ladrillos asociadas a colores
	resistenciaColor := make(map[int]color)
	resistenciaColor[0] = color{0, 0, 0, 0}       // NEGRO
	resistenciaColor[1] = color{0, 0, 0, 255}     // ROJO
	resistenciaColor[2] = color{0, 128, 128, 128} // GRIS CLARO
	resistenciaColor[3] = color{0, 0, 240, 30}    // VERDE BRILLANTE

	// Definimos el muro de ladrillos
	var muro []ladrillo = diagramar_mapa(pos{300, 200}, 50, 20, pixelesVentana, resistenciaColor)
	// Copia del muro de ladrillos para q cada vez q pierda el usuario las 3 vidas reiniciar el mapa de 0
	copiaMuro := make([]ladrillo, len(muro))
	for index, value := range muro {
		copiaMuro[index] = value
	}

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
				for j := 0; j < len(jugador.pelotas); j++ {
					// Por cada pelota creamos una gorrutina
					jugador.pelotas[j].impactoLadrillo(&muro[i], pixelesVentana, resistenciaColor)
				}
			}

			// Actualizamos el mov. de la pelotas q tiene el jugador
			for i := 0; i < len(jugador.pelotas); i++ {
				// Por cada pelota creamos una gorrutina
				llamarMovimiento(&jugador.pelotas[i])
			}

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
		llamarMovimiento(&jugador)

		// Creamos una variable de espera a ejecucion de tantas gorrutinas
		var dl sync.WaitGroup
		// Pido cuantas CPUs tengo
		numCPUs := runtime.NumCPU()
		// Esperamos tantas gorrutinas como CPUs tengamos
		dl.Add(numCPUs)
		// Al tamaño del muro lo divido en numCPUs intervalos + 1 por si nos da numero con coma
		pedazoMuro := len(muro)/numCPUs + 1

		// Itero por las CPUs y para c/u genero una gorrutina nueva
		for cpu := 0; cpu < numCPUs; cpu++ {
			go func(cpu int) {
				defer dl.Done()
				// Inicio del pedazo de muro
				start := cpu * pedazoMuro
				// Fin pedazo muro
				fin := start + pedazoMuro

				// Iteramos de inicio a fin en el muro para dibujar los ladrillos correspondientes a cada gorrutina
				for i := start; i < fin; i++ {
					// Antes de dibujar el ladrillo nos aseguramos que el indice esta dentro del largo del muro total
					// Porque sino nos escapariamos del muro
					if i < len(muro) {
						llamarDibujar(&muro[i], pixelesVentana)
					}
				}
			}(cpu)
		}
		// Esperamos a q la variable 'dl' termine de completarse para q se terminen de ejecutar todas las gorrutinas
		dl.Wait()

		// Dibujamos las pelotas q tiene el jugador
		// Creamos variable q espera a q se completen gorrutinas
		var dp sync.WaitGroup
		// La variable esperara tantas gorrutinas como pelotas tenga el jugador
		dp.Add(len(jugador.pelotas))
		for i := 0; i < len(jugador.pelotas); i++ {
			go func(i int) {
				defer dp.Done()
				llamarDibujar(&jugador.pelotas[i], pixelesVentana)
			}(i)
		}
		// Esperamos a q se complete la variable q espera a q se completen gorrutinas
		dp.Wait()

		// La dibujamos la barra
		llamarDibujar(&jugador, pixelesVentana)

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
