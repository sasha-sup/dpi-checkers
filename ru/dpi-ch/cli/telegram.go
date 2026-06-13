package cli

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Telegram credentials are read from a .env file (BOT_TOKEN / CHAT_ID).
// Default path is "dpich.env" next to the binary; override with DPICH_TG_ENV.
const defaultTgEnvPath = "dpich.env"

// tgDispatch sends each message to Telegram, loading credentials once.
// Silent no-op (with a stderr note) if there are no messages or no creds.
func tgDispatch(msgs []string) {
	if len(msgs) == 0 {
		return
	}

	path := os.Getenv("DPICH_TG_ENV")
	if path == "" {
		path = defaultTgEnvPath
	}

	env, err := loadEnv(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli: telegram env read failed (%s): %v\n", path, err)
		return
	}
	token, chat := env["BOT_TOKEN"], env["CHAT_ID"]
	if token == "" || chat == "" {
		fmt.Fprintln(os.Stderr, "cli: BOT_TOKEN/CHAT_ID missing, skip telegram")
		return
	}

	for _, m := range msgs {
		if err := sendTelegram(token, chat, m); err != nil {
			fmt.Fprintf(os.Stderr, "cli: telegram send failed: %v\n", err)
		}
	}
}

// formatDown renders the 🔴 alert for targets that just went down.
func formatDown(down []Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🔴 dpi-ch: %d target(s) down\n", len(down))
	for _, r := range down {
		reason := r.Alive
		if reason == "ok" {
			reason = "siberian:" + r.Siberian
		}
		fmt.Fprintf(&b, "\n• %s / %s (%s:%d)\n  %s",
			r.Section, r.Target, r.Ip, r.Port, reason)
	}
	return b.String()
}

// formatRecovered renders the 🟢 message for targets back up.
func formatRecovered(keys []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🟢 dpi-ch: %d target(s) recovered\n", len(keys))
	for _, k := range keys {
		fmt.Fprintf(&b, "\n• %s", k)
	}
	return b.String()
}

// formatHeartbeat renders the daily ✅ "all alive" confirmation.
func formatHeartbeat(results []Result) string {
	return fmt.Sprintf("✅ dpi-ch: all %d target(s) alive", len(results))
}

func sendTelegram(token, chat, text string) error {
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.PostForm(api, url.Values{
		"chat_id": {chat},
		"text":    {text},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api status %d", resp.StatusCode)
	}
	return nil
}

// loadEnv parses a minimal KEY=VALUE .env file.
func loadEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		env[k] = v
	}
	return env, sc.Err()
}
