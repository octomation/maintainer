---
title: Local service
description: Serve maintainer over HTTP at http://maintainer.localhost:17841/ on your computer, as your user.
---

# Local service

> [!NOTE]
> `maintainer service` is not released yet, and the Homebrew steps below apply from the release that ships it. To try it now, build `maintainer` from a source checkout:
>
> ```sh
> git clone https://github.com/octomation/maintainer.git
> cd maintainer && go install .
> ```

`maintainer service run` serves maintainer over HTTP at
`http://maintainer.localhost:17841/`. This version answers with a page that
says the service is running, and nothing else. It needs no token, configuration or network
access, runs as your user and listens only on loopback: `.localhost` names
resolve to your own computer, so there is nothing to set up and no
administrator step.

## Install and start

On macOS, the Homebrew Formula `octolab/tap/maintainer` installs the CLI and
the service definition; the service is the same binary as the CLI and
upgrades with it.

```sh
brew install --formula octolab/tap/maintainer
brew services start octolab/tap/maintainer
curl -fsS http://maintainer.localhost:17841/ -o /dev/null -w '%{http_code}\n'
```

```text
200
```

Open `http://maintainer.localhost:17841/` in a browser to see the page.
The service runs as your user and logs to
`$(brew --prefix)/var/log/maintainer.log`. Installing the Formula does not
start it. Do not run `sudo brew services`.

## Move from the Cask

If you installed the Cask, switch to the Formula once:

```sh
brew uninstall --cask maintainer && brew install --formula octolab/tap/maintainer
```

Your fetch configuration and state stay where they are.

## Upgrade, stop and remove

```sh
brew update
brew upgrade --formula octolab/tap/maintainer
brew services restart octolab/tap/maintainer
brew services stop octolab/tap/maintainer
brew uninstall --formula octolab/tap/maintainer
```

## Run it yourself

Without Homebrew, run the service in a terminal:

```sh
maintainer service run
```

```text
maintainer service: 127.0.0.1:17841, [::1]:17841 (URL: http://maintainer.localhost:17841/)
```

`maintainer.localhost` resolves to both `127.0.0.1` and `::1`, so the service
listens on both. If the system has no IPv6 loopback, the service says so,
skips `[::1]:17841` and keeps running on `127.0.0.1:17841`. `--listen`
replaces both addresses; repeat it for several:

```sh
maintainer service run --listen 127.0.0.1:8080
maintainer service run --listen 127.0.0.1:8080 --listen '[::1]:8080'
```

Each value is a literal loopback IP address with a port, such as
`127.0.0.1:17841` or `[::1]:17841`; `localhost`, `0.0.0.0` or another address
is refused with exit code `2`. An address given with `--listen` that cannot be
bound, as when its port is taken, stops the service with an error. `Ctrl-C` or
`SIGTERM` stops it once the responses in flight are sent. Check it from
another terminal:

```sh
curl -fsS http://maintainer.localhost:17841/
```

```html
<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>maintainer service</title></head>
<body><p>Service is running.</p></body>
</html>
```

Browsers and `curl`
resolve `.localhost` names themselves; on Linux, a system without such a
resolver reaches the service at `http://127.0.0.1:17841/`.

## What it answers

| Request | Response |
| --- | --- |
| `GET` or `HEAD` of `/` | `200`, the page as `text/html`, never cached (`Cache-Control: no-store`) |
| any other path | `404 Not Found` |
| any other method | `405 Method Not Allowed`, `Allow: GET, HEAD` |

Any program on your computer can read the page; the service holds nothing
private.
