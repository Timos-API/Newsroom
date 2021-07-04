package news

import (
	"fmt"

	authenticator "github.com/Timos-API/Authenticator"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	fmt.Println("News routes registered")
	s := router.PathPrefix("/newsroom").Subrouter()

	s.HandleFunc("", getAllNews).Methods("GET")
	s.HandleFunc("/featured", getFeaturedNews).Methods("GET")
	s.HandleFunc("/projects", getProjects).Methods("GET")
	s.HandleFunc("/project/{id}", getProjectNews).Methods("GET")
	s.HandleFunc("/{id}", getNews).Methods("GET")

	s.HandleFunc("", authenticator.AuthMiddleware(postNews, []string{"admin"})).Methods("POST")
	s.HandleFunc("/{id}", authenticator.AuthMiddleware(deleteNews, []string{"admin"})).Methods("DELETE")
	s.HandleFunc("/{id}", authenticator.AuthMiddleware(patchNews, []string{"admin"})).Methods("PATCH")
}
