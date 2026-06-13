// Package cli provides a headless (no-TUI) run mode for dpi-ch.
//
// It runs the webhost checker once over every target in the config,
// prints one JSON object per target to stdout (JSONL), and exits with
// a non-zero status if any target is down or blocked. On failure it can
// also push a summary to Telegram (see telegram.go).
//
// Intended for scheduled runs (cron / systemd timer) to monitor a VPN
// server's REALITY endpoints.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/hyperion-cs/dpi-checkers/ru/dpi-ch/checkers"
	"github.com/hyperion-cs/dpi-checkers/ru/dpi-ch/config"
	"github.com/hyperion-cs/dpi-checkers/ru/dpi-ch/inetlookup"
	"github.com/hyperion-cs/dpi-checkers/ru/dpi-ch/updater"
)

// Result is the JSON shape emitted per checked target.
type Result struct {
	Ts       string `json:"ts"`
	Section  string `json:"section"`
	Target   string `json:"target"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Asn      int32  `json:"asn"`
	Org      string `json:"org"`
	Country  string `json:"country"`
	Subnet   string `json:"subnet"`
	Sni      string `json:"sni"`
	TlsV     string `json:"tls"`
	Alive    string `json:"alive"`
	Tcp1620  string `json:"tcp1620"`
	Siberian string `json:"siberian"`
	Ok       bool   `json:"ok"`
}

// Cli is the headless entry point (invoked from main on --ui cli).
func Cli() {
	ctx := context.Background()

	// Refresh the geolite (asn/org/country/cidr) data on the same schedule the
	// TUI uses, so a cron-driven run keeps its lookup db current. Downloads on
	// first run if missing. Self-update is intentionally skipped here.
	ensureInetlookupData(ctx)

	// Warm up the inetlookup data, same as the TUI Init.
	inetlookup.Default()
	sections := config.Get().Checkers.Webhost.Sections

	results := []Result{}
	for _, sec := range sections {
		if len(sec.Targets) == 0 {
			continue
		}
		results = append(results, runSection(ctx, sec)...)
	}

	// Emit JSONL to stdout.
	enc := json.NewEncoder(os.Stdout)
	failed := []Result{}
	for _, r := range results {
		_ = enc.Encode(r)
		if !r.Ok {
			failed = append(failed, r)
		}
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "cli: no targets in config")
		os.Exit(2)
	}

	handleNotifications(results, failed)

	if len(failed) > 0 {
		os.Exit(1)
	}
}

// handleNotifications turns the current vs. previous run into Telegram messages:
// 🔴 on targets that just went down, 🟢 on targets that recovered, and a daily
// ✅ heartbeat when everything is alive. State is persisted for the next run.
func handleNotifications(results, failed []Result) {
	prev := loadState()
	prevDown := map[string]struct{}{}
	for _, k := range prev.Down {
		prevDown[k] = struct{}{}
	}

	curDownKeys := []string{}
	curDownSet := map[string]struct{}{}
	for _, r := range failed {
		k := key(r)
		curDownKeys = append(curDownKeys, k)
		curDownSet[k] = struct{}{}
	}

	// Newly down: failing now, was not failing before -> 🔴 (once, on transition).
	newlyDown := []Result{}
	for _, r := range failed {
		if _, was := prevDown[key(r)]; !was {
			newlyDown = append(newlyDown, r)
		}
	}

	// Recovered: was down before, ok now -> 🟢.
	recovered := []string{}
	for k := range prevDown {
		if _, still := curDownSet[k]; !still {
			recovered = append(recovered, k)
		}
	}
	sort.Strings(recovered)

	msgs := []string{}
	if len(newlyDown) > 0 {
		msgs = append(msgs, formatDown(newlyDown))
	}
	if len(recovered) > 0 {
		msgs = append(msgs, formatRecovered(recovered))
	}

	today := nowDate()
	hbDate := prev.HeartbeatDate
	if len(failed) == 0 && hbDate != today {
		msgs = append(msgs, formatHeartbeat(results))
		hbDate = today
	}

	tgDispatch(msgs)
	saveState(state{Down: curDownKeys, HeartbeatDate: hbDate})
}

// key is the stable identifier for a target across runs.
func key(r Result) string {
	return r.Section + "/" + r.Target
}

func nowDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// ensureInetlookupData downloads/refreshes the geolite data if the updater is
// enabled and due (or forced via --force-inetlookup-update). Errors are logged
// to stderr but not fatal — Default() will panic later if data is truly absent.
func ensureInetlookupData(ctx context.Context) {
	uc := config.Get().Updater
	due, _ := updater.TimeToUpdate(uc.InetlookupTsFile)
	if !uc.ForceInetlookupUpdate && !(uc.Enabled && due) {
		return
	}
	if err := updater.GeoliteUpdate(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "cli: geolite update failed:", err)
	}
}

// runSection runs the existing webhost pipeline for one config section
// and maps each raw result into a Result tagged with the section name.
func runSection(ctx context.Context, sec config.WebhostSection) []Result {
	out := checkers.WebhostGochanRunner(checkers.WebhostGochanRunnerOpt{
		Ctx:     ctx,
		Targets: sec.Targets,
	})

	results := []Result{}
	// Drain both channels until closed (same pattern as the TUI consumer).
	for out.Out != nil || out.Progress != nil {
		select {
		case v, ok := <-out.Out:
			if !ok {
				out.Out = nil
				continue
			}
			results = append(results, mapResult(sec.Name, v))
		case p, ok := <-out.Progress:
			if !ok {
				out.Progress = nil
				continue
			}
			if config.Get().Debug {
				fmt.Fprintln(os.Stderr, p)
			}
		}
	}
	return results
}

func mapResult(section string, v checkers.WebhostGochanOut[checkers.WebhostGochanBag]) Result {
	r := v.Out
	// A target is healthy if it answered (Alive) and is not under a
	// "siberian" DPI block. Throughput (tcp1620) is reported but not fatal.
	ok := r.Alive == nil && !isProblem(r.Siberian)

	return Result{
		Ts:       nowRFC3339(),
		Section:  section,
		Target:   v.Bag.Name,
		Ip:       r.IpInfo.Ip.String(),
		Port:     r.Port,
		Asn:      r.IpInfo.Asn,
		Org:      r.IpInfo.Org,
		Country:  r.IpInfo.CountryIso,
		Subnet:   r.IpInfo.Subnet.String(),
		Sni:      r.Sni,
		TlsV:     tlsVersion(r.TlsV),
		Alive:    status(r.Alive),
		Tcp1620:  status(r.Tcp1620),
		Siberian: status(r.Siberian),
		Ok:       ok,
	}
}

// status renders a check error as a short string for JSON.
func status(err error) string {
	switch {
	case err == nil:
		return "ok"
	case err == checkers.ErrWebhostSkip:
		return "skip"
	default:
		return err.Error()
	}
}

// isProblem reports whether err is a real failure (not nil, not a skip).
func isProblem(err error) bool {
	return err != nil && err != checkers.ErrWebhostSkip
}

func tlsVersion(v uint16) string {
	switch v {
	case 0x0304:
		return "1.3"
	case 0x0303:
		return "1.2"
	case 0x0302:
		return "1.1"
	case 0x0301:
		return "1.0"
	case 0:
		return ""
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
