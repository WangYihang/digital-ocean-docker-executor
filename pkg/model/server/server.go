package server

type Server interface {
	IPv4() string
	IPv6() string
	Tags() []string
}
