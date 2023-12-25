package scanner

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Target struct {
	Host  string
	Port  int
	Proto string
}

func Worker(tangos chan Target, results chan int) {
	for tango := range tangos {
		p := strconv.Itoa(tango.Port)

		conn, err := net.Dial(tango.Proto, tango.Host+":"+p)
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
		results <- tango.Port
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

func ParsePorts(portStr string) ([]int, error) {
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
