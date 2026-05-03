package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type WishlistItem struct {
	UsuarioId int `json:"usuarioId"`
	VueloId   int `json:"vueloId"`
}

var wishlist []WishlistItem

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

	for _, it := range wishlist {
		if it.UsuarioId == item.UsuarioId && it.VueloId == item.VueloId {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"mensaje": "Ya está en la wishlist"}`)
			return
		}
	}

	wishlist = append(wishlist, item)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, `{"mensaje": "Agregado correctamente"}`)
}

func obtener(w http.ResponseWriter, r *http.Request) {
	usuarioId := r.URL.Path[len("/wishlist/"):]
	userId, err := strconv.Atoi(usuarioId)
	if err != nil {
		http.Error(w, `{"error": "Usuario inválido"}`, http.StatusBadRequest)
		return
	}

	ticketsURL := os.Getenv("TICKETS_SERVICE_URL")
	if ticketsURL == "" {
		ticketsURL = "http://localhost:8000"
	}

	var resultado []map[string]interface{}

	for _, it := range wishlist {
		if it.UsuarioId == userId {
			resp, err := http.Get(fmt.Sprintf("%s/vuelos/%d", ticketsURL, it.VueloId))
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

	var nueva []WishlistItem
	for _, it := range wishlist {
		if !(it.UsuarioId == item.UsuarioId && it.VueloId == item.VueloId) {
			nueva = append(nueva, it)
		}
	}

	wishlist = nueva
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"mensaje": "Eliminado"}`)
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

	http.HandleFunc("/wishlist/", enableCORS(obtener))

	fmt.Println("Servidor de Wishlist (Go) corriendo en puerto 8082")
	http.ListenAndServe(":8082", nil)
}