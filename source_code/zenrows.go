package source_code

import (
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
)

var _ SourceGetter = &ZenRows{}

type ZenRows struct {
	client   *http.Client
	enabled  bool
	apiKey   string
	HostURL  string
	JSRender bool
}

func ZenRowFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("zenrows", pflag.ExitOnError)
	fs.String("zenrows-api-key", "", "ZenRows API Key")
	fs.Bool("zenrows-js-render", false, "Js Render to encode a Javascript")
	fs.String("zenrows-host-url", "https://api.zenrows.com/v1/", "Host URL")
	return fs
}

func NewZenRowsFromFlags(client *http.Client) *ZenRows {
	return &ZenRows{
		client:   client,
		apiKey:   viper.GetString("zenrows-api-key"),
		HostURL:  viper.GetString("zenrows-host-url"),
		JSRender: viper.GetBool("zenrows-js-render"),
	}
}

func NewZenRows(client *http.Client, apiKey string, JSRender bool) *ZenRows {
	return &ZenRows{
		client:   client,
		apiKey:   apiKey,
		HostURL:  "https://api.zenrows.com/v1/",
		JSRender: JSRender,
	}
}

func (z *ZenRows) Get(ctx context.Context, endpoint string, options ...SourceOptions) ([]byte, int, error) {
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (z *ZenRows) Ping(ctx context.Context) bool {
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
	return z.enabled

}

func (z *ZenRows) Enabled() bool {
	return z.enabled
}

func (z *ZenRows) Name() string {
	return "zen-rows"
}

func (z *ZenRows) buildRequest(ctx context.Context, endpoint string) (*http.Request, error) {
	values := url.Values{}
	values.Add("apiKey", z.apiKey)
	values.Add("url", endpoint)
	if z.JSRender {
		values.Add("js_render", "true")
	}

	endpoint = fmt.Sprintf("%s?%s", z.HostURL, values.Encode())

	return http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
}
