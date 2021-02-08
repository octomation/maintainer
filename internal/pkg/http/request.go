package http

import (
	"context"
	"net/http"
)

func NewGetRequestWithContext(ctx context.Context, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}
