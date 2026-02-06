package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jimezsa/jobcli/internal/config"
	"github.com/jimezsa/jobcli/internal/network"
	fhttp "github.com/bogdanfinn/fhttp"
)

type ProxiesCmd struct {
	Check ProxyCheckCmd `cmd:"" help:"Validate proxies against a target URL."`
}

type ProxyCheckCmd struct {
	Target  string `help:"Target URL." default:"https://www.google.com"`
	Timeout int    `help:"Timeout in seconds." default:"15"`
}

type ProxyCheckResult struct {
	Proxy     string `json:"proxy"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

func (p *ProxyCheckCmd) Run(ctx *Context) error {
	proxies, err := config.LoadProxies("")
	if err != nil {
		return err
	}
	if len(proxies) == 0 {
		return fmt.Errorf("no proxies configured")
	}

	results := make([]ProxyCheckResult, 0, len(proxies))
	for _, proxy := range proxies {
		result := ProxyCheckResult{Proxy: proxy}
		rotator, err := network.NewRotator([]string{proxy}, 5*time.Minute)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		client, err := network.NewClient(rotator)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		req, err := fhttp.NewRequest(fhttp.MethodGet, p.Target, nil)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		start := time.Now()
		resp, err := doWithTimeout(client, req, time.Duration(p.Timeout)*time.Second)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		_ = resp.Body.Close()

		result.LatencyMS = time.Since(start).Milliseconds()
		result.Status = fmt.Sprintf("%d", resp.StatusCode)
		results = append(results, result)
	}

	return writeProxyResults(ctx, results)
}

func doWithTimeout(client *network.Client, req *fhttp.Request, timeout time.Duration) (*fhttp.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), timeout)
	defer cancel()
	return client.Do(req.WithContext(ctx))
}

func writeProxyResults(ctx *Context, results []ProxyCheckResult) error {
	if ctx.JSONOutput {
		enc := json.NewEncoder(ctx.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if ctx.PlainText {
		for _, res := range results {
			line := []string{res.Proxy, res.Status, fmt.Sprintf("%d", res.LatencyMS), res.Error}
			fmt.Fprintln(ctx.Out, strings.Join(line, "\t"))
		}
		return nil
	}

	tw := tabwriter.NewWriter(ctx.Out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "proxy\tstatus\tlatency_ms\terror")
	for _, res := range results {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", res.Proxy, res.Status, res.LatencyMS, res.Error)
	}
	return tw.Flush()
}
