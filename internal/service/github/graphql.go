package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
)

// graphql is the GitHub GraphQL API v4 endpoint. Unlike the profile HTML it
// answers from a single consistent source: the same query never disagrees with
// itself between two calls, which the year-scoped calendar page does.
const graphql = "https://api.github.com/graphql"

// queryGraphQL executes the query with the provided variables and unmarshals
// the `data` object of the response into dst.
func (srv *Service) queryGraphQL(
	ctx context.Context,
	query string, vars map[string]any,
	dst any,
) error {
	payload, err := json.Marshal(struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables,omitempty"`
	}{query, vars})
	if err != nil {
		return fmt.Errorf("build graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphql, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.client.Client().Do(req)
	if err != nil {
		return fmt.Errorf("send graphql request: %w", err)
	}
	defer safe.Close(resp.Body, unsafe.Ignore)

	// The GraphQL API needs a token even for public data, so an unauthorized
	// or throttled client fails here instead of in the payload.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql request: unexpected status %s: %s", resp.Status, snippet(resp.Body))
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decode graphql response: %w", err)
	}
	// A GraphQL failure is reported as 200 OK with a non-empty errors array.
	if len(envelope.Errors) > 0 {
		return envelope.Errors
	}

	return json.Unmarshal(envelope.Data, dst)
}

// graphqlError is a single entry of the GraphQL `errors` array.
type graphqlError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// graphqlErrors is a non-empty GraphQL `errors` array.
type graphqlErrors []graphqlError

func (errs graphqlErrors) Error() string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		if err.Type == "" {
			messages = append(messages, err.Message)
			continue
		}
		messages = append(messages, err.Type+": "+err.Message)
	}
	return "graphql: " + strings.Join(messages, "; ")
}

// snippet returns the leading part of a response body for diagnostics.
func snippet(r io.Reader) string {
	raw, err := io.ReadAll(io.LimitReader(r, 1<<10))
	if err != nil {
		return "<unreadable body>"
	}
	return strings.TrimSpace(string(raw))
}
