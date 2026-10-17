package api

import (
	"fmt"
	"strings"
)

// MikrotikClient defines the common interface for RouterOS REST and SSH communication.
type MikrotikClient interface {
	Connect(host, user, password string, port int, useSSL bool) (string, error)
	Disconnect() error
	IsConnected() bool
	GetIdentity() string

	Get(path string) ([]map[string]interface{}, error)
	GetOne(path string) (map[string]interface{}, error)
	Add(path string, data map[string]interface{}) (map[string]interface{}, error)
	Set(path string, itemID string, data map[string]interface{}) error
	Remove(path string, itemID string) error
	Run(path string, data map[string]interface{}) (map[string]interface{}, error)

	RemoveByComment(path, comment string) (int, error)
	RemoveAllMatching(path string, filters map[string]string) (int, error)
	ExecRaw(cmd string) (string, error)
}

// Helper function to remove items matching a comment
func RemoveByCommentHelper(c MikrotikClient, path, comment string) (int, error) {
	items, err := c.Get(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		if cmt, ok := item["comment"].(string); ok && (cmt == comment || strings.HasPrefix(cmt, comment)) {
			if id, ok := item[".id"].(string); ok && id != "" {
				if err := c.Remove(path, id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}

// Helper function to remove items matching all filter key-values
func RemoveAllMatchingHelper(c MikrotikClient, path string, filters map[string]string) (int, error) {
	items, err := c.Get(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, item := range items {
		match := true
		for k, v := range filters {
			valStr := fmt.Sprintf("%v", item[k])
			if valStr != v {
				match = false
				break
			}
		}
		if match {
			if id, ok := item[".id"].(string); ok && id != "" {
				if err := c.Remove(path, id); err == nil {
					count++
				}
			}
		}
	}
	return count, nil
}
