package server

import "github.com/iivkis/pos-ninja-backend/internal/handler"

type Server struct {
	httphandler handler.HttpHandler
}

func NewServer(httphandler handler.HttpHandler) Server {
	return Server{
		httphandler: httphandler,
	}
}

func (s *Server) Listen(port string) error {
	return s.httphandler.Init().Run()
}
