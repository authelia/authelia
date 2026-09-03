package suites

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

// HostEntry represents an entry in /etc/hosts.
type HostEntry struct {
	Domain string
	IP     string
}

// HostEntries returns every domain the suites use and the address it has on the suite network. The addresses follow
// SuiteSubnet, so a slotted shell and an unslotted one describe different networks with the same names.
func HostEntries() []HostEntry {
	backend := SuiteAddress(50)
	portal := SuiteAddress(100)
	ssh := SuiteAddress(130)

	entries := []HostEntry{
		// For unit tests.
		{Domain: "local.example.com", IP: "127.0.0.1"},

		// For authelia backend.
		{Domain: "authelia.example.com", IP: backend},

		// For common tests.
		{Domain: "login.example.com", IP: portal},
		{Domain: "admin.example.com", IP: portal},
		{Domain: "singlefactor.example.com", IP: portal},
		{Domain: "deny.example.com", IP: portal},
		{Domain: "dev.example.com", IP: portal},
		{Domain: "home.example.com", IP: portal},
		{Domain: "mx1.mail.example.com", IP: portal},
		{Domain: "mx2.mail.example.com", IP: portal},
		{Domain: "public.example.com", IP: portal},
		{Domain: "secure.example.com", IP: portal},
		{Domain: "mail.example.com", IP: portal},
		{Domain: "duo.example.com", IP: portal},

		// For HAProxy suite.
		{Domain: "haproxy.example.com", IP: portal},

		// Kubernetes dashboard.
		{Domain: "kubernetes.example.com", IP: portal},

		// OIDC tester app.
		{Domain: "oidc.example.com", IP: portal},
		{Domain: "oidc-public.example.com", IP: portal},

		// The external OpenID Connect 1.0 Provider of the OpenIDConnectRelyingParty suite.
		{Domain: "auth-upstream.example.com", IP: portal},

		// For Traefik suite.
		{Domain: "traefik.example.com", IP: portal},

		// For testing network ACLs.
		{Domain: "proxy-client1.example.com", IP: SuiteAddress(201)},
		{Domain: "proxy-client2.example.com", IP: SuiteAddress(202)},
		{Domain: "proxy-client3.example.com", IP: SuiteAddress(203)},

		// Redis Replicas.
		{Domain: "redis-node-0.example.com", IP: SuiteAddress(110)},
		{Domain: "redis-node-1.example.com", IP: SuiteAddress(111)},
		{Domain: "redis-node-2.example.com", IP: SuiteAddress(112)},

		// Redis Sentinel Replicas.
		{Domain: "redis-sentinel-0.example.com", IP: SuiteAddress(120)},
		{Domain: "redis-sentinel-1.example.com", IP: SuiteAddress(121)},
		{Domain: "redis-sentinel-2.example.com", IP: SuiteAddress(122)},

		// For PAM suite.
		{Domain: "ssh.example.com", IP: ssh},
	}

	return slices.Concat(entries, hostEntriesCookieDomains(portal))
}

func hostEntriesCookieDomains(ip string) []HostEntry {
	domains := []string{"example2.com", "example3.com"}
	subdomains := []string{"login", "admin", "singlefactor", "dev", "home", "mx1.mail", "mx2.mail", "public", "secure", "mail", "duo"}

	entries := make([]HostEntry, 0, len(domains)*len(subdomains))

	for _, domain := range domains {
		for _, subdomain := range subdomains {
			entries = append(entries, HostEntry{Domain: subdomain + "." + domain, IP: ip})
		}
	}

	return entries
}

func hostAddresses() map[string]string {
	entries := HostEntries()

	addresses := make(map[string]string, len(entries))

	for _, entry := range entries {
		addresses[entry.Domain] = entry.IP
	}

	return addresses
}

// HostResolverRules renders HostEntries as the value of Chrome's --host-resolver-rules. Chrome resolves the suite
// domains from this rather than from /etc/hosts, which is what lets concurrent runs on one machine each reach their
// own network while a single /etc/hosts describes only one of them.
func HostResolverRules() string {
	entries := HostEntries()

	rules := make([]string, len(entries))

	for i, entry := range entries {
		rules[i] = fmt.Sprintf("MAP %s %s", entry.Domain, entry.IP)
	}

	return strings.Join(rules, ",")
}

// ResolveAddr returns addr with the suite address substituted for its host when that host is a suite domain, which is
// the same resolution Chrome performs through HostResolverRules. Anything else is returned unchanged.
func ResolveAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	ip, ok := hostAddresses()[host]
	if !ok {
		return addr
	}

	return net.JoinHostPort(ip, port)
}

// DialContext dials addr through ResolveAddr. Resolving at the dial layer rather than through /etc/hosts is what lets
// concurrent runs on one machine each reach their own network, and it leaves the SNI name, the Host header and the
// cookie domain as the name the caller asked for.
func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}

	return dialer.DialContext(ctx, network, ResolveAddr(addr))
}
