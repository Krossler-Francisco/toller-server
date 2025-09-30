package server

import (
	"fmt"
	"net/http"
)

func Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TollerChat corriendo!")
	})

	fmt.Println("🌍 Servidor en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
