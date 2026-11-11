package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"
)

// healthy reports whether an apfel server answers /health at baseURL.
func healthy(baseURL string) bool {
	c := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := c.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ApfelAvailable is true when the apfel binary is installed (macOS 26+ with
// Apple Intelligence) or a server already answers on baseURL.
func ApfelAvailable(baseURL string) bool {
	if baseURL == "" {
		baseURL = DefaultLocalURL
	}
	if healthy(baseURL) {
		return true
	}
	_, err := exec.LookPath("apfel")
	return err == nil
}

// StartApfel returns the URL of a running apfel server. If one already answers
// on baseURL it is reused; otherwise `apfel --serve` is started on a free
// loopback port and stopped by the returned function. Bound to 127.0.0.1 only.
func StartApfel(ctx context.Context, baseURL string) (url string, stop func(), err error) {
	if baseURL == "" {
		baseURL = DefaultLocalURL
	}
	if healthy(baseURL) {
		return baseURL, func() {}, nil
	}
	bin, err := exec.LookPath("apfel")
	if err != nil {
		return "", nil, fmt.Errorf("apfel is not installed (brew install apfel; needs macOS 26+ with Apple Intelligence)")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	cmd := exec.CommandContext(ctx, bin, "--serve", "--host", "127.0.0.1", "--port", fmt.Sprint(port))
	if err := cmd.Start(); err != nil {
		return "", nil, err
	}
	url = fmt.Sprintf("http://127.0.0.1:%d", port)
	stop = func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }
	for i := 0; i < 50; i++ {
		if healthy(url) {
			return url, stop, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	stop()
	return "", nil, fmt.Errorf("apfel --serve did not become healthy on %s", url)
}
