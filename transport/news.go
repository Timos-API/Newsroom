package transport

import (
	"Timos-API/Newsroom/persistence"
	"Timos-API/Newsroom/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Timos-API/authenticator"
	"github.com/gorilla/mux"
)

type NewsController struct {
	s *service.NewsService
}

func NewNewsController(s *service.NewsService) *NewsController {
	return &NewsController{s}
}

func (c *NewsController) RegisterNewsRoutes(router *mux.Router) {
	router.HandleFunc("/newsroom", c.getAllNews).Methods("GET")
	router.HandleFunc("/newsroom/featured", c.getFeaturedNews).Methods("GET")
	router.HandleFunc("/newsroom/projects", c.getProjects).Methods("GET")
	router.HandleFunc("/newsroom/project/{id}", c.getProjectNews).Methods("GET")
	router.HandleFunc("/newsroom/{id}", c.getNews).Methods("GET")

	router.HandleFunc("/newsroom", authenticator.Middleware(c.postNews, authenticator.Guard().G("admin").P("newsroom.post"))).Methods("POST")
	router.HandleFunc("/newsroom/{id}", authenticator.Middleware(c.deleteNews, authenticator.Guard().G("admin").P("newsroom.delete"))).Methods("DELETE")
	router.HandleFunc("/newsroom/{id}", authenticator.Middleware(c.patchNews, authenticator.Guard().G("admin").P("newsroom.patch"))).Methods("PATCH")
}

func (c *NewsController) printError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (c *NewsController) extractQueryLimitSkip(req *http.Request) (*string, *int, *int) {
	q := req.URL.Query()
	qLimit, qSkip, qQuery := q.Get("limit"), q.Get("skip"), q.Get("query")
	var query *string
	var limit *int
	var skip *int

	if len(qQuery) > 0 {
		query = &qQuery
	}

	if len(qLimit) > 0 {
		if l, err := strconv.Atoi(qLimit); err == nil {
			limit = &l
		}
	}

	if len(qSkip) > 0 {
		if s, err := strconv.Atoi(qSkip); err == nil {
			skip = &s
		}
	}
	return query, limit, skip
}

func (c *NewsController) getAllNews(w http.ResponseWriter, req *http.Request) {
	query, limit, skip := c.extractQueryLimitSkip(req)

	news, err := c.s.GetAllNews(req.Context(), query, limit, skip)
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) getFeaturedNews(w http.ResponseWriter, req *http.Request) {
	news, err := c.s.GetFeaturedNews(req.Context())
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) getProjects(w http.ResponseWriter, req *http.Request) {
	projects, err := c.s.GetProjects(req.Context())
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(projects)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) getProjectNews(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "missing param: id", http.StatusBadRequest)
		return
	}

	query, limit, skip := c.extractQueryLimitSkip(req)

	news, err := c.s.GetProjectNews(req.Context(), id, query, limit, skip)
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) getNews(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "missing param: id", http.StatusBadRequest)
		return
	}

	news, err := c.s.GetNews(req.Context(), id)
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) postNews(w http.ResponseWriter, req *http.Request) {
	var content persistence.News
	err := json.NewDecoder(req.Body).Decode(&content)
	if err != nil {
		c.printError(w, err)
		return
	}

	news, err := c.s.PostNews(req.Context(), content)
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}

func (c *NewsController) deleteNews(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "missing param: id", http.StatusBadRequest)
		return
	}

	success, err := c.s.DeleteNews(req.Context(), id)
	if err != nil {
		c.printError(w, err)
		return
	}
	if !success {
		c.printError(w, errors.New("couldn't delete news"))
	}
}

func (c *NewsController) patchNews(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		http.Error(w, "missing param: id", http.StatusBadRequest)
		return
	}

	var update persistence.News
	err := json.NewDecoder(req.Body).Decode(&update)
	if err != nil {
		c.printError(w, err)
		return
	}

	news, err := c.s.PatchNews(req.Context(), id, update)
	if err != nil {
		c.printError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(news)
	if err != nil {
		c.printError(w, err)
	}
}
