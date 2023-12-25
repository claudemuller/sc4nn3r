package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

type target struct {
	host  string
	port  int
	proto string
}

func main() {
	var host string
	var portStr string
	var proto string
	var threads int

	flag.StringVar(&host, "host", "", "the host to scan")
	flag.StringVar(&portStr, "ports", "1-1024", "the port(s) or range of ports to scan")
	flag.StringVar(&proto, "proto", "tcp", "the proto to use for the scan")
	flag.IntVar(&threads, "threads", 100, "the number of \"threads\" i.e. goroutines to use")

	flag.Parse()

	if host == "" {
		flag.Usage()
		return
	}

	ports, err := parsePorts(portStr)
	if err != nil {
		log.Fatalf("error parsing ports: %v\n", err)
	}

	if err := printHeader(); err != nil {
		log.Fatalf("error printing banner: %v", err)
	}
	log.Printf("[+] Beginning scan on: %s\n", host)
	log.Printf("[+] Scanning %d ports\n", len(ports))
	log.Printf("[+] Using %d goroutines\n", threads)
	log.Println("[+] -------------------------------------------------------------------")

	results := make(chan int)
	workQ := make(chan target, threads)

	// Start up workers, each consuming from the workQ(ueue)
	for i := 0; i < cap(workQ); i++ {
		go worker(workQ, results)
	}

	// Fill the workQ(ueue) with ports to scan
	go func() {
		for _, port := range ports {
			workQ <- target{
				host:  host,
				port:  port,
				proto: proto,
			}
		}
	}()

	var openPorts []int

	for i := 0; i < len(ports); i++ {
		p := <-results
		if p != 0 {
			openPorts = append(openPorts, p)
		}
	}

	// Shutdown workers
	close(workQ)
	close(results)

	sort.Ints(openPorts)
	for _, p := range openPorts {
		log.Printf("[+] Open port: %d\n", p)
	}

	log.Println("[+] -------------------------------------------------------------------")
	log.Printf("[+] Found %d open ports\n", len(openPorts))
}

func worker(tangos chan target, results chan int) {
	for tango := range tangos {
		p := strconv.Itoa(tango.port)

		conn, err := net.Dial(tango.proto, tango.host+":"+p)
		if err != nil {
			if handlePortClosed(err, p) {
				results <- 0
				continue
			}

			// TODO: Send errors on error channel
			log.Printf("error creating connection: %T %v\n", err, err)
			results <- 0
			continue
		}

		conn.Close()
		results <- tango.port
	}
}

func handlePortClosed(err error, port string) bool {
	if ne, ok := err.(*net.OpError); ok {
		if e, ok := (ne.Err).(*os.SyscallError); ok {
			if e.Syscall == "connect" && e.Err.Error() == "connection refused" {
				return true
			}
		}
	}

	return false
}

func parsePorts(portStr string) ([]int, error) {
	switch {
	case strings.Contains(portStr, "-"):
		// Port format: 80-100
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
		// Port format: 25,80,443
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

func printHeader() error {
	data, err := os.ReadFile("banner.txt")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", data)

	return nil
}
