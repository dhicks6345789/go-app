package auth

import (
	"net/http"
	"os"
	"strings"
)

type UserInfo struct {
	Username string `json:"username"`
	AuthType string `json:"auth_type"`
	Mode     string `json:"mode"`
}

// Common headers used by reverse proxies (Traefik, Pangolin, Cloudflare Tunnel, Authelia, etc.)
var proxyHeaders = []string{
	"X-Forwarded-User",
	"Remote-User",
	"X-User",
	"CF-Access-Authenticated-User-Email",
	"X-Auth-Request-User",
	"Pangolin-User",
}

// GetUser resolves user identity and operation mode based on environmental context or request headers.
func GetUser(r *http.Request, isServerMode bool) UserInfo {
	if isServerMode {
		for _, header := range proxyHeaders {
			if val := r.Header.Get(header); val != "" {
				return UserInfo{
					Username: strings.TrimSpace(val),
					AuthType: "proxy_header (" + header + ")",
					Mode:     "server",
				}
			}
		}
		return UserInfo{
			Username: "anonymous",
			AuthType: "none",
			Mode:     "server",
		}
	}

	// Desktop / Local mode: read local user environment variable
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = os.Getenv("LOGNAME")
	}
	if username == "" {
		username = "localuser"
	}

	return UserInfo{
		Username: username,
		AuthType: "local_env",
		Mode:     "desktop",
	}
}
