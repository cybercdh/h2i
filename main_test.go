package main

import "testing"

func TestHostFromLine(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"example.com", "example.com", true},
		{"  example.com.  ", "example.com", true},
		{"httpbin.org", "httpbin.org", true},
		{"https://httpbin.org/get?x=1", "httpbin.org", true},
		{"http://user:pw@example.com:8443/", "example.com", true},
		{"example.com:443", "example.com", true},
		{"[2606:4700::1111]:53", "2606:4700::1111", true},
		{"", "", false},
		{"# comment", "", false},
		{"http://", "", false},
	}
	for _, c := range cases {
		got, ok := hostFromLine(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("hostFromLine(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}
