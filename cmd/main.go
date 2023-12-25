package main

import (
	"flag"
	"log"
	"net"
)

func main() {
	var host string
	var port string
	var proto string

	flag.StringVar(&host, "host", "", "the host to scan")
	flag.StringVar(&port, "port", "", "the port to scan")
	flag.StringVar(&proto, "proto", "tcp", "the proto to use for the scan")

	flag.Parse()

	if host == "" || port == "" {
		flag.Usage()
		return
	}

	_, err := net.Dial(proto, host+":"+port)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	log.Print("Connected successfully.\n")
}
