package server

import (
	"net/http"
)

type Server struct {
}

func myHandler(w http.ResponseWriter, r *http.Request) {}

func New() *Server {
	return &Server{}
}

func (s *Server) Run() {

}
