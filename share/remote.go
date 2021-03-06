package chshare

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// short-hand conversions
//   foobar.com:3000 ->
//		local 127.0.0.1:3000
//		remote foobar.com:3000
//   3000:google.com:80 ->
//		local 127.0.0.1:3000
//		remote google.com:80
//   192.168.0.1:3000:google.com:80 ->
//		local 192.168.0.1:3000
//		remote google.com:80
//   127.0.0.1:80:127.0.0.1:80@B ->
//		local 127.0.0.1:80
//		remote 127.0.0.1:80
// 		over client named B
//	80@B ->
//		local 127.0.0.1:80
//		remote 127.0.0.1:80
//		over client named B

type Remote struct {
	LocalHost, LocalPort, RemoteHost, RemotePort string
	Socks                                        bool
	Proxy                                        string
}

func DecodeRemote(s string) (*Remote, error) {
	parts := strings.Split(s, ":")
	if len(parts) <= 0 || len(parts) >= 5 {
		return nil, errors.New("Invalid remote")
	}
	r := &Remote{}
	//TODO fix up hacky decode
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		//last part "socks"?
		if i == len(parts)-1 && p == "socks" {
			r.Socks = true
			continue
		}

		if isProxy(p) {
			proxyParts := strings.Split(p, "@")
			p = proxyParts[0]
			r.Proxy = proxyParts[1]
		}

		if isPort(p) {
			if !r.Socks && r.RemotePort == "" {
				r.RemotePort = p
				r.LocalPort = p
			} else {
				r.LocalPort = p
			}
			continue
		}
		if !r.Socks && (r.RemotePort == "" && r.LocalPort == "") {
			return nil, errors.New("Missing ports")
		}
		if !isHost(p) {
			return nil, errors.New("Invalid host")
		}
		if !r.Socks && r.RemoteHost == "" {
			r.RemoteHost = p
		} else {
			r.LocalHost = p
		}
	}
	if r.LocalHost == "" {
		if r.Socks {
			r.LocalHost = "127.0.0.1"
		} else {
			r.LocalHost = "0.0.0.0"
		}
	}
	if r.LocalPort == "" && r.Socks {
		r.LocalPort = "1080"
	}
	if !r.Socks && r.RemoteHost == "" {
		r.RemoteHost = "0.0.0.0"
	}
	return r, nil
}

//check if remote contains proxy settings
func isProxy(s string) bool {
	if strings.Contains(s, "@") {
		return true
	}

	return false
}

var isPortRegExp = regexp.MustCompile(`^\d+$`)

func isPort(s string) bool {
	if !isPortRegExp.MatchString(s) {
		return false
	}
	return true
}

var isHTTP = regexp.MustCompile(`^http?:\/\/`)

func isHost(s string) bool {
	_, err := url.Parse(s)
	if err != nil {
		return false
	}
	return true
}

//implement Stringer
func (r *Remote) String() string {
	return r.LocalHost + ":" + r.LocalPort + "=>" + r.Remote()
}

func (r *Remote) Remote() string {
	if r.Socks {
		return "socks"
	}

	if r.Proxy != "" {
		return r.RemoteHost + ":" + r.RemotePort + "@" + r.Proxy
	}

	return r.RemoteHost + ":" + r.RemotePort
}
