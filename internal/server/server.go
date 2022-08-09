package server

type Runner interface {
	Run(addr ...string) error
}

type Server struct {
	errors chan error
	runner Runner
}

func NewServer(runner Runner) (*Server, chan error) {
	server := Server{
		errors: make(chan error),
		runner: runner,
	}

	return &server, server.errors
}

func (s *Server) Run(host, port string) {
	go func() {
		s.errors <- s.runner.Run(host + ":" + port)
	}()
}
