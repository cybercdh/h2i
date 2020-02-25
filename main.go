package main
 
import (
	"bufio"
	"flag"
	"io"
	"net"
	"net/url"
	"os"
	"fmt"
	"strings"
	"sync"
)
 
func main() {

	// parse flags
	var concurrency int
	flag.IntVar(&concurrency, "c", 20, "set the concurrency level")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Show hostname with the corresponding IP")

	var v_verbose bool
	flag.BoolVar(&v_verbose, "vv", false, "Show any errors and relevant info")

	flag.Parse()

	// make a hosts channel
	hosts := make(chan string)

	// spin up a bunch of workers
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			for host := range hosts {

				// perform the host->IP lookup
				addr,err := net.LookupIP(host)
				if err != nil {
					if v_verbose {
						fmt.Printf("%s could not be found\n", host)	
					}
				} else {	
					if verbose {
						fmt.Printf("%s %s \n", host, addr[0])	
					} else {
						fmt.Println(addr[0])	
					}
				}	
			}	
			wg.Done()
		}()
	}

	// take hostnames from stdin
	var input_hosts io.Reader
	input_hosts = os.Stdin

	// or take hostnames piped to the code
	arg_hosts := flag.Arg(0)
	if arg_hosts != "" {
		input_hosts = strings.NewReader(arg_hosts)
	}

	sc := bufio.NewScanner(input_hosts)

	// keep track of anything we've seen, to avoid dupes
	seen := make(map[string]bool)

	// loop over the input
	for sc.Scan() {

		tmp_host := sc.Text()

		// accept URLs too, parse them to extract the hostname
		if strings.HasPrefix(tmp_host, "http") {
			u, err := url.Parse(tmp_host)
			if err != nil {
				if v_verbose {
					fmt.Printf("%s could not be found\n", tmp_host)	
				}
			} else {
				tmp_host = u.Hostname()
			}
		}

		// if we've already seen, report and skip
		if _, ok := seen[tmp_host]; ok {
			if v_verbose {
				fmt.Printf("Already seen %s\n", tmp_host)	
			}
			continue
		}
		
		// add to seen
		seen[tmp_host] = true

		// add host to the channel
		hosts <- tmp_host
	}

	close(hosts)

	// check there were no errors reading stdin (unlikely)
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[!]	failed to read input: %s\n", err)
	}

	// wait until all the workers have finished
	wg.Wait()


}