package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	var host string
	var portStr string
	var proto string

	flag.StringVar(&host, "host", "", "the host to scan")
	flag.StringVar(&portStr, "ports", "", "the ports to scan")
	flag.StringVar(&proto, "proto", "tcp", "the proto to use for the scan")

	flag.Parse()

	if host == "" {
		flag.Usage()
		return
	}

	ports, err := parsePorts(portStr)
	if err != nil {
		log.Fatalf("error parsing ports: %v\n", err)
	}

	log.Printf("[+] Beginning scan on: %s\n", host)

	for _, port := range ports {
		p := strconv.Itoa(port)

		conn, err := net.Dial(proto, host+":"+p)
		if err != nil {
			// TODO: clean this up
			if ne, ok := err.(*net.OpError); ok {
				if e, ok := (ne.Err).(*os.SyscallError); ok {
					if e.Syscall == "connect" && e.Err.Error() == "connection refused" {
						log.Printf("[-] Closed port: %s\n", p)
						continue
					}
				}
			}

			log.Printf("error creating connection: %T %v\n", err, err)
			continue
		}

		conn.Close()

		log.Printf("[+] Open port: %s\n", p)
	}
}

func parsePorts(portStr string) ([]int, error) {
	switch {
	case strings.Contains(portStr, "-"):
		pRange := strings.Split(portStr, "-")

		min, err := strconv.Atoi(pRange[0])
		if err != nil {
			return nil, fmt.Errorf("error determining min port range value: %v", err)
		}

		max, err := strconv.Atoi(pRange[1])
		if err != nil {
			return nil, fmt.Errorf("error determining max port range value: %v", err)
		}

		ports := make([]int, 0, max-min+1)

		for i := min; i <= max; i++ {
			ports = append(ports, i)
		}

		return ports, nil

	case strings.Contains(portStr, ","):
		portStrs := strings.Split(portStr, ",")
		ports := make([]int, 0, len(portStrs))

		for _, p := range portStrs {
			port, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("error parsing port string value: %v", err)
			}
			ports = append(ports, port)
		}

		return ports, nil

	default:
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}

		if p == 0 {
			return nil, errors.New("invalid port")
		}

		return []int{p}, nil
	}
}
