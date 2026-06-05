package github

import (
	"context"
	"errors"
)

// errGraphQLDeferred marks the deferred GraphQL discoverer (milestone 8).
var errGraphQLDeferred = errors.New("graphql discoverer is deferred to milestone 8 (REST only in the PoC)")

// GraphQLDiscoverer is a placeholder for the milestone-8 GraphQL implementation
// of the Discoverer port (fetch plan §12, §13). The PoC ships REST only; this
// stub keeps the seam visible so a future `--api=graphql` switch is a one-file
// change without touching the Planner (§2.3). It hypothetically collapses the
// 2–3 REST round trips per owner into one cursor-paginated query (§14).
type GraphQLDiscoverer struct{}

var _ Discoverer = (*GraphQLDiscoverer)(nil)

// List is not implemented in the PoC.
func (*GraphQLDiscoverer) List(context.Context, Profile) (Discovery, error) {
	return Discovery{}, errGraphQLDeferred
}
