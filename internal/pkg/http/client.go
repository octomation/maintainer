package http

import (
	"context"
	"net/http"

	"go.octolab.org/toolkit/config"
	"golang.org/x/oauth2"
)

func TokenSourcedClient(ctx context.Context, token config.Secret) *http.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
	return oauth2.NewClient(ctx, source)
}
