package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type WishlistItem struct {
	UsuarioId int `json:"usuarioId"`
	VueloId   int `json:"vueloId"`
}

var (
	mu       sync.RWMutex
	wishlist = make(map[int]map[int]bool)
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "healthy", "service": "ms-wishlist-go"}`)
}

func agregar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var item WishlistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, `{"error": "Error al leer JSON"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if wishlist[item.UsuarioId] == nil {
		wishlist[item.UsuarioId] = make(map[int]bool)
	}

	if wishlist[item.UsuarioId][item.VueloId] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"mensaje": "Ya está en la wishlist"}`)
		return
	}

	wishlist[item.UsuarioId][item.VueloId] = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"mensaje": "Agregado correctamente"}`)
}

func obtener(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, `{"error": "Usuario inválido"}`, http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, `{"error": "Usuario inválido"}`, http.StatusBadRequest)
		return
	}

	ticketsURL := os.Getenv("TICKETS_SERVICE_URL")
	if ticketsURL == "" {
		ticketsURL = "http://localhost:8000"
	}

	var resultado []map[string]interface{}

	mu.RLock()
	userVuelos := wishlist[userId]
	mu.RUnlock()

	if userVuelos != nil {
		for vueloId := range userVuelos {
			resp, err := http.Get(fmt.Sprintf("%s/vuelos/%d", ticketsURL, vueloId))
			if err == nil && resp.StatusCode == 200 {
				var vuelo map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&vuelo); err == nil {
					resultado = append(resultado, vuelo)
				}
				resp.Body.Close()
			}
		}
	}

	if resultado == nil {
		resultado = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultado)
}

func eliminar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var item WishlistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, `{"error": "Error al leer JSON"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if wishlist[item.UsuarioId] != nil {
		delete(wishlist[item.UsuarioId], item.VueloId)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"mensaje": "Item eliminado del carrito"}`)
}

func vaciarCarrito(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, `{"error": "Ruta inválida"}`, http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, `{"error": "Usuario inválido"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	delete(wishlist, userId)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"mensaje": "Carrito reiniciado por logout"}`)
}

func main() {
	http.HandleFunc("/health", enableCORS(healthCheck))

	http.HandleFunc("/wishlist", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			agregar(w, r)
		case "DELETE":
			eliminar(w, r)
		default:
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
	}))

	// Maneja las peticiones con ID en la URL
	http.HandleFunc("/wishlist/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/clear") && r.Method == "DELETE" {
			vaciarCarrito(w, r)
		} else if r.Method == "GET" {
			obtener(w, r)
		} else {
			http.Error(w, `{"error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		}
	}))

	fmt.Println("Servidor de Wishlist (Go) corriendo en puerto 8082")
	http.ListenAndServe(":8082", nil)
}