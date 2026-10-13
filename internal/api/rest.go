package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RESTClient implements MikrotikClient using RouterOS 7 REST API.
type RESTClient struct {
	baseURL    string
	user       string
	password   string
	httpClient *http.Client
	connected  bool
	identity   string
}

// NewRESTClient creates an uninitialized REST client.
func NewRESTClient() *RESTClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &RESTClient{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   15 * time.Second,
		},
	}
}

func (c *RESTClient) Connect(host, user, password string, port int, useSSL bool) (string, error) {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	c.baseURL = fmt.Sprintf("%s://%s:%d/rest", scheme, host, port)
	c.user = user
	c.password = password

	req, err := http.NewRequest("GET", c.baseURL+"/system/identity", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(user, password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.connected = false
		return "", fmt.Errorf("không thể kết nối đến %s:%d: %w", host, port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		c.connected = false
		return "", fmt.Errorf("sai tên đăng nhập hoặc mật khẩu")
	}
	if resp.StatusCode == 404 {
		c.connected = false
		return "", fmt.Errorf("REST API không tìm thấy. Vui lòng bật www service trên router")
	}
	if resp.StatusCode >= 400 {
		c.connected = false
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("lỗi kết nối (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ident map[string]interface{}
	if err := json.Unmarshal(body, &ident); err != nil {
		// Could be array
		var arr []map[string]interface{}
		if err2 := json.Unmarshal(body, &arr); err2 == nil && len(arr) > 0 {
			ident = arr[0]
		}
	}

	name, _ := ident["name"].(string)
	if name == "" {
		name = host
	}
	c.identity = name
	c.connected = true
	return c.identity, nil
}

func (c *RESTClient) Disconnect() error {
	c.connected = false
	c.baseURL = ""
	c.user = ""
	c.password = ""
	c.identity = ""
	return nil
}

func (c *RESTClient) IsConnected() bool {
	return c.connected
}

func (c *RESTClient) GetIdentity() string {
	return c.identity
}

func (c *RESTClient) Get(path string) ([]map[string]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("chưa kết nối đến router")
	}
	cleanPath := strings.Trim(path, "/")
	req, err := http.NewRequest("GET", c.baseURL+"/"+cleanPath, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET /%s thất bại (%d): %s", cleanPath, resp.StatusCode, string(body))
	}

	// Result can be []map[string]interface{} or map[string]interface{}
	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err == nil {
		return list, nil
	}

	var single map[string]interface{}
	if err := json.Unmarshal(body, &single); err == nil {
		return []map[string]interface{}{single}, nil
	}

	return []map[string]interface{}{}, nil
}

func (c *RESTClient) GetOne(path string) (map[string]interface{}, error) {
	items, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items[0], nil
	}
	return map[string]interface{}{}, nil
}

func (c *RESTClient) Add(path string, data map[string]interface{}) (map[string]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("chưa kết nối đến router")
	}
	cleanPath := strings.Trim(path, "/")
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/"+cleanPath, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("POST /%s thất bại (%d): %s", cleanPath, resp.StatusCode, string(body))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)
	return res, nil
}

func (c *RESTClient) Set(path string, itemID string, data map[string]interface{}) error {
	if !c.connected {
		return fmt.Errorf("chưa kết nối đến router")
	}
	cleanPath := strings.Trim(path, "/")
	url := c.baseURL + "/" + cleanPath
	if itemID != "" {
		url += "/" + itemID
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PATCH %s thất bại (%d): %s", url, resp.StatusCode, string(body))
	}
	return nil
}

func (c *RESTClient) Remove(path string, itemID string) error {
	if !c.connected {
		return fmt.Errorf("chưa kết nối đến router")
	}
	cleanPath := strings.Trim(path, "/")
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, cleanPath, itemID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s thất bại (%d): %s", url, resp.StatusCode, string(body))
	}
	return nil
}

func (c *RESTClient) Run(path string, data map[string]interface{}) (map[string]interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("chưa kết nối đến router")
	}
	cleanPath := strings.Trim(path, "/")
	jsonData, _ := json.Marshal(data)
	if data == nil {
		jsonData = []byte("{}")
	}

	req, err := http.NewRequest("POST", c.baseURL+"/"+cleanPath, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("RUN command /%s thất bại (%d): %s", cleanPath, resp.StatusCode, string(body))
	}

	var res map[string]interface{}
	_ = json.Unmarshal(body, &res)
	return res, nil
}

func (c *RESTClient) RemoveByComment(path, comment string) (int, error) {
	return RemoveByCommentHelper(c, path, comment)
}

func (c *RESTClient) RemoveAllMatching(path string, filters map[string]string) (int, error) {
	return RemoveAllMatchingHelper(c, path, filters)
}

func (c *RESTClient) ExecRaw(cmd string) (string, error) {
	return "", fmt.Errorf("ExecRaw chỉ hỗ trợ trên kết nối SSH")
}
