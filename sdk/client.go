package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"net/url"
	"time"
)

const MaasProxyURL = "http://maas-proxy.vke-system:6789"

// MaasClient ...
type MaasClient struct {
	URL    *url.URL
	Client *http.Client
	ApiKey string
}

func NewMaasClient(apiKey string) (*MaasClient, error) {
	url, err := url.Parse(MaasProxyURL) // 这块写死
	if err != nil {
		return nil, err
	}
	cli := http.DefaultClient
	cli.Transport = NewHTTPMetricsTransport("maas-proxy", func(req *http.Request) string {
		return req.URL.Path
	})
	cli.Timeout = time.Second * 10
	mc := &MaasClient{
		URL:    url,
		Client: cli,
		ApiKey: apiKey,
	}
	return mc, nil
}

func (client *MaasClient) Send(ctx context.Context, method, path string, reqData, respData any) (httpCode int, err error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return
	}
	var respRawData []byte
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Apikey", client.ApiKey)
	resp, err := client.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close() // nolint
	respRawData, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	switch t := respData.(type) {
	case *[]byte:
		*t = respRawData
	case *string:
		*t = string(respRawData)
	case nil:

	default:
		err = json.Unmarshal(respRawData, respData)
	}
	return resp.StatusCode, err

}

func (client *MaasClient) Chat(req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	resp := &openai.ChatCompletionResponse{}
	_, err := client.Send(context.Background(), http.MethodPost, "/v1/chat/completions", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// HTTPMetricsTransport ...
type HTTPMetricsTransport struct {
	http.Transport
	service      string
	actionGetter HTTPActionGetter
}

// NewHTTPMetricsTransport returns a transport which records http request metrics
func NewHTTPMetricsTransport(service string, actionGetter HTTPActionGetter) *HTTPMetricsTransport {
	return &HTTPMetricsTransport{
		service:      service,
		actionGetter: actionGetter,
	}
}

// HTTPActionGetter is an interface which get action from a http request
type HTTPActionGetter func(req *http.Request) string
