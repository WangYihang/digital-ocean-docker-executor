package digitalocean

import (
	"log/slog"

	"github.com/digitalocean/godo"
)

type Server struct {
	droplet *godo.Droplet
}

func NewServer(droplet *godo.Droplet) *Server {
	return &Server{
		droplet: droplet,
	}
}

func (s *Server) IPv4() string {
	ip, err := s.droplet.PublicIPv4()
	if err != nil {
		slog.Error("error occured while getting public ipv4", slog.String("error", err.Error()))
	}
	return ip
}

func (s *Server) IPv6() string {
	ip, err := s.droplet.PublicIPv6()
	if err != nil {
		slog.Error("error occured while getting public ipv6", slog.String("error", err.Error()))
	}
	return ip
}

func (s *Server) Tags() []string {
	return s.droplet.Tags
}
