package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

// helloTimeout bounds the request that checks the service.
const helloTimeout = 3 * time.Second

// launch opens path in the browser; the label is never one of its arguments.
var launch = func(ctx context.Context, path string) (*exec.Cmd, error) {
	name := "xdg-open"
	if goruntime.GOOS == "darwin" {
		name = "open"
	}
	cmd := exec.CommandContext(ctx, name, path)
	return cmd, cmd.Start()
}

// redirectWait is how long the redirect page stays for the browser to read.
var redirectWait = 5 * time.Second

func openCommand() *cobra.Command {
	var listen, configPath string
	var print bool
	command := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "open",
		Short: "open the running local maintainer service in the browser",
		Long: "Opens the running local maintainer service in the browser at the\n" +
			"private address of its current run, the only address that serves the\n" +
			"data. It reads the address from listen in service.toml, else uses\n" +
			"127.0.0.1:17841 (--listen names another), and the label of the run from\n" +
			"$XDG_STATE_HOME/maintainer/service/<port>.label, checks that the\n" +
			"service on that address knows the label, and opens the address through a\n" +
			"temporary page, so the label is in no command line. --print prints the\n" +
			"address instead. Run it again after every start of the service.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			address, err := openAddress(listen, cmd.Flags().Changed("listen"), configPath, cmd.Flags().Changed("config"))
			if err != nil {
				return exit.WithUser(err)
			}
			_, portText, _ := net.SplitHostPort(address)
			port, _ := strconv.Atoi(portText)
			if port == 0 {
				return exit.WithUser(fmt.Errorf("the address %s has no fixed port: name the port of the service with --listen", address))
			}
			notRunning := fmt.Errorf("the service is not running on %s: start it with `brew services start octolab/tap/maintainer` or `maintainer service run`", address)
			dir := stateDir()
			label, err := web.ReadLabel(dir, port)
			if errors.Is(err, web.ErrNoLabel) {
				return notRunning
			}
			if err != nil {
				return err
			}
			nonce := make([]byte, 32)
			if _, err := rand.Read(nonce); err != nil {
				return err
			}
			proof, err := hello(cmd.Context(), address, nonce)
			if errors.Is(err, errNoAnswer) {
				return notRunning
			}
			if err != nil || !web.ProofEqual(proof, web.Proof(label, nonce)) {
				return fmt.Errorf("another program listens on %s: it is not the service that wrote the label; stop it, then start the service again", address)
			}
			url := web.LabelURL(label, port)
			if print {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), url)
				return err
			}
			if err := openURL(cmd.Context(), dir, url, port); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "maintainer service: opened the service in the browser")
			return nil
		},
	}
	command.Flags().StringVar(&listen, "listen", "", "address of the service (default: the first of listen in service.toml, else "+Listen[0]+")")
	command.Flags().StringVar(&configPath, "config", "", "service config path (default: $XDG_CONFIG_HOME/maintainer/service.toml)")
	command.Flags().BoolVar(&print, "print", false, "print the address to stdout and open nothing")
	return command
}

// openAddress returns the address of the service: --listen, else the first
// of listen in service.toml, else the default. It never writes the config.
func openAddress(listen string, listenSet bool, configPath string, configSet bool) (string, error) {
	if listenSet {
		if err := config.CheckLoopback(listen); err != nil {
			return "", fmt.Errorf("--listen %w", err)
		}
		return listen, nil
	}
	fs := afero.NewOsFs()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	lookup := config.ServiceConfigLookup{Explicit: configPath, ExplicitSet: configSet, Getenv: os.Getenv, Home: home}
	path, _ := lookup.Resolve(fs)
	if configSet {
		path = config.ResolveAgainst(home, configPath)
	}
	service, err := config.LoadService(fs, path)
	if err != nil {
		return "", err
	}
	if len(service.Listen) > 0 {
		return service.Listen[0], nil
	}
	return Listen[0], nil
}

// errNoAnswer means nothing answered on the address.
var errNoAnswer = errors.New("no answer")

// hello asks the service on address for the proof of nonce: on loopback,
// without a proxy, without following a redirect, with a short timeout.
func hello(ctx context.Context, address string, nonce []byte) (string, error) {
	client := &http.Client{
		Timeout:       helloTimeout,
		Transport:     &http.Transport{Proxy: nil, DisableKeepAlives: true},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/api/v0/hello?nonce="+hex.EncodeToString(nonce), http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", errNoAnswer
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("hello answered %d", resp.StatusCode)
	}
	var body struct {
		Data web.Hello `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<10)).Decode(&body); err != nil {
		return "", fmt.Errorf("hello: %w", err)
	}
	return body.Data.Proof, nil
}

// openURL opens url, the address of the run on port, in the browser through
// a redirect page: created exclusively with mode 0600 in a new folder of mode
// 0700 under the state folder, and removed after a short wait, so the label is
// in no command line.
func openURL(ctx context.Context, stateDir, url string, port int) error {
	if err := web.PrepareStateDir(stateDir); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(stateDir, "open-")
	if err != nil {
		return fmt.Errorf("create the redirect page: %w", err)
	}
	defer os.RemoveAll(dir)
	page := filepath.Join(dir, "open.html")
	file, err := os.OpenFile(page, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create the redirect page: %w", err)
	}
	_, werr := io.WriteString(file, redirectPage(url, web.URL(port)+"status?back=1"))
	if cerr := file.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return fmt.Errorf("write the redirect page: %w", werr)
	}
	cmd, err := launch(ctx, page)
	if err != nil {
		return fmt.Errorf("open the browser: %w; run `maintainer service open --print` and open the address yourself", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	timer := time.NewTimer(redirectWait)
	defer timer.Stop()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("open the browser: %w; run `maintainer service open --print` and open the address yourself", err)
		}
		<-timer.C
	case <-timer.C:
	case <-ctx.Done():
	}
	return nil
}

// redirectPage is the page that opens target, the address of the run, after a
// visit to plain, a page of the plain address: the browser loads that page as
// a top-level document, so its Clear-Site-Data and its own cleanup empty the
// storage that earlier releases left there, and the page goes back here
// (back=1). The visit carries no label: not in its URL, not in a header (a
// file page sends no Referer), and the plain page cannot read this one. Only
// the final navigation holds the label. Without script, the page goes to
// target at once.
func redirectPage(target, plain string) string {
	t, p := strconv.Quote(target), strconv.Quote(plain)
	return "<!doctype html>\n<meta charset=\"utf-8\">\n<meta name=\"referrer\" content=\"no-referrer\">\n" +
		"<title>maintainer</title>\n<p>Opening maintainer…</p>\n" +
		"<noscript><meta http-equiv=\"refresh\" content=\"0;url=" + html.EscapeString(target) + "\"></noscript>\n" +
		"<script>\n" +
		"const plain = " + p + ";\n" +
		"const go = () => location.replace(" + t + ");\n" +
		"addEventListener('pageshow', (event) => { if (event.persisted && location.hash === '#cleared') go(); });\n" +
		"if (location.hash === '#cleared') go();\n" +
		// After load, so the visit is a new history entry, not a replacement.
		"else addEventListener('load', () => setTimeout(() => { history.replaceState(null, '', '#cleared'); location.assign(plain); }));\n" +
		"</script>\n"
}
