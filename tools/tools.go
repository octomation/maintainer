// Code generated by github.com/kamilsk/egg. DO NOT EDIT.

//go:build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/kyoh86/git-vertag"
	_ "github.com/marwan-at-work/mod/cmd/mod"
	_ "golang.org/x/exp/cmd/gorelease"
	_ "golang.org/x/tools/cmd/benchcmp"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/tools/cmd/gomvpkg"
	_ "golang.org/x/tools/cmd/gorename"
)

//go:generate go install github.com/golang/mock/mockgen
//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install github.com/kyoh86/git-vertag
//go:generate go install github.com/marwan-at-work/mod/cmd/mod
//go:generate go install golang.org/x/exp/cmd/gorelease
//go:generate go install golang.org/x/tools/cmd/benchcmp
//go:generate go install golang.org/x/tools/cmd/goimports
//go:generate go install golang.org/x/tools/cmd/gomvpkg
//go:generate go install golang.org/x/tools/cmd/gorename
