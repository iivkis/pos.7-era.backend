package server

type Runner interface {
	Run(addr ...string) error
}

type Server struct {
	runner Runner
}

func NewServer(runner Runner) Server {
	return Server{
		runner: runner,
	}
}

func (s *Server) Run(host, port string) error {
	return s.runner.Run(host + ":" + port)
}
