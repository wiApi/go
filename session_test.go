package wi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// capturedReq holds the last request seen by the test server.
type capturedReq struct {
	method string
	path   string
	header http.Header
	body   map[string]any
}

func newTestServer(t *testing.T, status int, respBody string) (*httptest.Server, *capturedReq) {
	t.Helper()
	var cap capturedReq
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.method = r.Method
		cap.path = r.URL.Path
		cap.header = r.Header.Clone()
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			if len(b) > 0 {
				_ = json.Unmarshal(b, &cap.body)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(ts.Close)
	return ts, &cap
}

func testClient(t *testing.T, status int, body string) (*Client, *capturedReq) {
	t.Helper()
	ts, cap := newTestServer(t, status, body)
	c := New("test-api-key", WithBaseURL(ts.URL))
	return c, cap
}

// ─── SendText ─────────────────────────────────────────────────────────────────

func TestSendText(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"msg_1","id":"msg_1"}`)
	ctx := context.Background()

	msg, err := c.Session("sess-1").SendText(ctx, SendTextParams{
		To:   "5511999999999",
		Text: "Hello!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.method != http.MethodPost {
		t.Errorf("method = %q, want POST", cap.method)
	}
	if cap.path != "/sessions/sess-1/messages/send-text" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.header.Get("x-api-key") != "test-api-key" {
		t.Error("x-api-key header missing or wrong")
	}
	if cap.body["to"] != "5511999999999" {
		t.Errorf("body.to = %v", cap.body["to"])
	}
	if cap.body["text"] != "Hello!" {
		t.Errorf("body.text = %v", cap.body["text"])
	}
	if msg.MessageID != "msg_1" {
		t.Errorf("MessageID = %q", msg.MessageID)
	}
}

func TestSendText_quotedAndMentions(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m","id":"m"}`)
	_, err := c.Session("s").SendText(context.Background(), SendTextParams{
		To:       "55",
		Text:     "hi",
		Quoted:   "quoted-id",
		Mentions: []string{"5511111111111"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.body["quoted"] != "quoted-id" {
		t.Errorf("quoted = %v", cap.body["quoted"])
	}
	mentions, ok := cap.body["mentions"].([]any)
	if !ok || len(mentions) != 1 {
		t.Errorf("mentions = %v", cap.body["mentions"])
	}
}

func TestSendText_4xx(t *testing.T) {
	c, _ := testClient(t, 401, `{"message":"Unauthorized"}`)
	_, err := c.Session("s").SendText(context.Background(), SendTextParams{To: "55", Text: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	wiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if wiErr.Status != 401 {
		t.Errorf("Status = %d, want 401", wiErr.Status)
	}
	if wiErr.Message != "Unauthorized" {
		t.Errorf("Message = %q, want Unauthorized", wiErr.Message)
	}
}

func TestSendText_5xx(t *testing.T) {
	c, _ := testClient(t, 500, `{"message":"Internal Server Error","code":"INTERNAL"}`)
	_, err := c.Session("s").SendText(context.Background(), SendTextParams{To: "55", Text: "hi"})
	if err == nil {
		t.Fatal("expected error on 500")
	}
	wiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if wiErr.Status != 500 {
		t.Errorf("Status = %d, want 500", wiErr.Status)
	}
	if wiErr.Code != "INTERNAL" {
		t.Errorf("Code = %q, want INTERNAL", wiErr.Code)
	}
}

// ─── SendImage ────────────────────────────────────────────────────────────────

func TestSendImage(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m2","id":"m2"}`)
	msg, err := c.Session("s").SendImage(context.Background(), SendImageParams{
		To:      "55",
		URL:     "https://img.com/a.jpg",
		Caption: "look",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-image" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["caption"] != "look" {
		t.Errorf("caption = %v", cap.body["caption"])
	}
	if cap.body["url"] != "https://img.com/a.jpg" {
		t.Errorf("url = %v", cap.body["url"])
	}
	if msg.MessageID != "m2" {
		t.Errorf("MessageID = %q", msg.MessageID)
	}
}

// ─── SendAudio ────────────────────────────────────────────────────────────────

func TestSendAudio_ptt(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m3","id":"m3"}`)
	_, err := c.Session("s").SendAudio(context.Background(), SendAudioParams{
		To:  "55",
		URL: "https://cdn/a.ogg",
		PTT: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-audio" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["ptt"] != true {
		t.Errorf("ptt = %v, want true", cap.body["ptt"])
	}
}

func TestSendAudio_noPtt(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m3","id":"m3"}`)
	_, err := c.Session("s").SendAudio(context.Background(), SendAudioParams{
		To:  "55",
		URL: "https://cdn/a.ogg",
		PTT: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	// ptt omitempty: false should be absent from JSON body
	if _, present := cap.body["ptt"]; present {
		t.Error("ptt should be omitted when false")
	}
}

// ─── SendVideo ────────────────────────────────────────────────────────────────

func TestSendVideo(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"mv","id":"mv"}`)
	_, err := c.Session("s").SendVideo(context.Background(), SendVideoParams{
		To:      "55",
		URL:     "https://cdn/v.mp4",
		Caption: "watch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-video" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["caption"] != "watch" {
		t.Errorf("caption = %v", cap.body["caption"])
	}
}

// ─── SendDocument ─────────────────────────────────────────────────────────────

func TestSendDocument(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m4","id":"m4"}`)
	_, err := c.Session("s").SendDocument(context.Background(), SendDocumentParams{
		To:       "55",
		URL:      "https://cdn/doc.pdf",
		Filename: "report.pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-document" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["filename"] != "report.pdf" {
		t.Errorf("filename = %v", cap.body["filename"])
	}
}

// ─── SendLocation ─────────────────────────────────────────────────────────────

func TestSendLocation(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m5","id":"m5"}`)
	_, err := c.Session("s").SendLocation(context.Background(), SendLocationParams{
		To:        "55",
		Latitude:  -23.5505,
		Longitude: -46.6333,
		Title:     "SP",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-location" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["latitude"] == nil {
		t.Error("latitude missing from body")
	}
	if cap.body["title"] != "SP" {
		t.Errorf("title = %v", cap.body["title"])
	}
}

// ─── SendContact ──────────────────────────────────────────────────────────────

func TestSendContact(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"mc","id":"mc"}`)
	_, err := c.Session("s").SendContact(context.Background(), SendContactParams{
		To:            "55",
		ContactName:   "Alice",
		ContactNumber: "5511888888888",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-contact" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["contactName"] != "Alice" {
		t.Errorf("contactName = %v", cap.body["contactName"])
	}
	if cap.body["contactNumber"] != "5511888888888" {
		t.Errorf("contactNumber = %v", cap.body["contactNumber"])
	}
}

// ─── SendSticker ──────────────────────────────────────────────────────────────

func TestSendSticker(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"ms","id":"ms"}`)
	_, err := c.Session("s").SendSticker(context.Background(), SendStickerParams{
		To:  "55",
		URL: "https://cdn/sticker.webp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-sticker" {
		t.Errorf("path = %q", cap.path)
	}
}

// ─── SendPoll ─────────────────────────────────────────────────────────────────

func TestSendPoll(t *testing.T) {
	c, cap := testClient(t, 200, `{"messageId":"m6","id":"m6"}`)
	_, err := c.Session("s").SendPoll(context.Background(), SendPollParams{
		To:       "55",
		Question: "Best?",
		Options:  []string{"A", "B", "C"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/messages/send-poll" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["question"] != "Best?" {
		t.Errorf("question = %v", cap.body["question"])
	}
	opts, ok := cap.body["options"].([]any)
	if !ok || len(opts) != 3 {
		t.Errorf("options = %v", cap.body["options"])
	}
}

// ─── React ────────────────────────────────────────────────────────────────────

func TestReact(t *testing.T) {
	c, cap := testClient(t, 204, ``)
	err := c.Session("s").React(context.Background(), SendReactionParams{
		To:        "55",
		MessageID: "msg_x",
		Emoji:     "\U0001f44d",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/chat/react" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["messageId"] != "msg_x" {
		t.Errorf("messageId = %v", cap.body["messageId"])
	}
}

// ─── MarkRead ─────────────────────────────────────────────────────────────────

func TestMarkRead(t *testing.T) {
	c, cap := testClient(t, 204, ``)
	err := c.Session("s").MarkRead(context.Background(), MarkReadParams{
		ChatID: "55@s.whatsapp.net",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/chat/markread" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["chatId"] != "55@s.whatsapp.net" {
		t.Errorf("chatId = %v", cap.body["chatId"])
	}
}

// ─── Presence ─────────────────────────────────────────────────────────────────

func TestPresence(t *testing.T) {
	c, cap := testClient(t, 204, ``)
	err := c.Session("s").Presence(context.Background(), PresenceParams{
		ChatID:   "55@s.whatsapp.net",
		Presence: PresenceComposing,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/sessions/s/chat/presence" {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["presence"] != "composing" {
		t.Errorf("presence = %v", cap.body["presence"])
	}
	if cap.body["chatId"] != "55@s.whatsapp.net" {
		t.Errorf("chatId = %v", cap.body["chatId"])
	}
}

func TestPresence_recording(t *testing.T) {
	c, cap := testClient(t, 204, ``)
	err := c.Session("s").Presence(context.Background(), PresenceParams{
		ChatID:   "55@s.whatsapp.net",
		Presence: PresenceRecording,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.body["presence"] != "recording" {
		t.Errorf("presence = %v", cap.body["presence"])
	}
}

// ─── Status ───────────────────────────────────────────────────────────────────

func TestStatus(t *testing.T) {
	c, cap := testClient(t, 200, `{"connected":true,"phone":"5511999"}`)
	status, err := c.Session("abc").Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.method != http.MethodGet {
		t.Errorf("method = %q, want GET", cap.method)
	}
	if cap.path != "/sessions/abc/status" {
		t.Errorf("path = %q", cap.path)
	}
	if !status.Connected {
		t.Error("Connected should be true")
	}
	if status.Phone != "5511999" {
		t.Errorf("Phone = %q", status.Phone)
	}
}

func TestStatus_disconnected(t *testing.T) {
	c, _ := testClient(t, 200, `{"connected":false}`)
	status, err := c.Session("s").Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Connected {
		t.Error("Connected should be false")
	}
	if status.Phone != "" {
		t.Errorf("Phone should be empty, got %q", status.Phone)
	}
}

// ─── QR ───────────────────────────────────────────────────────────────────────

func TestQR(t *testing.T) {
	c, cap := testClient(t, 200, `{"qr":"data:image/png;base64,abc"}`)
	qr, err := c.Session("s").QR(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.method != http.MethodGet {
		t.Errorf("method = %q, want GET", cap.method)
	}
	if cap.path != "/sessions/s/qr" {
		t.Errorf("path = %q", cap.path)
	}
	if qr.QR != "data:image/png;base64,abc" {
		t.Errorf("QR = %q", qr.QR)
	}
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────

func TestConnect(t *testing.T) {
	c, cap := testClient(t, 200, `{}`)
	err := c.Session("s").Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.method != http.MethodPost {
		t.Errorf("method = %q, want POST", cap.method)
	}
	if !strings.HasSuffix(cap.path, "/connect") {
		t.Errorf("path = %q", cap.path)
	}
}

func TestDisconnect(t *testing.T) {
	c, cap := testClient(t, 200, `{}`)
	err := c.Session("s").Disconnect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(cap.path, "/disconnect") {
		t.Errorf("path = %q", cap.path)
	}
}

func TestLogout(t *testing.T) {
	c, cap := testClient(t, 200, `{}`)
	err := c.Session("s").Logout(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(cap.path, "/logout") {
		t.Errorf("path = %q", cap.path)
	}
}

// ─── PairPhone ────────────────────────────────────────────────────────────────

func TestPairPhone(t *testing.T) {
	c, cap := testClient(t, 200, `{"pairCode":"ABCD1234"}`)
	result, err := c.Session("s").PairPhone(context.Background(), "5511999999999")
	if err != nil {
		t.Fatal(err)
	}
	if cap.method != http.MethodPost {
		t.Errorf("method = %q, want POST", cap.method)
	}
	if !strings.HasSuffix(cap.path, "/pairphone") {
		t.Errorf("path = %q", cap.path)
	}
	if cap.body["phone"] != "5511999999999" {
		t.Errorf("phone = %v", cap.body["phone"])
	}
	if result.PairCode != "ABCD1234" {
		t.Errorf("PairCode = %q", result.PairCode)
	}
}

// ─── API key header ───────────────────────────────────────────────────────────

func TestAPIKeyHeader_allRequests(t *testing.T) {
	for _, name := range []string{"GET", "POST"} {
		t.Run(name, func(t *testing.T) {
			c, cap := testClient(t, 200, `{"connected":true}`)
			var err error
			if name == "GET" {
				_, err = c.Session("s").Status(context.Background())
			} else {
				err = c.Session("s").Connect(context.Background())
			}
			if err != nil {
				t.Fatal(err)
			}
			if cap.header.Get("x-api-key") != "test-api-key" {
				t.Errorf("x-api-key header missing on %s request", name)
			}
		})
	}
}
