package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/config"
	"github.com/amirsober-cloud/sober-tunnel/engines/camouflage/internal/ingress"
	"github.com/fatih/color"
)

// cmdProbe tests every endpoint (mimic) of a node individually and reports which
// are reachable and which are blocked — the fast way to see, per node, which
// camouflage protocols actually work on the current network path.
func cmdProbe(args []string) {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)
	nodeAlias := fs.String("node", "", "node alias (default: all nodes)")
	samples := fs.Int("samples", 1, "independent connection attempts per endpoint (1-5)")
	_ = fs.Parse(args)
	if *samples < 1 || *samples > 5 {
		fail("samples must be between 1 and 5")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fail("no config: %v", err)
	}
	if cfg.Role != "iran" {
		fail("probe runs on the Iran hub")
	}

	nodes := cfg.ForeignNodes
	if *nodeAlias != "" {
		nodes = nil
		for _, n := range cfg.ForeignNodes {
			if n.Alias == *nodeAlias {
				nodes = append(nodes, n)
			}
		}
		if len(nodes) == 0 {
			fail("node %q not found", *nodeAlias)
		}
	}
	if len(nodes) == 0 {
		fail("no foreign nodes configured")
	}

	anyFail := false
	for _, n := range nodes {
		fmt.Printf("\nNode %q (%d endpoints):\n", n.Alias, len(n.Endpoints))
		for _, ep := range n.Endpoints {
			if *samples > 1 {
				if !probeSamples(ep, n.AuthToken, *samples) {
					anyFail = true
				}
				continue
			}
			rtt, err := ingress.ProbeEndpoint(ep, n.AuthToken)
			if err != nil {
				anyFail = true
				color.Red("  ✗ %-6s %-24s FAIL: %v", ep.Mimic, ep.Target, err)
			} else {
				color.Green("  ✓ %-6s %-24s OK  (%.0f ms)", ep.Mimic, ep.Target, float64(rtt.Microseconds())/1000)
			}
		}
	}
	if anyFail {
		// Non-zero exit so scripts/CI can detect a degraded path.
		fail("one or more endpoint probes failed")
	}
}

func probeSamples(ep config.Endpoint, token string, samples int) bool {
	var total time.Duration
	var successes int
	for attempt := 1; attempt <= samples; attempt++ {
		rtt, err := ingress.ProbeEndpoint(ep, token)
		if err != nil {
			color.Red("  %-6s %-24s attempt %d/%d: %v", ep.Mimic, ep.Target, attempt, samples, err)
			continue
		}
		total += rtt
		successes++
	}
	fmt.Printf("  %-6s %-24s %d/%d attempts succeeded", ep.Mimic, ep.Target, successes, samples)
	if successes > 0 {
		fmt.Printf(" (mean mux RTT %.1f ms)", float64(total)/float64(time.Millisecond)/float64(successes))
	}
	fmt.Println()
	return successes == samples
}
