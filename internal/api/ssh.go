package api

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var pathMap = map[string]string{
	"interface/ethernet":     "/interface ethernet",
	"interface/bridge":       "/interface bridge",
	"interface/bridge/port":  "/interface bridge port",
	"interface/bridge/vlan":  "/interface bridge vlan",
	"interface/vlan":         "/interface vlan",
	"interface/pppoe-client": "/interface pppoe-client",
	"ip/address":             "/ip address",
	"ip/route":               "/ip route",
	"ip/pool":                "/ip pool",
	"ip/dns":                 "/ip dns",
	"ip/firewall/filter":       "/ip firewall filter",
	"ip/firewall/mangle":       "/ip firewall mangle",
	"ip/firewall/nat":          "/ip firewall nat",
	"ip/firewall/address-list": "/ip firewall address-list",
	"ip/service":               "/ip service",
	"ip/dhcp-server":           "/ip dhcp-server",
	"ip/dhcp-server/network":   "/ip dhcp-server network",
	"ip/dhcp-server/lease":     "/ip dhcp-server lease",
	"ip/ipsec/profile":         "/ip ipsec profile",
	"ip/ipsec/proposal":        "/ip ipsec proposal",
	"ip/ipsec/peer":            "/ip ipsec peer",
	"ip/ipsec/identity":        "/ip ipsec identity",
	"ip/ipsec/policy":          "/ip ipsec policy",
	"routing/table":            "/routing table",
	"system/identity":          "/system identity",
	"system/resource":          "/system resource",
	"system/reboot":            "/system reboot",
	"system/backup":            "/system backup",
	"system/script":            "/system script",
	"system/scheduler":         "/system scheduler",
	"ip/cloud":                 "/ip cloud",
	"queue/simple":             "/queue simple",
}

func restToCLI(path string) string {
	clean := strings.Trim(path, "/")
	if cli, ok := pathMap[clean]; ok {
		return cli
	}
	return "/" + strings.ReplaceAll(clean, "/", " ")
}

// SSHClient implements MikrotikClient using SSH CLI execution.
type SSHClient struct {
	client    *ssh.Client
	connected bool
	identity  string
}

// NewSSHClient creates an uninitialized SSH client.
func NewSSHClient() *SSHClient {
	return &SSHClient{}
}

func (c *SSHClient) Connect(host, user, password string, port int, useSSL bool) (string, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		c.connected = false
		return "", fmt.Errorf("SSH kết nối đến %s thất bại: %w", addr, err)
	}

	c.client = client
	c.connected = true

	// Get identity
	out, err := c.ExecRaw("/system identity print")
	if err == nil {
		re := regexp.MustCompile(`name:\s*(\S+)`)
		match := re.FindStringSubmatch(out)
		if len(match) > 1 {
			c.identity = match[1]
		}
	}
	if c.identity == "" {
		c.identity = host
	}

	return c.identity, nil
}

func (c *SSHClient) Disconnect() error {
	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}
	c.connected = false
	c.identity = ""
	return nil
}

func (c *SSHClient) IsConnected() bool {
	return c.connected
}

func (c *SSHClient) GetIdentity() string {
	return c.identity
}

func (c *SSHClient) ExecRaw(cmd string) (string, error) {
	if !c.connected || c.client == nil {
		return "", fmt.Errorf("SSH chưa kết nối")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	runErr := session.Run(cmd)
	outStr := stdout.String()
	errStr := stderr.String()

	combined := strings.TrimSpace(outStr + "\n" + errStr)
	lower := strings.ToLower(combined)

	if runErr != nil || strings.Contains(lower, "expected command name") ||
		strings.Contains(lower, "bad command name") ||
		strings.Contains(lower, "syntax error") ||
		(strings.Contains(lower, "failure:") && !strings.Contains(lower, "already have")) {
		return outStr, fmt.Errorf("SSH lỗi lệnh: %s", combined)
	}
	return outStr, nil
}

func parseGetOutput(output string) []map[string]interface{} {
	var list []map[string]interface{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		item := make(map[string]interface{})
		tokens := strings.Split(line, ";")
		for _, token := range tokens {
			idx := strings.Index(token, "=")
			if idx > 0 {
				k := token[:idx]
				v := token[idx+1:]
				item[k] = v
			}
		}
		if len(item) > 0 {
			list = append(list, item)
		}
	}
	return list
}

func parseTerse(output string) []map[string]interface{} {
	var results []map[string]interface{}
	lines := strings.Split(output, "\n")
	reRow := regexp.MustCompile(`^(\d+)\s+(.*)`)
	reFlag := regexp.MustCompile(`^[A-Za-z]{1,3}\s+`)
	reKV := regexp.MustCompile(`(\S+?)=(?:"([^"]*?)"|(\S+))`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := reRow.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		idx := m[1]
		rest := m[2]
		rest = reFlag.ReplaceAllString(rest, "")

		item := map[string]interface{}{
			".id": "*" + idx,
		}
		matches := reKV.FindAllStringSubmatch(rest, -1)
		for _, kv := range matches {
			k := kv[1]
			v := kv[2]
			if v == "" && len(kv) > 3 {
				v = kv[3]
			}
			item[k] = v
		}
		results = append(results, item)
	}
	return results
}

func (c *SSHClient) Get(path string) ([]map[string]interface{}, error) {
	cli := restToCLI(path)
	clean := strings.Trim(path, "/")

	// Singletons: ip/dns, system/resource, system/identity
	if clean == "ip/dns" || clean == "system/resource" || clean == "system/identity" {
		out, err := c.ExecRaw(cli + " print")
		if err != nil {
			return nil, err
		}
		item := map[string]interface{}{".id": ""}
		re := regexp.MustCompile(`(?m)^\s*([\w\-]+):\s+(.*)$`)
		matches := re.FindAllStringSubmatch(out, -1)
		for _, m := range matches {
			if len(m) > 2 {
				item[m[1]] = strings.TrimSpace(m[2])
			}
		}
		if len(item) > 1 {
			return []map[string]interface{}{item}, nil
		}
		return []map[string]interface{}{}, nil
	}

	// File listing: use print terse to avoid reading full file contents
	if clean == "file" {
		out, err := c.ExecRaw(cli + " print terse")
		if err != nil {
			return nil, err
		}
		return parseTerse(out), nil
	}

	// All other tables: use :foreach get to obtain exact internal .id and full properties
	cmd := fmt.Sprintf(`:foreach i in=[%s find] do={ :put [%s get $i] }`, cli, cli)
	out, err := c.ExecRaw(cmd)
	if err != nil {
		outFallback, errFallback := c.ExecRaw(cli + " print terse")
		if errFallback == nil {
			return parseTerse(outFallback), nil
		}
		return nil, err
	}
	return parseGetOutput(out), nil
}

func (c *SSHClient) GetOne(path string) (map[string]interface{}, error) {
	items, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items[0], nil
	}
	return map[string]interface{}{}, nil
}

func buildArgs(data map[string]interface{}) string {
	var parts []string
	for k, v := range data {
		if strings.HasPrefix(k, ".") {
			continue
		}
		switch val := v.(type) {
		case bool:
			if val {
				parts = append(parts, fmt.Sprintf("%s=yes", k))
			} else {
				parts = append(parts, fmt.Sprintf("%s=no", k))
			}
		case int, int64, int32, float64, float32:
			parts = append(parts, fmt.Sprintf("%s=%v", k, val))
		default:
			valStr := fmt.Sprintf("%v", val)
			if valStr == "yes" || valStr == "no" {
				parts = append(parts, fmt.Sprintf("%s=%s", k, valStr))
			} else {
				// Always quote string values and escape special characters (quotes, backslashes, newlines)
				escaped := strings.ReplaceAll(valStr, `\`, `\\`)
				escaped = strings.ReplaceAll(escaped, `"`, `\"`)
				escaped = strings.ReplaceAll(escaped, "\r\n", `\n`)
				escaped = strings.ReplaceAll(escaped, "\n", `\n`)
				escaped = strings.ReplaceAll(escaped, "\r", `\n`)
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, escaped))
			}
		}
	}
	return strings.Join(parts, " ")
}

func (c *SSHClient) Add(path string, data map[string]interface{}) (map[string]interface{}, error) {
	cli := restToCLI(path)
	args := buildArgs(data)
	cmd := fmt.Sprintf("%s add %s", cli, args)
	_, err := c.ExecRaw(cmd)
	return map[string]interface{}{}, err
}

func (c *SSHClient) Set(path string, itemID string, data map[string]interface{}) error {
	cli := restToCLI(path)
	args := buildArgs(data)
	var cmd string
	if itemID != "" {
		cmd = fmt.Sprintf("%s set %s %s", cli, itemID, args)
	} else {
		cmd = fmt.Sprintf("%s set %s", cli, args)
	}
	_, err := c.ExecRaw(cmd)
	return err
}

func (c *SSHClient) Remove(path string, itemID string) error {
	cli := restToCLI(path)
	cmd := fmt.Sprintf("%s remove %s", cli, itemID)
	_, err := c.ExecRaw(cmd)
	return err
}

func (c *SSHClient) Run(path string, data map[string]interface{}) (map[string]interface{}, error) {
	cli := restToCLI(path)
	args := buildArgs(data)
	cmd := strings.TrimSpace(fmt.Sprintf("%s %s", cli, args))
	_, err := c.ExecRaw(cmd)
	return map[string]interface{}{}, err
}

func (c *SSHClient) RemoveByComment(path, comment string) (int, error) {
	cli := restToCLI(path)
	cmd := fmt.Sprintf(`%s remove [find where comment~"%s"]`, cli, comment)
	_, err := c.ExecRaw(cmd)
	if err == nil {
		return 1, nil
	}
	return RemoveByCommentHelper(c, path, comment)
}

func (c *SSHClient) RemoveAllMatching(path string, filters map[string]string) (int, error) {
	return RemoveAllMatchingHelper(c, path, filters)
}
