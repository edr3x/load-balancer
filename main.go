package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Servers interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func serverAddress(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)

	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port             string
	roundRobbinCount int
	servers          []Servers
}

func NewLoadBalancer(port string, servers []Servers) *LoadBalancer {
	return &LoadBalancer{
		port:             port,
		roundRobbinCount: 0,
		servers:          servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v \n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string { return s.addr }

func (lb *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb *LoadBalancer) getNextAvailabeServer() Servers {
	server := lb.servers[lb.roundRobbinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobbinCount++
		server = lb.servers[lb.roundRobbinCount%len(lb.servers)]
	}
	lb.roundRobbinCount++
	return server
}

func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailabeServer()
	fmt.Printf("forwarding to address %q\n", targetServer.Address())
	targetServer.Serve(rw, req)
}

func main() {
	servers := []Servers{
		serverAddress("https://www.google.com"),
		serverAddress("https://otakujin.net"),
		serverAddress("https://duckduckgo.com"),
	}
	lb := NewLoadBalancer("8080", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {

		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Server started at localhost:%s \n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
