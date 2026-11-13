package wi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// ─── VerifyWebhook (v2: signed timestamp) ─────────────────────────────────────

func generateSigV2(body []byte, secret string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.", ts)
	mac.Write(body)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifyWebhook_v2(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	const now = int64(1_800_000_000)
	h := func(v2 string) http.Header {
		hd := http.Header{}
		hd.Set("x-wi-signature-v2", v2)
		return hd
	}
	sig := generateSigV2(body, webhookSecret, now)

	if !VerifyWebhook(body, h(sig), webhookSecret, VerifyOptions{Now: now + 299}) {
		t.Error("fresh signature rejected")
	}
	if VerifyWebhook(body, h(sig), webhookSecret, VerifyOptions{Now: now + 301}) {
		t.Error("stale replay accepted")
	}
	if VerifyWebhook(body, h(sig), webhookSecret, VerifyOptions{Now: now - 301}) {
		t.Error("future timestamp beyond tolerance accepted")
	}
	if VerifyWebhook(body, h(sig), webhookSecret, VerifyOptions{Now: now + 60, ToleranceSeconds: 30}) {
		t.Error("custom tolerance ignored")
	}
	if !VerifyWebhook(body, h(sig), webhookSecret, VerifyOptions{Now: now + 60, ToleranceSeconds: 120}) {
		t.Error("custom tolerance rejected a valid delivery")
	}

	v1 := strings.SplitN(sig, ",", 2)[1]
	if VerifyWebhook(body, h(fmt.Sprintf("t=%d,%s", now+1, v1)), webhookSecret, VerifyOptions{Now: now}) {
		t.Error("tampered timestamp accepted")
	}
	if VerifyWebhook([]byte(`{"event":"x"}`), h(sig), webhookSecret, VerifyOptions{Now: now}) {
		t.Error("tampered body accepted")
	}
	if VerifyWebhook(body, h(sig), "wrong", VerifyOptions{Now: now}) {
		t.Error("wrong secret accepted")
	}
}

func TestVerifyWebhook_malformedV2NeverFallsBack(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	for _, v2 := range []string{"garbage", "t=abc,v1=00", "t=1800000000", "v1=00", "t=1800000000,v1=zz"} {
		hd := http.Header{}
		hd.Set("x-wi-signature-v2", v2)
		hd.Set("x-wi-signature", generateSig(body, webhookSecret))
		if VerifyWebhook(body, hd, webhookSecret, VerifyOptions{Now: 1_800_000_000}) {
			t.Errorf("malformed v2 %q accepted", v2)
		}
	}
}

func TestVerifyWebhook_legacyFallback(t *testing.T) {
	body := []byte(`{"event":"message"}`)
	hd := http.Header{}
	hd.Set("x-wi-signature", generateSig(body, webhookSecret))
	if !VerifyWebhook(body, hd, webhookSecret, VerifyOptions{}) {
		t.Error("legacy signature rejected when no v2 header")
	}
	if VerifyWebhook(body, http.Header{}, webhookSecret, VerifyOptions{}) {
		t.Error("unsigned request accepted")
	}
}

func TestWebhookHandler_rejectsStaleV2(t *testing.T) {
	body := []byte(`{"event":"message","session_id":"s","data":{},"timestamp":"x"}`)
	h := WebhookHandler(webhookSecret, func(*Event) error { return nil })
	do := func(ts int64) int {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set("x-wi-signature-v2", generateSigV2(body, webhookSecret, ts))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if got := do(time.Now().Unix()); got != http.StatusNoContent {
		t.Errorf("fresh = %d, want 204", got)
	}
	if got := do(time.Now().Unix() - 3600); got != http.StatusUnauthorized {
		t.Errorf("stale = %d, want 401", got)
	}
}

// ─── receipt ──────────────────────────────────────────────────────────────────

func TestParseEvent_receipt(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		status ReceiptStatus
		ids    int
	}{
		{"delivered", `{"event":"receipt","session_id":"s","data":{"type":"","rawType":"","status":"delivered","chat":"5511@s.whatsapp.net","sender":"5511@s.whatsapp.net","messageIds":["A","B"],"timestamp":1748790003,"isFromMe":false,"messageFromMe":true},"timestamp":"t"}`, ReceiptDelivered, 2},
		{"read", `{"event":"receipt","session_id":"s","data":{"type":"read","rawType":"read","status":"read","chat":"5511@s.whatsapp.net","sender":"5511@s.whatsapp.net","messageIds":["A"],"timestamp":1,"isFromMe":false,"messageFromMe":true},"timestamp":"t"}`, ReceiptRead, 1},
		{"played", `{"event":"receipt","session_id":"s","data":{"type":"played","rawType":"played","status":"played","chat":"5511@s.whatsapp.net","sender":"5511@s.whatsapp.net","messageIds":["C"],"timestamp":1,"isFromMe":false,"messageFromMe":true},"timestamp":"t"}`, ReceiptPlayed, 1},
		{"group", `{"event":"receipt","session_id":"s","data":{"type":"read","rawType":"read","status":"read","chat":"120363025246125486@g.us","sender":"5511912345678@s.whatsapp.net","messageIds":["G1"],"timestamp":1,"isFromMe":false,"messageFromMe":true},"timestamp":"t"}`, ReceiptRead, 1},
		{"retry has no status", `{"event":"receipt","session_id":"s","data":{"type":"retry","rawType":"retry","chat":"5511@s.whatsapp.net","sender":"5511@s.whatsapp.net","messageIds":["A"],"timestamp":1,"isFromMe":false,"messageFromMe":true},"timestamp":"t"}`, "", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := ParseEvent([]byte(tc.body))
			if err != nil || e.Event != EventReceipt {
				t.Fatalf("event = %+v, err = %v", e, err)
			}
			var r Receipt
			if err := json.Unmarshal(e.Data, &r); err != nil {
				t.Fatal(err)
			}
			if r.Status != tc.status || len(r.MessageIDs) != tc.ids || r.RawType != r.Type || !r.MessageFromMe {
				t.Errorf("receipt = %+v", r)
			}
		})
	}
}

// ─── disconnected ─────────────────────────────────────────────────────────────

func TestParseEvent_disconnected(t *testing.T) {
	cases := []struct {
		name   string
		data   string
		reason DisconnectReason
		will   bool
		resume string
		code   int
		retry  int
	}{
		{"transient", `{"reason":"transient","willReconnect":true}`, DisconnectTransient, true, "", 0, 0},
		{"logged_out", `{"reason":"logged_out","willReconnect":false,"resume":"connect","code":401,"detail":"401: logged out"}`, DisconnectLoggedOut, false, "connect", 401, 0},
		{"temporary_ban", `{"reason":"temporary_ban","willReconnect":false,"resume":"connect","code":101,"expireSeconds":3600,"retryAfterSeconds":3600}`, DisconnectTemporaryBan, false, "connect", 101, 3600},
		{"paused", `{"reason":"paused","willReconnect":false,"resume":"resume"}`, DisconnectPaused, false, "resume", 0, 0},
		{"a reason a newer server adds", `{"reason":"brand_new","willReconnect":true}`, DisconnectReason("brand_new"), true, "", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"event":"disconnected","session_id":"s","data":` + tc.data + `,"timestamp":"t"}`
			e, err := ParseEvent([]byte(body))
			if err != nil || e.Event != EventDisconnected {
				t.Fatalf("event = %+v, err = %v", e, err)
			}
			var d Disconnected
			if err := json.Unmarshal(e.Data, &d); err != nil {
				t.Fatalf("an unknown reason must still decode: %v", err)
			}
			if d.Reason != tc.reason || d.WillReconnect != tc.will || d.Resume != tc.resume || d.Code != tc.code || d.RetryAfterSeconds != tc.retry {
				t.Errorf("disconnected = %+v", d)
			}
		})
	}
}

func TestDisconnectReasons_closedSet(t *testing.T) {
	want := []DisconnectReason{"transient", "manual", "paused", "logged_out", "replaced", "temporary_ban", "client_outdated", "connect_failure"}
	got := DisconnectReasons()
	if len(got) != len(want) {
		t.Fatalf("DisconnectReasons() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("DisconnectReasons()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	for _, r := range want {
		if !r.Known() {
			t.Errorf("%q is not Known()", r)
		}
	}
	if DisconnectReason("brand_new").Known() {
		t.Error("an unknown reason reported Known()")
	}
}

func TestParseEvent_sessionMetadata(t *testing.T) {
	body := []byte(`{"event":"connected","session_id":"s","session_metadata":{"purpose":"briefing","team":"ops"},"data":{},"timestamp":"2026-09-03T00:00:00Z"}`)
	event, err := ParseEvent(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.SessionMetadata["purpose"] != "briefing" || event.SessionMetadata["team"] != "ops" || len(event.SessionMetadata) != 2 {
		t.Errorf("SessionMetadata = %v", event.SessionMetadata)
	}
}

func TestParseEvent_sessionMetadataAbsentOrEmpty(t *testing.T) {
	for name, body := range map[string]string{
		"older server": `{"event":"connected","session_id":"s","data":{},"timestamp":"t"}`,
		"no labels":    `{"event":"connected","session_id":"s","session_metadata":{},"data":{},"timestamp":"t"}`,
	} {
		t.Run(name, func(t *testing.T) {
			event, err := ParseEvent([]byte(body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := event.SessionMetadata["purpose"]; got != "" {
				t.Errorf("purpose = %q, want empty", got)
			}
		})
	}
}

// ─── pairing ──────────────────────────────────────────────────────────────────

func TestParseEvent_pairingFailed(t *testing.T) {
	cases := []struct {
		name   string
		data   string
		method PairingMethod
		reason PairingFailureReason
	}{
		{"an unread QR", `{"method":"qr","reason":"timeout","at":"2026-11-13T12:02:40Z"}`, PairingQR, PairingTimeout},
		{"a code that expired", `{"method":"code","reason":"timeout","at":"2026-11-13T12:02:40Z"}`, PairingCode, PairingTimeout},
		{"a refused link", `{"method":"qr","reason":"rejected","at":"2026-11-13T12:02:40Z"}`, PairingQR, PairingRejected},
		{"a cancelled attempt", `{"method":"qr","reason":"cancelled","at":"2026-11-13T12:02:40Z"}`, PairingQR, PairingCancelled},
		{"a reason a newer server adds", `{"method":"qr","reason":"brand_new","at":"2026-11-13T12:02:40Z"}`, PairingQR, PairingFailureReason("brand_new")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"event":"pairing.failed","session_id":"s","data":` + tc.data + `,"timestamp":"t"}`
			e, err := ParseEvent([]byte(body))
			if err != nil || e.Event != EventPairingFailed {
				t.Fatalf("event = %+v, err = %v", e, err)
			}
			var p PairingFailed
			if err := json.Unmarshal(e.Data, &p); err != nil {
				t.Fatalf("an unknown reason must still decode: %v", err)
			}
			if p.Method != tc.method || p.Reason != tc.reason || p.At != "2026-11-13T12:02:40Z" {
				t.Errorf("pairing.failed = %+v", p)
			}
		})
	}
}

func TestParseEvent_pairingSucceeded(t *testing.T) {
	body := `{"event":"pairing.succeeded","session_id":"s","data":{"method":"code","at":"2026-11-13T12:02:40Z"},"timestamp":"t"}`
	e, err := ParseEvent([]byte(body))
	if err != nil || e.Event != EventPairingSucceeded {
		t.Fatalf("event = %+v, err = %v", e, err)
	}
	var p PairingSucceeded
	if err := json.Unmarshal(e.Data, &p); err != nil {
		t.Fatal(err)
	}
	if p.Method != PairingCode || p.At != "2026-11-13T12:02:40Z" {
		t.Errorf("pairing.succeeded = %+v", p)
	}
}

func TestPairingFailureReasons_closedSet(t *testing.T) {
	want := []PairingFailureReason{"timeout", "rejected", "cancelled", "client_outdated", "error"}
	got := PairingFailureReasons()
	if len(got) != len(want) {
		t.Fatalf("PairingFailureReasons() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("PairingFailureReasons()[%d] = %q, want %q", i, got[i], want[i])
		}
		if !want[i].Known() {
			t.Errorf("%q is not Known()", want[i])
		}
	}
	if PairingFailureReason("brand_new").Known() {
		t.Error("an unknown reason reported Known()")
	}
}
