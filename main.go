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
	"time"
)

const lookupTimeout = 10 * time.Second

func main() {
	var concurrency int
	var verbose, vVerbose, dnsTCP, all bool
	var customDNS, dnsPort string

	flag.IntVar(&concurrency, "c", 20, "set the concurrency level")
	flag.BoolVar(&verbose, "v", false, "Show hostname with the corresponding IP")
	flag.BoolVar(&vVerbose, "vv", false, "Show any errors and relevant info (on stderr)")
	flag.BoolVar(&all, "a", false, "Show every IP address a host resolves to, not just the first")
	flag.StringVar(&customDNS, "dns", "", "Custom DNS server to use for resolution")
	flag.StringVar(&dnsPort, "port", "53", "DNS server port")
	flag.BoolVar(&dnsTCP, "dns-tcp", false, "Use DNS over TCP instead of the default UDP. Useful for SOCKS proxy environments where UDP is not supported.")
	flag.Parse()

	if concurrency < 1 {
		fmt.Fprintf(os.Stderr, "[!] -c must be at least 1 (got %d)\n", concurrency)
		os.Exit(2)
	}

	resolver := net.DefaultResolver
	if customDNS != "" {
		// JoinHostPort brackets IPv6 server addresses; plain concatenation
		// produced "::1:53", which does not parse.
		server := net.JoinHostPort(customDNS, dnsPort)
		network := "udp"
		if dnsTCP {
			network = "tcp"
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second}
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
				// Honour the lookup's context so a stalled server cannot
				// wedge a worker; net.Dial had no deadline at all.
				return dialer.DialContext(ctx, network, server)
			},
		}
	}

	hosts := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for host := range hosts {
				ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
				addrs, err := resolver.LookupIPAddr(ctx, host)
				cancel()
				if err != nil || len(addrs) == 0 {
					if vVerbose {
						fmt.Fprintf(os.Stderr, "%s could not be found\n", host)
					}
					continue
				}
				if !all {
					addrs = addrs[:1]
				}
				for _, a := range addrs {
					if verbose {
						fmt.Printf("%s,%s\n", host, a.IP)
					} else {
						fmt.Println(a.IP)
					}
				}
			}
		}()
	}

	var input io.Reader = os.Stdin
	if arg := flag.Arg(0); arg != "" {
		input = strings.NewReader(arg)
	}

	sc := bufio.NewScanner(input)
	seen := make(map[string]bool)
	for sc.Scan() {
		host, ok := hostFromLine(sc.Text())
		if !ok {
			if raw := strings.TrimSpace(sc.Text()); vVerbose && raw != "" && !strings.HasPrefix(raw, "#") {
				fmt.Fprintf(os.Stderr, "%s could not be parsed\n", raw)
			}
			continue
		}
		if seen[host] {
			if vVerbose {
				fmt.Fprintf(os.Stderr, "Already seen %s\n", host)
			}
			continue
		}
		seen[host] = true
		hosts <- host
	}

	close(hosts)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] failed to read input: %s\n", err)
	}

	wg.Wait()
}

// hostFromLine extracts the hostname to resolve from one input line: a bare
// hostname, a host:port pair, or a URL. Blank lines and comments are skipped.
// URLs are detected by "://" rather than an "http" prefix, which used to
// swallow real hostnames such as httpbin.org.
func hostFromLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", false
	}
	if strings.Contains(line, "://") {
		u, err := url.Parse(line)
		if err != nil || u.Hostname() == "" {
			return "", false
		}
		return u.Hostname(), true
	}
	if h, _, err := net.SplitHostPort(line); err == nil && h != "" {
		return h, true
	}
	return strings.TrimSuffix(line, "."), true
}
