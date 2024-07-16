package source_code

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io"
	"net/http"
)

var _ SourceGetter = &FlareSolver{}

type FlareSolver struct {
	client  *http.Client
	enabled bool
	HostURL string
}

type FlareParserRequest struct {
	Cmd        string `json:"cmd"`
	Url        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
}

func FlareSolverFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("flaresolver", pflag.ExitOnError)
	fs.String("flaresolver-host-url", "http://localhost:8191/v1", "Host URL")
	return fs
}

func NewFlareSolverFromFlags(client *http.Client) *FlareSolver {
	return &FlareSolver{
		client:  client,
		HostURL: viper.GetString("flaresolver-host-url"),
	}
}

func NewFlareSolver(client *http.Client) *FlareSolver {
	return &FlareSolver{
		client:  client,
		HostURL: "http://localhost:8191/v1",
	}
}

func (z *FlareSolver) Get(ctx context.Context, endpoint string, options ...SourceOptions) ([]byte, int, error) {
	if !z.Enabled() {
		return nil, http.StatusNotImplemented, fmt.Errorf("%s is not enabled", z.Name())
	}
	r, err := z.buildRequest(ctx, endpoint)
	if err != nil {
		z.enabled = false
		return nil, http.StatusNotImplemented, err
	}
	resp, err := z.client.Do(r)
	if err != nil {
		return nil, http.StatusNotImplemented, err
	}
	responseBody := FlareResponse{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return []byte(responseBody.Solution.Response), resp.StatusCode, nil
}

func (z *FlareSolver) Ping(ctx context.Context) bool {
	r, err := z.buildRequest(ctx, PingURL)
	if err != nil {
		z.enabled = false
		return false
	}
	resp, err := z.client.Do(r)
	if err != nil {
		z.enabled = false
		return z.enabled
	}
	z.enabled = resp.StatusCode >= http.StatusOK || resp.StatusCode < 300
	if !z.enabled {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if HasChallange(body) {
		z.enabled = false
		return false
	}
	return z.enabled

}

func (z *FlareSolver) Enabled() bool {
	return z.enabled
}

func (z *FlareSolver) Name() string {
	return "flaresolver"
}

func (z *FlareSolver) buildRequest(ctx context.Context, endpoint string) (*http.Request, error) {
	body := FlareParserRequest{
		Cmd:        "request.get",
		Url:        endpoint,
		MaxTimeout: 60000,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, z.HostURL, bytes.NewBuffer(b))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

type FlareResponse struct {
	Solution struct {
		Url     string `json:"url"`
		Status  int    `json:"status"`
		Headers struct {
			Status                  string `json:"status"`
			Date                    string `json:"date"`
			Expires                 string `json:"expires"`
			CacheControl            string `json:"cache-control"`
			ContentType             string `json:"content-type"`
			StrictTransportSecurity string `json:"strict-transport-security"`
			P3P                     string `json:"p3p"`
			ContentEncoding         string `json:"content-encoding"`
			Server                  string `json:"server"`
			ContentLength           string `json:"content-length"`
			XXssProtection          string `json:"x-xss-protection"`
			XFrameOptions           string `json:"x-frame-options"`
			SetCookie               string `json:"set-cookie"`
		} `json:"headers"`
		Response string `json:"response"`
		Cookies  []struct {
			Name     string  `json:"name"`
			Value    string  `json:"value"`
			Domain   string  `json:"domain"`
			Path     string  `json:"path"`
			Expires  float64 `json:"expires"`
			Size     int     `json:"size"`
			HttpOnly bool    `json:"httpOnly"`
			Secure   bool    `json:"secure"`
			Session  bool    `json:"session"`
			SameSite string  `json:"sameSite"`
		} `json:"cookies"`
		UserAgent string `json:"userAgent"`
	} `json:"solution"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Version        string `json:"version"`
}
