package wi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const webhookSecret = "test-webhook-secret-32-chars!!!"

func generateSig(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// ─── VerifySignature ──────────────────────────────────────────────────────────

func TestVerifySignature_valid(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	sig := generateSig(body, webhookSecret)
	if !VerifySignature(body, sig, webhookSecret) {
		t.Error("expected true for valid signature")
	}
}

func TestVerifySignature_invalid(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	zeros := strings.Repeat("0", 64)
	if VerifySignature(body, "sha256="+zeros, webhookSecret) {
		t.Error("expected false for invalid signature")
	}
}

func TestVerifySignature_tampered(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	sig := generateSig(body, webhookSecret)
	tampered := []byte(`{"event":"tampered"}`)
	if VerifySignature(tampered, sig, webhookSecret) {
		t.Error("expected false for tampered body")
	}
}

func TestVerifySignature_wrongScheme(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	validHex := hex.EncodeToString(mac.Sum(nil))
	// Correct hex, wrong prefix — should be rejected
	if VerifySignature(body, "sha512="+validHex, webhookSecret) {
		t.Error("expected false for wrong scheme prefix")
	}
}

func TestVerifySignature_empty(t *testing.T) {
	if VerifySignature([]byte(`{}`), "", webhookSecret) {
		t.Error("expected false for empty signature")
	}
}

func TestVerifySignature_noEquals(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	if VerifySignature(body, "sha256deadbeef", webhookSecret) {
		t.Error("expected false for signature with no '=' separator")
	}
}

func TestVerifySignature_wrongSecret(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	sig := generateSig(body, webhookSecret)
	if VerifySignature(body, sig, "wrong-secret") {
		t.Error("expected false for wrong secret")
	}
}

func TestVerifySignature_emptyBody(t *testing.T) {
	body := []byte{}
	sig := generateSig(body, webhookSecret)
	if !VerifySignature(body, sig, webhookSecret) {
		t.Error("expected true for valid empty-body signature")
	}
}

// ─── ParseEvent ───────────────────────────────────────────────────────────────

func TestParseEvent_valid(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"sess-1","data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	event, err := ParseEvent(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != EventMessage {
		t.Errorf("Event = %q, want %q", event.Event, EventMessage)
	}
	if event.SessionID != "sess-1" {
		t.Errorf("SessionID = %q, want sess-1", event.SessionID)
	}
	if event.Timestamp != "2026-09-03T00:00:00Z" {
		t.Errorf("Timestamp = %q", event.Timestamp)
	}
}

func TestParseEvent_allEventTypes(t *testing.T) {
	types := []EventType{
		EventMessage, EventMessageSent, EventMessageDeleted,
		EventMessageReaction, EventConnected, EventDisconnected,
		EventQR, EventCallIncoming, EventCallAccepted, EventCallRejected,
	}
	for _, et := range types {
		t.Run(string(et), func(t *testing.T) {
			body := fmt.Appendf(nil, `{"event":%q,"session_id":"s","data":{},"timestamp":"2026-09-03T00:00:00Z"}`, et)
			event, err := ParseEvent(body)
			if err != nil {
				t.Fatalf("ParseEvent error: %v", err)
			}
			if event.Event != et {
				t.Errorf("Event = %q, want %q", event.Event, et)
			}
		})
	}
}

func TestParseEvent_invalidJSON(t *testing.T) {
	_, err := ParseEvent([]byte(`not-json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseEvent_emptyObject(t *testing.T) {
	// Valid JSON but no fields — should not error, just zero values
	event, err := ParseEvent([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Event != "" {
		t.Errorf("Event = %q, want empty", event.Event)
	}
}

// ─── WebhookHandler ───────────────────────────────────────────────────────────

func TestWebhookHandler_valid(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"s","data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	sig := generateSig(body, webhookSecret)

	var called bool
	var receivedEvent *Event
	handler := WebhookHandler(webhookSecret, func(e *Event) error {
		called = true
		receivedEvent = e
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(signatureHeader, sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if !called {
		t.Error("handler fn was not called")
	}
	if receivedEvent == nil || receivedEvent.Event != EventMessage {
		t.Errorf("received event = %v", receivedEvent)
	}
}

func TestWebhookHandler_invalidSig(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"s","data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	handler := WebhookHandler(webhookSecret, func(e *Event) error { return nil })

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(signatureHeader, "sha256=bad")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestWebhookHandler_missingSig(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"s","data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	handler := WebhookHandler(webhookSecret, func(e *Event) error { return nil })

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	// No signature header set
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestWebhookHandler_handlerError(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"s","data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	sig := generateSig(body, webhookSecret)
	handler := WebhookHandler(webhookSecret, func(e *Event) error {
		return fmt.Errorf("handler error")
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(signatureHeader, sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestWebhookHandler_invalidJSON(t *testing.T) {
	body := []byte(`not-json`)
	sig := generateSig(body, webhookSecret)
	handler := WebhookHandler(webhookSecret, func(e *Event) error { return nil })

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(signatureHeader, sig)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ─── Error ────────────────────────────────────────────────────────────────────

func TestError_format(t *testing.T) {
	e := &Error{Status: 400, Message: "Bad input"}
	got := e.Error()
	if !strings.Contains(got, "400") || !strings.Contains(got, "Bad input") {
		t.Errorf("Error() = %q", got)
	}
	// No code: should not contain parentheses
	if strings.Contains(got, "(") {
		t.Errorf("Error() should not contain code parentheses, got %q", got)
	}
}

func TestError_withCode(t *testing.T) {
	e := &Error{Status: 429, Message: "Rate limited", Code: "RATE_LIMIT"}
	got := e.Error()
	if !strings.Contains(got, "429") {
		t.Errorf("Error() missing status, got %q", got)
	}
	if !strings.Contains(got, "Rate limited") {
		t.Errorf("Error() missing message, got %q", got)
	}
	if !strings.Contains(got, "RATE_LIMIT") {
		t.Errorf("Error() missing code, got %q", got)
	}
}

func TestError_implementsError(t *testing.T) {
	var _ error = &Error{}
}
