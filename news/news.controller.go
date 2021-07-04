package news

import (
	"fmt"

	authenticator "github.com/Timos-API/Authenticator"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	fmt.Println("News routes registered")

	router.HandleFunc("/", getAllNews).Methods("GET")
	router.HandleFunc("/featured", getFeaturedNews).Methods("GET")
	router.HandleFunc("/project/{id}", getProjectNews).Methods("GET")
	router.HandleFunc("/{id}", getNews).Methods("GET")

	router.HandleFunc("/", authenticator.AuthMiddleware(postNews, []string{"admin"})).Methods("POST")
	router.HandleFunc("/{id}", authenticator.AuthMiddleware(deleteNews, []string{"admin"})).Methods("DELETE")
	router.HandleFunc("/{id}", authenticator.AuthMiddleware(patchNews, []string{"admin"})).Methods("PATCH")
}
