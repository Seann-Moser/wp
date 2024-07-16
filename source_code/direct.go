package source_code

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

var _ SourceGetter = &Direct{}

type Direct struct {
	client  *http.Client
	enabled bool
	err     error
}

func NewDirect(client *http.Client) *Direct {
	return &Direct{
		client: client,
	}
}

func (d *Direct) Get(ctx context.Context, endpoint string, options ...SourceOptions) ([]byte, int, error) {
	if !d.Enabled() {
		return nil, http.StatusNotImplemented, fmt.Errorf("%s is not enabled", d.Name())
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	resp, err := d.client.Do(r)
	if err != nil {
		return nil, 0, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if HasChallange(body) {
		return body, resp.StatusCode, HasChallengeErr
	}
	return body, resp.StatusCode, nil
}

func (d *Direct) Ping(ctx context.Context) bool {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, PingURL, nil)
	if err != nil {
		d.enabled = false
		d.err = err
		return false
	}
	resp, err := d.client.Do(r)
	if err != nil {
		d.enabled = false
		d.err = err
		return d.enabled
	}
	d.enabled = resp.StatusCode >= http.StatusOK || resp.StatusCode < 300
	if !d.enabled {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if HasChallange(body) {
		d.enabled = false
		d.err = HasChallengeErr
		return false
	}
	return d.enabled
}

func (d *Direct) Enabled() bool {
	return d.enabled
}

func (d *Direct) Err() error {
	return d.err
}

func (d *Direct) Name() string {
	return "direct"
}
