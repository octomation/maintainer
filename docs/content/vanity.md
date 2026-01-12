---
title: Go vanity URLs
description: Generate static Go import-path pages from a module manifest.
---

# Go vanity URLs

Give your Go packages a stable, recognizable home such as `go.octolab.org/toolset`, even when their source lives on GitHub. `go vanity build` makes this practical: it reads `modules.yml` and writes static `index.html` pages with the `go-import` and `go-source` metadata Go needs. You can serve the result with GitHub Pages under your custom domain.

## Prepare a module manifest

Create `modules.yml` in your working directory:

```yaml
- name: go.example.org/toolkit
  prefix: go.example.org/toolkit
  import:
    - vcs: git
      url: https://github.com/example/toolkit
      source:
        url: https://github.com/example/toolkit
        dir: https://github.com/example/toolkit/tree/main{/dir}
        file: https://github.com/example/toolkit/blob/main{/dir}/{file}#L{line}
  packages:
    - go.example.org/toolkit
    - go.example.org/toolkit/config
```

The first `import` entry supplies the metadata; without an `import` entry, the module is skipped. `prefix` falls back to `name` if omitted. Every package path **must be inside the prefix** (or equal to it), as in the example. A path outside the prefix currently makes the generator loop instead of returning an error. Parent paths are generated too. Check the source URL templates against your code host before publishing.

## Build and publish

```sh
maintainer go vanity --host go.example.org --file modules.yml build ./site
```

The optional final argument is the output directory; without it the command writes in the current directory. It creates an `index.html` under each generated path, so use a dedicated output directory and review it before copying to your static host. Your host must serve these pages at the matching import paths. Each page also redirects a human visitor to `pkg.go.dev`.

This is a generator, not a deployment command. Re-run it when the manifest changes, then publish the resulting files through GitHub Pages or your normal static hosting workflow. The branded import path can stay stable while the source repository remains on GitHub.
