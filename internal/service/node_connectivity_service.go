package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NodeOfflineConnectivityResult is the API response of downloader test.
type NodeOfflineConnectivityResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Downloader string `json:"downloader"`
	RPCURL     string `json:"rpc_url"`
	Version    string `json:"version"`
	TestedAt   string `json:"tested_at"`
}

func TestNodeOfflineConnectivity(payload NodeOfflinePayload) (*NodeOfflineConnectivityResult, error) {
	downloader := strings.TrimSpace(payload.Downloader)
	if downloader == "" {
		downloader = "Aria2"
	}

	url := strings.TrimSpace(payload.RPCURL)
	if url == "" {
		return nil, fmt.Errorf("rpc server url cannot be empty")
	}

	client := &http.Client{Timeout: 8 * time.Second}

	switch downloader {
	case "Aria2":
		return testAria2Connectivity(client, url, strings.TrimSpace(payload.RPCSecret))
	case "qBittorrent":
		return testQBittorrentConnectivity(client, url, strings.TrimSpace(payload.RPCSecret))
	case "Transmission":
		return testTransmissionConnectivity(client, url, strings.TrimSpace(payload.RPCSecret))
	default:
		return nil, fmt.Errorf("unsupported downloader type")
	}
}

func testAria2Connectivity(client *http.Client, url, secret string) (*NodeOfflineConnectivityResult, error) {
	params := []interface{}{}
	if secret != "" {
		params = append(params, "token:"+secret)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "node-connectivity",
		"method":  "aria2.getVersion",
		"params":  params,
	})

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build aria2 request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect aria2 failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("aria2 returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var data struct {
		Result struct {
			Version string `json:"version"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode aria2 response failed: %w", err)
	}
	if data.Error != nil {
		return nil, fmt.Errorf("aria2 error: %s", data.Error.Message)
	}

	version := strings.TrimSpace(data.Result.Version)
	if version == "" {
		version = "unknown"
	}

	return &NodeOfflineConnectivityResult{
		Success:    true,
		Message:    fmt.Sprintf("Aria2 通信成功，版本 %s", version),
		Downloader: "Aria2",
		RPCURL:     url,
		Version:    version,
		TestedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func testQBittorrentConnectivity(client *http.Client, url, secret string) (*NodeOfflineConnectivityResult, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build qBittorrent request failed: %w", err)
	}

	if secret != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(secret)))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect qBittorrent failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("qBittorrent returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	version := strings.TrimSpace(string(raw))
	if version == "" {
		version = "unknown"
	}

	return &NodeOfflineConnectivityResult{
		Success:    true,
		Message:    fmt.Sprintf("qBittorrent 通信成功，版本 %s", version),
		Downloader: "qBittorrent",
		RPCURL:     url,
		Version:    version,
		TestedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func testTransmissionConnectivity(client *http.Client, url, secret string) (*NodeOfflineConnectivityResult, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"method": "session-get",
	})

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build Transmission request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(secret)))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect Transmission failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		sessionID := strings.TrimSpace(resp.Header.Get("X-Transmission-Session-Id"))
		if sessionID == "" {
			return nil, fmt.Errorf("transmission requires session id but did not provide one")
		}

		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("rebuild Transmission request failed: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Transmission-Session-Id", sessionID)
		if secret != "" {
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(secret)))
		}

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("reconnect Transmission failed: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("Transmission returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var data struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode Transmission response failed: %w", err)
	}

	if strings.TrimSpace(data.Result) == "" {
		data.Result = "success"
	}

	return &NodeOfflineConnectivityResult{
		Success:    true,
		Message:    fmt.Sprintf("Transmission 通信成功，结果 %s", data.Result),
		Downloader: "Transmission",
		RPCURL:     url,
		Version:    data.Result,
		TestedAt:   time.Now().Format(time.RFC3339),
	}, nil
}
