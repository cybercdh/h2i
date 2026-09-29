# h2i

Converts a list of hostnames or URLs to their corresponding IP address


## Install

If you have Go installed and configured (i.e. with `$GOPATH/bin` in your `$PATH`):

```
go install github.com/cybercdh/h2i@latest
```

## Usage

```
$ h2i <url>
```
or 
```
$ cat <file> | h2i
```

By default, the code will simply print the list of IP's to the console. For more details, use the -v flag, per below.

### Options

```
Usage of h2i:
  -a  Show every IP address a host resolves to, not just the first
  -c int
      set the concurrency level (default 20)
  -dns string
      Custom DNS server to use for resolution
  -dns-tcp
      Use DNS over TCP instead of the default UDP. Useful for SOCKS proxy environments where UDP is not supported.
  -port string
      DNS server port (default "53")
  -v  Show hostname with the corresponding IP
  -vv
      Show any errors and relevant info (on stderr)
```

Input lines may be bare hostnames, `host:port` pairs, or full URLs. Diagnostics from `-vv` go to stderr, so stdout stays a clean list when piped.
