package wi

import (
	"context"
	"fmt"
)

// SessionClient is a scoped client for a single WhatsApp session.
// Obtain one via [Client.Session].
type SessionClient struct {
	id     string
	client *Client
}

func (s *SessionClient) url(path string) string {
	return fmt.Sprintf("/sessions/%s%s", s.id, path)
}

// ─── Session lifecycle ────────────────────────────────────────────────────────

// Status returns the connection state and phone number (if connected).
func (s *SessionClient) Status(ctx context.Context) (*SessionStatus, error) {
	var out SessionStatus
	return &out, s.client.get(ctx, s.url("/status"), &out)
}

// QR returns the current QR code as a base64-encoded PNG string.
// Returns an empty QR field when no QR is pending.
func (s *SessionClient) QR(ctx context.Context) (*QRCode, error) {
	var out QRCode
	return &out, s.client.get(ctx, s.url("/qr"), &out)
}

// Connect initiates the WhatsApp connection flow.
// Poll [SessionClient.Status] or [SessionClient.QR] for progress.
func (s *SessionClient) Connect(ctx context.Context) error {
	return s.client.post(ctx, s.url("/connect"), nil, nil)
}

// Disconnect closes the WhatsApp connection without removing session keys.
func (s *SessionClient) Disconnect(ctx context.Context) error {
	return s.client.post(ctx, s.url("/disconnect"), nil, nil)
}

// Logout disconnects and deletes session keys. The session must be re-paired.
func (s *SessionClient) Logout(ctx context.Context) error {
	return s.client.post(ctx, s.url("/logout"), nil, nil)
}

// PairPhone requests a pairing code for the given phone number.
// Returns an 8-character code the user enters on their phone.
func (s *SessionClient) PairPhone(ctx context.Context, phone string) (*PairPhoneResult, error) {
	var out PairPhoneResult
	return &out, s.client.post(ctx, s.url("/pairphone"), map[string]string{"phone": phone}, &out)
}

// ─── Send ─────────────────────────────────────────────────────────────────────

func (s *SessionClient) SendText(ctx context.Context, p SendTextParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-text"), p, &out)
}

func (s *SessionClient) SendImage(ctx context.Context, p SendImageParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-image"), p, &out)
}

func (s *SessionClient) SendAudio(ctx context.Context, p SendAudioParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-audio"), p, &out)
}

func (s *SessionClient) SendVideo(ctx context.Context, p SendVideoParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-video"), p, &out)
}

func (s *SessionClient) SendDocument(ctx context.Context, p SendDocumentParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-document"), p, &out)
}

func (s *SessionClient) SendLocation(ctx context.Context, p SendLocationParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-location"), p, &out)
}

func (s *SessionClient) SendContact(ctx context.Context, p SendContactParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-contact"), p, &out)
}

func (s *SessionClient) SendSticker(ctx context.Context, p SendStickerParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-sticker"), p, &out)
}

func (s *SessionClient) SendPoll(ctx context.Context, p SendPollParams) (*MessageResponse, error) {
	var out MessageResponse
	return &out, s.client.post(ctx, s.url("/messages/send-poll"), p, &out)
}

func (s *SessionClient) React(ctx context.Context, p SendReactionParams) error {
	return s.client.post(ctx, s.url("/chat/react"), p, nil)
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

func (s *SessionClient) MarkRead(ctx context.Context, p MarkReadParams) error {
	return s.client.post(ctx, s.url("/chat/markread"), p, nil)
}

func (s *SessionClient) Presence(ctx context.Context, p PresenceParams) error {
	return s.client.post(ctx, s.url("/chat/presence"), p, nil)
}
