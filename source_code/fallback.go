package source_code

import (
	"context"
	"fmt"
)

var _ SourceGetter = &Fallback{}

type Fallback struct {
	Getters []SourceGetter
	status  map[string]bool
}

func NewFallback(getters ...SourceGetter) *Fallback {
	return &Fallback{
		Getters: getters,
		status:  make(map[string]bool),
	}
}

func (f *Fallback) Get(ctx context.Context, endpoint string, options ...SourceOptions) ([]byte, int, error) {
	for _, getter := range f.Getters {
		enabled, found := f.status[endpoint]
		if !found {
			f.status[getter.Name()] = getter.Ping(ctx)
			enabled = f.status[getter.Name()]
		}
		if !enabled {
			continue
		}
		data, status, err := getter.Get(ctx, endpoint, options...)
		if err != nil {
			continue
		}
		return data, status, err
	}
	return nil, 0, fmt.Errorf("no fallback source")
}

func (f *Fallback) Ping(ctx context.Context) bool {
	if len(f.Getters) == 0 {
		return false
	}
	for _, getter := range f.Getters {
		f.status[getter.Name()] = getter.Ping(ctx)
	}
	return f.Enabled()
}

func (f *Fallback) Name() string {
	return "fallback"
}

func (f *Fallback) Enabled() bool {
	if len(f.Getters) == 0 {
		return false
	}
	for _, getter := range f.Getters {
		if !getter.Enabled() {
			return false
		}
	}
	return true
}
