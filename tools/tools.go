// Code generated by github.com/kamilsk/egg. DO NOT EDIT.

//go:build tools
// +build tools

package tools

import (
	_ "github.com/cube2222/octosql/cmd/octosql"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/exp/cmd/gorelease"
	_ "golang.org/x/tools/cmd/goimports"
)

//go:generate go install github.com/cube2222/octosql/cmd/octosql
//go:generate go install github.com/golang/mock/mockgen
//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install golang.org/x/exp/cmd/gorelease
//go:generate go install golang.org/x/tools/cmd/goimports
