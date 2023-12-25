package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/claudemuller/scanner/internal/pkg/scanner"
)

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

	ports, err := scanner.ParsePorts(portStr)
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
	workQ := make(chan scanner.Target, threads)

	// Start up workers, each consuming from the workQ(ueue)
	for i := 0; i < cap(workQ); i++ {
		go scanner.Worker(workQ, results)
	}

	// Fill the workQ(ueue) with ports to scan
	go func() {
		for _, port := range ports {
			workQ <- scanner.Target{
				Host:  host,
				Port:  port,
				Proto: proto,
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

func printHeader() error {
	data, err := os.ReadFile("banner.txt")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", data)

	return nil
}
