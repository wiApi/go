package wi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const signatureHeader = "x-wi-signature"

// VerifySignature checks that the HMAC-SHA256 signature on a webhook payload
// matches the given secret. The signature format is "sha256=<hex>".
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

		sig := r.Header.Get(signatureHeader)
		if !VerifySignature(body, sig, secret) {
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
