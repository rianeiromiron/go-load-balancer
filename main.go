package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// PeticionTrabajo representa una solicitud que entra vía HTTP
type PeticionTrabajo struct {
	ID        int
	Payload   string
	Respuesta chan string // Canal individual para devolver la respuesta a ESTE cliente HTTP
}

// Configuración del balanceador
const (
	NumWorkers    = 4  // 4 servidores/workers procesando en paralelo
	CapacidadCola = 10 // Búfer máximo de 10 peticiones en espera
)

var (
	colaTrabajos = make(chan PeticionTrabajo, CapacidadCola)
	contadorID   = 0
	muContador   sync.Mutex
)

// Worker: Servidor interno que procesa las peticiones de la cola
func worker(id int, trabajos <-chan PeticionTrabajo, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("   🟢 Worker #%d listo y escuchando peticiones...\n", id)

	for t := range trabajos {
		fmt.Printf("👷 Worker #%d: Procesando petición #%d ('%s')...\n", id, t.ID, t.Payload)

		// Simula tiempo de procesamiento intensivo (1.5 segundos)
		time.Sleep(1500 * time.Millisecond)

		// Envía la respuesta al canal privado de este cliente HTTP
		t.Respuesta <- fmt.Sprintf("✅ [Worker #%d] Petición #%d procesada con éxito a las %s",
			id, t.ID, time.Now().Format("15:04:05"))
	}
	fmt.Printf("   🛑 Worker #%d apagado limpiamente.\n", id)
}

// Handler HTTP: Recibe peticiones de los usuarios en la web
func handlePeticion(w http.ResponseWriter, r *http.Request) {
	// Extraemos el parámetro ?tarea= de la URL
	tarea := r.URL.Query().Get("tarea")
	if tarea == "" {
		tarea = "Tarea-General"
	}

	muContador.Lock()
	contadorID++
	id := contadorID
	muContador.Unlock()

	// Creamos un canal de respuesta exclusivo para esta solicitud
	canalResp := make(chan string)

	peticion := PeticionTrabajo{
		ID:        id,
		Payload:   tarea,
		Respuesta: canalResp,
	}

	// 🚦 BALANCEO DE CARGA Y CONTRAPRESIÓN (BACKPRESSURE) CON SELECT
	select {
	case colaTrabajos <- peticion:
		// Se encoló con éxito: Esperamos la respuesta del worker
		fmt.Printf("📥 [HTTP Router] Petición #%d encolada. Esperando a un worker libre...\n", id)
		resultado := <-canalResp

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, resultado)

	default:
		// ⚠️ LA COLA ESTÁ LLENA: Rechazo instantáneo sin bloquear (HTTP 503)
		fmt.Printf("⚠️ [HTTP Router] ¡COLA SATURADA! Rechazando petición #%d\n", id)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "❌ Error 503: Servidores saturados. Por favor, reintenta en unos momentos.")
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 INICIANDO BALANCEADOR DE CARGA CONCURRENTE EN GO")
	fmt.Println("==================================================")

	var wg sync.WaitGroup

	// 1. Iniciamos el Worker Pool
	for w := 1; w <= NumWorkers; w++ {
		wg.Add(1)
		go worker(w, colaTrabajos, &wg)
	}

	// 2. Configuramos el servidor HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("/procesar", handlePeticion)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 3. Lanzamos el servidor HTTP en segundo plano
	go func() {
		fmt.Println("\n🌐 Servidor Web escuchando en http://localhost:8080/procesar?tarea=mi-tarea")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en servidor HTTP: %v\n", err)
		}
	}()

	// 4. ESCUCHAMOS SEÑAL DE APAGADO LIMPIO (Ctrl + C)
	canalInterrupcion := make(chan os.Signal, 1)
	signal.Notify(canalInterrupcion, os.Interrupt, syscall.SIGTERM)

	<-canalInterrupcion // Se bloquea aquí hasta que presiones Ctrl + C
	fmt.Println("\n\n🛑 [Main] Señal de apagado recibida. Iniciando Graceful Shutdown...")

	// Cerramos el servidor HTTP (no acepta más tráfico nuevo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	// Cerramos la cola de trabajos para que los workers terminen lo pendiente y salgan
	close(colaTrabajos)

	// Esperamos a que los workers activos terminen
	wg.Wait()

	fmt.Println("🏁 [Main] Todos los recursos fueron liberados. ¡Servidor apagado con éxito!")
}
