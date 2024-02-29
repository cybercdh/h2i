package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
)

func main() {
	var concurrency int
	var verbose, v_verbose bool
	var customDNS string
	var dnsPort string

	flag.IntVar(&concurrency, "c", 20, "set the concurrency level")
	flag.BoolVar(&verbose, "v", false, "Show hostname with the corresponding IP")
	flag.BoolVar(&v_verbose, "vv", false, "Show any errors and relevant info")
	flag.StringVar(&customDNS, "dns", "", "Custom DNS server to use for resolution")
	flag.StringVar(&dnsPort, "port", "53", "DNS server port")
	flag.Parse()

	hosts := make(chan string)
	var wg sync.WaitGroup
	resolver := &net.Resolver{}

	// Use custom DNS server if specified
	if customDNS != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial("udp", customDNS+":"+dnsPort)
			},
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			for host := range hosts {
				// Use custom resolver for lookup
				addr, err := resolver.LookupIPAddr(context.Background(), host)
				if err != nil {
					if v_verbose {
						fmt.Printf("%s could not be found\n", host)
					}
				} else {
					if verbose {
						fmt.Printf("%s,%s \n", host, addr[0].IP.String())
					} else {
						fmt.Println(addr[0].IP.String())
					}
				}
			}
			wg.Done()
		}()
	}

	var input_hosts io.Reader = os.Stdin
	arg_hosts := flag.Arg(0)
	if arg_hosts != "" {
		input_hosts = strings.NewReader(arg_hosts)
	}

	sc := bufio.NewScanner(input_hosts)
	seen := make(map[string]bool)

	for sc.Scan() {
		tmp_host := sc.Text()
		if strings.HasPrefix(tmp_host, "http") {
			u, err := url.Parse(tmp_host)
			if err != nil {
				if v_verbose {
					fmt.Printf("%s could not be parsed\n", tmp_host)
				}
				continue
			}
			tmp_host = u.Hostname()
		}

		if _, ok := seen[tmp_host]; ok {
			if v_verbose {
				fmt.Printf("Already seen %s\n", tmp_host)
			}
			continue
		}

		seen[tmp_host] = true
		hosts <- tmp_host
	}

	close(hosts)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] failed to read input: %s\n", err)
	}

	wg.Wait()
}
