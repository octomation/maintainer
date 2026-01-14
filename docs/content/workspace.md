---
title: Workspace
description: The shared layout, pins and state that define which checkouts fetch and status manage.
---

# Workspace

> [!NOTE]
> `maintainer fetch` and `maintainer status` are not released yet. To try them, build `maintainer` from a source checkout:
>
> ```sh
> git clone https://github.com/octomation/maintainer.git
> cd maintainer && go install .
> ```

`fetch` and `status` share one local discovery boundary. A workspace describes
both where ordinary repositories belong and which existing checkouts stay put:

```toml
[workspace]
root = "~/Development"
path = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
pins = ["prototyping"] # optional; omit when there is no such directory

[defaults]
clone_url = "ssh"
concurrency = 8

[[repos]]
match = { id = 154873464 }
path = "~/.dotfiles"
```

The config remains `fetch.toml` or `fetch.yaml`, shared by both commands. One
workspace is supported today; named/multiple workspaces are reserved for later.

## Layout is scope

The default `path` matches exactly `{public,private,internal}/<owner>/<repo>`.
Visibility is a finite set, not a wildcard directory name. Adding `research`,
`prototyping`, or any future directory beside those branches does not expand
either command's scope. Out-of-scope branches are pruned before Git inspection
or identity resolution. Finding a checkout stops descent (no submodule scan).
Directory symlinks are not followed by discovery; use a real explicit path.

Templates support literals and scalar `.Owner`, `.Repo`, `.Visibility`,
`.DefaultBranch`, `.IsFork`, `.IsTemplate`, `.IsArchived`, with `lower`/`upper`.
Each field occupies part of one path component. Conditions, loops and other
non-reversible expressions are rejected, never interpreted as a broad scan.
Per-owner path templates add their corresponding managed regions.

For managed checkouts, a confirmed rename/transfer/visibility change updates
the target rendered from GitHub metadata. Moves happen only with `fetch --apply`.
Local path spelling need not already match the current origin: stale names are
still discovered by layout, then identity is verified by fetch.

## Pins keep checkouts in place

`pins` contains literal paths to existing checkouts or trees of checkouts.
Relative paths use `workspace.root`; absolute paths and `~/…` also work.
Duplicate and overlapping entries do not duplicate rows. Missing pin roots are
visible errors. A workspace pin overrides layout management at that location:
fetch never clones or moves into it, and never moves its checkout out.

Each discovered checkout is adopted at its current path. Future owner/name
changes update its verified origin; visibility changes update metadata, not
the directory. Explicitly included local repositories can be verified by ID
even outside the owners selected for discovery of new remote repositories.
Unknown/inaccessible identity blocks changes; status remains entirely offline.

A broad pin does not choose between duplicate checkouts. Both rows are marked
as ambiguous in status and fetch reports a conflict. A specific per-repository
`path` selects the active checkout and keeps mutation precedence over copies.
Other copies remain visible as `orphan`, with their own local
changes and the selected active path. They are never fetched, moved or deleted.

Pins granted by workspace rules are recorded with `pin_source = "workspace"`.
Removing the rule suspends that checkout, even if it now matches the managed
layout. It does not silently enable moves. Re-add the pin or explicitly select
the checkout with a per-repo path. Existing per-repo `pinned_path` records keep
their legacy persistent-pin semantics.

## State and boundaries

An ordinary state record outside the current scope does not grant management
permission. Both commands show it as `orphan`; fetch keeps the record,
and does not move or re-clone it. Changing root is not a request to import
checkouts from the previous root.

Unknown out-of-scope copies are intentionally not searched. A remote repository
selected for a new managed clone can therefore also have an unknown copy in
`research`. There is no global duplicate search outside the workspace.

Former active paths are retained as diagnostic-only `previous_paths` on
successful adoption of a different checkout. Existing copies remain visible
after the switch, including outside root. These exact known paths can be
inspected for reporting; they never become recursive discovery roots or gain
mutation permission. Missing historical copies are omitted.

An orphan with JSON `orphan_reason: remote-gone` means fetch confirmed a GitHub
ID returned 404. Human output uses just `orphan`, without bracketed codes. Offline
status can only show the last saved observation, labelled cached with its time.
Plan-only does not persist observations; apply does, without deleting the clone.
Missing listings, authentication failures and network errors are not proof of
deletion. A later successful confirmation clears the cached orphan observation.

Status reads local checkouts and eligible remembered paths without tokens,
network, index refresh or state writes. Fetch additionally discovers remote
repositories and plans their materialisation; thus a not-yet-cloned remote
repository can appear in a fetch plan but not in an offline status table.

## Migration

Legacy `defaults.root/path` are still read as the workspace layout, with the
same new bounded discovery rules. To adopt the new spelling, move those two
keys into `[workspace]`; leave transport, concurrency and state_file in
`[defaults]`. Using both spatial forms is an error. No config or state file is
automatically rewritten by loading it. `status --root` overrides the workspace
root and rebases relative pins, not absolute pins.
