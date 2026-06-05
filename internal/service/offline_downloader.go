package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type offlineDownloader interface {
	Submit(ctx context.Context, task *modelOfflineRuntime) (string, error)
	Status(ctx context.Context, task *modelOfflineRuntime) (*offlineDownloaderStatus, error)
	Pause(ctx context.Context, task *modelOfflineRuntime) error
	Resume(ctx context.Context, task *modelOfflineRuntime) error
	Delete(ctx context.Context, task *modelOfflineRuntime) error
}

type modelOfflineRuntime struct {
	URL            string
	Downloader     string
	RPCURL         string
	RPCSecret      string
	TaskOptions    string
	TempDir        string
	RemoteTaskID   string
	WaitForSeeding bool
}

type offlineDownloaderStatus struct {
	Status          string
	Completed       bool
	WaitingSeeding  bool
	DownloadedBytes int64
	TotalBytes      int64
	SpeedText       string
}

func newOfflineDownloader(name string) offlineDownloader {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "aria2":
		return &aria2Downloader{client: &http.Client{Timeout: 30 * time.Second}}
	case "qbittorrent", "transmission":
		return unsupportedDownloader(strings.TrimSpace(name))
	default:
		return unsupportedDownloader(strings.TrimSpace(name))
	}
}

type unsupportedDownloader string

func (d unsupportedDownloader) Submit(context.Context, *modelOfflineRuntime) (string, error) {
	return "", fmt.Errorf("%s does not support submitting offline download tasks yet", string(d))
}

func (d unsupportedDownloader) Status(context.Context, *modelOfflineRuntime) (*offlineDownloaderStatus, error) {
	return nil, fmt.Errorf("%s does not support querying offline download status yet", string(d))
}

func (d unsupportedDownloader) Pause(context.Context, *modelOfflineRuntime) error {
	return fmt.Errorf("%s does not support pausing offline download tasks yet", string(d))
}

func (d unsupportedDownloader) Resume(context.Context, *modelOfflineRuntime) error {
	return fmt.Errorf("%s does not support resuming offline download tasks yet", string(d))
}

func (d unsupportedDownloader) Delete(context.Context, *modelOfflineRuntime) error {
	return fmt.Errorf("%s does not support deleting offline download tasks yet", string(d))
}

type aria2Downloader struct {
	client *http.Client
}

type aria2Request struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type aria2Response struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type aria2Status struct {
	Status          string `json:"status"`
	CompletedLength string `json:"completedLength"`
	TotalLength     string `json:"totalLength"`
	DownloadSpeed   string `json:"downloadSpeed"`
	UploadSpeed     string `json:"uploadSpeed"`
}

func (d *aria2Downloader) Submit(ctx context.Context, task *modelOfflineRuntime) (string, error) {
	options, err := parseAria2Options(task.TaskOptions)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(task.TempDir) != "" {
		options["dir"] = strings.TrimSpace(task.TempDir)
	}
	params := []interface{}{[]string{task.URL}}
	if len(options) > 0 {
		params = append(params, options)
	}
	var gid string
	if err := d.call(ctx, task, "aria2.addUri", params, &gid); err != nil {
		return "", err
	}
	return gid, nil
}

func (d *aria2Downloader) Status(ctx context.Context, task *modelOfflineRuntime) (*offlineDownloaderStatus, error) {
	var raw aria2Status
	if err := d.call(ctx, task, "aria2.tellStatus", []interface{}{task.RemoteTaskID, []string{"status", "completedLength", "totalLength", "downloadSpeed", "uploadSpeed"}}, &raw); err != nil {
		return nil, err
	}
	completed := parseInt64(raw.CompletedLength)
	total := parseInt64(raw.TotalLength)
	progressComplete := raw.Status == "complete"
	waitingSeeding := false
	if task.WaitForSeeding && isMagnetURL(task.URL) && raw.Status == "active" && total > 0 && completed >= total {
		waitingSeeding = true
		progressComplete = false
	}
	if !task.WaitForSeeding && isMagnetURL(task.URL) && raw.Status == "active" && total > 0 && completed >= total {
		progressComplete = true
	}
	return &offlineDownloaderStatus{
		Status:          raw.Status,
		Completed:       progressComplete,
		WaitingSeeding:  waitingSeeding,
		DownloadedBytes: completed,
		TotalBytes:      total,
		SpeedText:       formatBytes(parseInt64(raw.DownloadSpeed)) + "/s",
	}, nil
}

func (d *aria2Downloader) Pause(ctx context.Context, task *modelOfflineRuntime) error {
	var ignored json.RawMessage
	return d.call(ctx, task, "aria2.pause", []interface{}{task.RemoteTaskID}, &ignored)
}

func (d *aria2Downloader) Resume(ctx context.Context, task *modelOfflineRuntime) error {
	var ignored json.RawMessage
	return d.call(ctx, task, "aria2.unpause", []interface{}{task.RemoteTaskID}, &ignored)
}

func (d *aria2Downloader) Delete(ctx context.Context, task *modelOfflineRuntime) error {
	var ignored json.RawMessage
	err := d.call(ctx, task, "aria2.remove", []interface{}{task.RemoteTaskID}, &ignored)
	if err == nil {
		return nil
	}
	return d.call(ctx, task, "aria2.removeDownloadResult", []interface{}{task.RemoteTaskID}, &ignored)
}

func (d *aria2Downloader) call(ctx context.Context, task *modelOfflineRuntime, method string, params []interface{}, target interface{}) error {
	rpcURL := strings.TrimSpace(task.RPCURL)
	if rpcURL == "" {
		return fmt.Errorf("aria2 rpc_url cannot be empty")
	}
	params = appendAuthToken(task.RPCSecret, params)
	body, err := json.Marshal(aria2Request{JSONRPC: "2.0", ID: "xingyunpan", Method: method, Params: params})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("aria2 rpc returned HTTP %s", resp.Status)
	}
	var decoded aria2Response
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return err
	}
	if decoded.Error != nil {
		return fmt.Errorf("aria2 rpc error %d: %s", decoded.Error.Code, decoded.Error.Message)
	}
	if target == nil {
		return nil
	}
	if len(decoded.Result) == 0 || string(decoded.Result) == "null" {
		return nil
	}
	return json.Unmarshal(decoded.Result, target)
}

func appendAuthToken(secret string, params []interface{}) []interface{} {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return params
	}
	result := make([]interface{}, 0, len(params)+1)
	result = append(result, "token:"+secret)
	result = append(result, params...)
	return result
}

func parseAria2Options(raw string) (map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}, nil
	}
	var options map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil, fmt.Errorf("parse aria2 task_options failed: %w", err)
	}
	return options, nil
}

func parseInt64(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func isMagnetURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "magnet:")
}
