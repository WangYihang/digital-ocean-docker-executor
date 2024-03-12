package server

type Server interface {
	Name() string
	ID() string
	IPv4() string
	IPv6() string
	Tags() []string
}
