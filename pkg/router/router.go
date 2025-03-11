package router

import (
	"db_explorer/pkg/router/routertrie"
	"fmt"
	"net/http"
)

type MuxRouter struct {
	mux *http.ServeMux
	t   *routertrie.Trie
}

func NewMuxRouter() *MuxRouter {
	router := MuxRouter{
		mux: http.NewServeMux(),
		t:   routertrie.NewTrie(),
	}

	return &router
}

func (router *MuxRouter) Route(method string, path string, handler http.HandlerFunc) {
	router.t.Put(method, path, handler)
}

func (router *MuxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request %v \n", r.URL.Path)

	h, req := router.t.FindHandler(r)

	if h == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	h(w, req)
}
