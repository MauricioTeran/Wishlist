package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type WishlistItem struct {
	UsuarioId int `json:"usuarioId"`
	VueloId   int `json:"vueloId"`
}

type VuelosResponse struct {
	Items []map[string]interface{} `json:"items"`
}

var wishlist []WishlistItem

func agregar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var item WishlistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Error al leer JSON", http.StatusBadRequest)
		return
	}

	wishlist = append(wishlist, item)

	fmt.Println(">> Agregado:", item)
	fmt.Println(">> Wishlist actual:", wishlist)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Agregado")
}

func obtener(w http.ResponseWriter, r *http.Request) {

	// Obtener usuarioId
	usuarioId := r.URL.Path[len("/wishlist/"):]
	userId, err := strconv.Atoi(usuarioId)
	if err != nil {
		http.Error(w, "Usuario inválido", http.StatusBadRequest)
		return
	}

	var items []WishlistItem
	for _, it := range wishlist {
		if it.UsuarioId == userId {
			items = append(items, it)
		}
	}

	fmt.Println(">> Wishlist completa:", wishlist)
	fmt.Println(">> Items filtrados:", items)

	// Llamar a FastAPI
	resp, err := http.Get("http://localhost:8000/vuelos")
	if err != nil {
		http.Error(w, "Error llamando tickets", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error leyendo respuesta", http.StatusInternalServerError)
		return
	}

	var data VuelosResponse
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parseando JSON", http.StatusInternalServerError)
		return
	}

	vuelos := data.Items
	fmt.Println(">> Vuelos recibidos:", len(vuelos))

	var resultado []map[string]interface{}

	for _, item := range items {
		for _, vuelo := range vuelos {

			idRaw := vuelo["id_vuelo"]
			if idRaw == nil {
				idRaw = vuelo["id"]
			}

			var id int
			switch v := idRaw.(type) {
			case float64:
				id = int(v)
			case int:
				id = v
			default:
				continue
			}

			fmt.Println(">> Comparando:", id, "con", item.VueloId)

			if id == item.VueloId {
				fmt.Println(">> MATCH encontrado")
				resultado = append(resultado, vuelo)
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
		http.Error(w, "Error al leer JSON", http.StatusBadRequest)
		return
	}

	var nueva []WishlistItem
	for _, it := range wishlist {
		if !(it.UsuarioId == item.UsuarioId && it.VueloId == item.VueloId) {
			nueva = append(nueva, it)
		}
	}

	wishlist = nueva

	fmt.Println(">> Eliminado:", item)
	fmt.Println(">> Wishlist actual:", wishlist)

	fmt.Fprintln(w, "Eliminado")
}

func main() {

	http.HandleFunc("/wishlist", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			agregar(w, r)
		case "DELETE":
			eliminar(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/wishlist/", obtener)

	fmt.Println("Servidor corriendo en puerto 8082")
	http.ListenAndServe(":8082", nil)
}
