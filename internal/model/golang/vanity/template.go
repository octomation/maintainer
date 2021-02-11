package vanity

import "html/template"

var tpl = template.Must(template.New("vanity").Parse(`<!DOCTYPE html>
<html lang="en">
    <head>
        <meta name="go-import" content="{{.Import.Prefix}} {{.Import.VCS}} {{.Import.RepoRoot}}">
        <meta name="go-source" content="{{.Import.Prefix}} {{.Source.URL}} {{.Source.Dir}} {{.Source.File}}">
        <meta http-equiv="content-type" content="text/html; charset=utf-8">
        <meta http-equiv="refresh" content="0; url=https://pkg.go.dev/{{.Package}}">
        <title>{{.Package}}</title>
    </head>
    <body>
        Nothing to see here. Please <a href="https://pkg.go.dev/{{.Package}}">move along</a>.
    </body>
</html>
`))
