package wi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	signatureHeader   = "x-wi-signature"
	signatureV2Header = "x-wi-signature-v2"

	// DefaultToleranceSeconds is the default max age of x-wi-timestamp (5 minutes).
	DefaultToleranceSeconds = 300
)

// VerifyOptions tunes [VerifyWebhook]. The zero value means "5 minute tolerance, real clock".
type VerifyOptions struct {
	// ToleranceSeconds is the max distance between the signed timestamp and now.
	// Zero means [DefaultToleranceSeconds].
	ToleranceSeconds int
	// Now overrides the clock (unix seconds); for tests. Zero means time.Now().
	Now int64
}

// VerifyWebhook verifies a webhook delivery with replay protection.
//
// It checks "x-wi-signature-v2: t=<unix seconds>,v1=<hex>", the HMAC-SHA256 of
// "<t>.<body>", and rejects it when t is further than the tolerance from now: a
// captured request cannot be replayed later, and a tampered timestamp fails the
// signature. A delivery with no v2 header (older servers) falls back to the
// legacy "x-wi-signature" body signature, which has no replay protection. A
// malformed v2 header is rejected, never downgraded.
func VerifyWebhook(body []byte, header http.Header, secret string, opts VerifyOptions) bool {
	v2 := header.Get(signatureV2Header)
	if v2 == "" {
		legacy := header.Get(signatureHeader)
		return legacy != "" && VerifySignature(body, legacy, secret)
	}

	var ts, v1 string
	for _, part := range strings.Split(v2, ",") {
		k, v, _ := strings.Cut(strings.TrimSpace(part), "=")
		switch k {
		case "t":
			ts = v
		case "v1":
			v1 = v
		}
	}
	t, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || v1 == "" {
		return false
	}
	got, err := hex.DecodeString(v1)
	if err != nil {
		return false
	}

	tolerance := int64(opts.ToleranceSeconds)
	if tolerance == 0 {
		tolerance = DefaultToleranceSeconds
	}
	now := opts.Now
	if now == 0 {
		now = time.Now().Unix()
	}
	if d := now - t; d > tolerance || -d > tolerance {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), got)
}

// VerifySignature checks the legacy body-only HMAC-SHA256 signature
// ("sha256=<hex>", no replay protection). Prefer [VerifyWebhook].
//
// Returns false (not an error) when the signature is wrong — that is the
// expected outcome for invalid requests; treat it as HTTP 401.
func VerifySignature(body []byte, signature, secret string) bool {
	before, after, found := strings.Cut(signature, "=")
	if !found || before != "sha256" {
		return false
	}
	expected, err := hex.DecodeString(after)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), expected)
}

// ParseEvent decodes a raw webhook body into an [Event].
func ParseEvent(body []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, fmt.Errorf("wi: parse event: %w", err)
	}
	return &e, nil
}

// WebhookHandlerFunc handles a verified, parsed webhook event.
type WebhookHandlerFunc func(event *Event) error

// WebhookHandler returns an [http.Handler] that verifies signatures and
// dispatches events to fn.
//
//	mux.Handle("/webhook", wi.WebhookHandler(secret, func(e *wi.Event) error {
//	    log.Println(e.Event, e.SessionID)
//	    return nil
//	}))
func WebhookHandler(secret string, fn WebhookHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		if !VerifyWebhook(body, r.Header, secret, VerifyOptions{}) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		event, err := ParseEvent(body)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := fn(event); err != nil {
			http.Error(w, "handler error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
