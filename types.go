package wi

import "encoding/json"

// ─── Session ──────────────────────────────────────────────────────────────────

// SessionStatus is returned by [SessionClient.Status].
type SessionStatus struct {
	Connected bool   `json:"connected"`
	Phone     string `json:"phone,omitempty"`
}

// QRCode is returned by [SessionClient.QR].
type QRCode struct {
	// Base64-encoded PNG. Empty when no QR is pending.
	QR string `json:"qr,omitempty"`
}

// PairPhoneResult is returned by [SessionClient.PairPhone].
type PairPhoneResult struct {
	PairCode string `json:"pairCode,omitempty"`
}

// MessageResponse is returned by all Send* methods.
type MessageResponse struct {
	MessageID string `json:"messageId"`
	ID        string `json:"id"`
}

// ─── Send params ──────────────────────────────────────────────────────────────

// SendTextParams are the parameters for [SessionClient.SendText].
type SendTextParams struct {
	// Phone number or JID, e.g. "5511999999999" or "5511999999999@s.whatsapp.net"
	To       string   `json:"to"`
	Text     string   `json:"text"`
	Quoted   string   `json:"quoted,omitempty"`
	Mentions []string `json:"mentions,omitempty"`
}

type SendImageParams struct {
	To      string `json:"to"`
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Quoted  string `json:"quoted,omitempty"`
}

type SendAudioParams struct {
	To  string `json:"to"`
	URL string `json:"url"`
	// PTT sends as a voice note. Default: false.
	PTT bool `json:"ptt,omitempty"`
}

type SendVideoParams struct {
	To      string `json:"to"`
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Quoted  string `json:"quoted,omitempty"`
}

type SendDocumentParams struct {
	To       string `json:"to"`
	URL      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type SendLocationParams struct {
	To        string  `json:"to"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     string  `json:"title,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type SendContactParams struct {
	To            string `json:"to"`
	ContactName   string `json:"contactName"`
	ContactNumber string `json:"contactNumber"`
}

type SendStickerParams struct {
	To  string `json:"to"`
	URL string `json:"url"`
}

type SendPollParams struct {
	To              string   `json:"to"`
	Question        string   `json:"question"`
	Options         []string `json:"options"`
	MultipleAnswers bool     `json:"multipleAnswers,omitempty"`
}

type SendReactionParams struct {
	To        string `json:"to"`
	MessageID string `json:"messageId"`
	// Emoji character, or empty string to remove the reaction.
	Emoji string `json:"emoji"`
}

// ─── Chat params ──────────────────────────────────────────────────────────────

type MarkReadParams struct {
	ChatID        string `json:"chatId"`
	LastMessageID string `json:"lastMessageId,omitempty"`
}

// PresenceType is the presence state to broadcast.
type PresenceType string

const (
	PresenceComposing PresenceType = "composing"
	PresenceRecording PresenceType = "recording"
	PresencePaused    PresenceType = "paused"
)

type PresenceParams struct {
	ChatID   string       `json:"chatId"`
	Presence PresenceType `json:"presence"`
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

// EventType is the type discriminator for webhook events.
type EventType string

const (
	EventMessage         EventType = "message"
	EventMessageSent     EventType = "message.sent"
	EventMessageDeleted  EventType = "message.deleted"
	EventMessageReaction EventType = "message.reaction"
	EventConnected       EventType = "connected"
	EventDisconnected    EventType = "disconnected"
	EventQR              EventType = "qr"
	EventCallIncoming    EventType = "call.incoming"
	EventCallAccepted    EventType = "call.accepted"
	EventCallRejected    EventType = "call.rejected"
)

// Event is a raw webhook event. Use [ParseEvent] to decode it.
type Event struct {
	Event     EventType       `json:"event"`
	SessionID string          `json:"session_id"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
}

// IncomingMessage is the data payload for [EventMessage] and [EventMessageSent].
type IncomingMessage struct {
	ID        string          `json:"id"`
	Chat      string          `json:"chat"`
	From      string          `json:"from"`
	FromMe    bool            `json:"fromMe"`
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Caption   string          `json:"caption,omitempty"`
	MediaURL  string          `json:"mediaUrl,omitempty"`
	Mimetype  string          `json:"mimetype,omitempty"`
	Filename  string          `json:"filename,omitempty"`
	Latitude  float64         `json:"latitude,omitempty"`
	Longitude float64         `json:"longitude,omitempty"`
	Timestamp int64           `json:"timestamp"`
	Quoted    *QuotedMessage  `json:"quoted,omitempty"`
}

type QuotedMessage struct {
	ID   string `json:"id"`
	From string `json:"from"`
	Text string `json:"text,omitempty"`
}
