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
	EventMessage          EventType = "message"
	EventMessageSent      EventType = "message.sent"
	EventReceipt          EventType = "receipt"
	EventMessageDeleted   EventType = "message.deleted"
	EventMessageReaction  EventType = "message.reaction"
	EventConnected        EventType = "connected"
	EventDisconnected     EventType = "disconnected"
	EventQR               EventType = "qr"
	EventPairingFailed    EventType = "pairing.failed"
	EventPairingSucceeded EventType = "pairing.succeeded"
	EventCallIncoming     EventType = "call.incoming"
	EventCallAccepted     EventType = "call.accepted"
	EventCallRejected     EventType = "call.rejected"
)

// Event is a raw webhook event. Use [ParseEvent] to decode it.
type Event struct {
	Event     EventType `json:"event"`
	SessionID string    `json:"session_id"`
	// SessionMetadata is the session's metadata (labels set with `metadata` on
	// POST /sessions or PATCH /sessions/{id}) when the event was sent, so you can
	// route on it without a lookup. Empty when the session has none, and nil on a
	// server older than this field.
	SessionMetadata map[string]string `json:"session_metadata,omitempty"`
	Data            json.RawMessage   `json:"data"`
	Timestamp       string            `json:"timestamp"`
}

// IncomingMessage is the data payload for [EventMessage] and [EventMessageSent].
type IncomingMessage struct {
	ID        string         `json:"id"`
	Chat      string         `json:"chat"`
	From      string         `json:"from"`
	FromMe    bool           `json:"fromMe"`
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	Caption   string         `json:"caption,omitempty"`
	MediaURL  string         `json:"mediaUrl,omitempty"`
	Mimetype  string         `json:"mimetype,omitempty"`
	Filename  string         `json:"filename,omitempty"`
	Latitude  float64        `json:"latitude,omitempty"`
	Longitude float64        `json:"longitude,omitempty"`
	Timestamp int64          `json:"timestamp"`
	Quoted    *QuotedMessage `json:"quoted,omitempty"`
}

// DisconnectReason says why a session went offline. The set is closed, but a
// newer server may add a value: decoding never fails on one, so branch on
// [Disconnected.WillReconnect] for a reason you do not know ([DisconnectReason.Known]).
type DisconnectReason string

const (
	// DisconnectTransient is a network or WhatsApp drop the engine retries by itself.
	DisconnectTransient DisconnectReason = "transient"
	// DisconnectManual is a disconnect you asked for.
	DisconnectManual DisconnectReason = "manual"
	// DisconnectPaused is a durable pause you asked for; resume it explicitly.
	DisconnectPaused DisconnectReason = "paused"
	// DisconnectLoggedOut means the link was removed; pair again.
	DisconnectLoggedOut DisconnectReason = "logged_out"
	// DisconnectReplaced means another client took over the session.
	DisconnectReplaced DisconnectReason = "replaced"
	// DisconnectTemporaryBan is a temporary ban; the engine reconnects after RetryAfterSeconds.
	DisconnectTemporaryBan DisconnectReason = "temporary_ban"
	// DisconnectClientOutdated means WhatsApp refused the engine's client version.
	DisconnectClientOutdated DisconnectReason = "client_outdated"
	// DisconnectConnectFailure is a connect error WhatsApp reported (see Code).
	DisconnectConnectFailure DisconnectReason = "connect_failure"
)

// DisconnectReasons returns every reason the server publishes today.
func DisconnectReasons() []DisconnectReason {
	return []DisconnectReason{
		DisconnectTransient, DisconnectManual, DisconnectPaused, DisconnectLoggedOut,
		DisconnectReplaced, DisconnectTemporaryBan, DisconnectClientOutdated, DisconnectConnectFailure,
	}
}

// Known reports whether r is one of [DisconnectReasons].
func (r DisconnectReason) Known() bool {
	for _, k := range DisconnectReasons() {
		if r == k {
			return true
		}
	}
	return false
}

// Disconnected is the data payload for [EventDisconnected]. WillReconnect is
// always present: true means the engine reconnects by itself, false means the
// session stays offline until you act (see Resume).
type Disconnected struct {
	Reason        DisconnectReason `json:"reason"`
	WillReconnect bool             `json:"willReconnect"`
	// Resume is the call that brings the session back ("connect" or "resume").
	// Empty while WillReconnect is true.
	Resume string `json:"resume,omitempty"`
	// Code is the code WhatsApp reported (logged_out, temporary_ban, connect_failure).
	Code int `json:"code,omitempty"`
	// Detail is diagnostic text; do not branch on it. "local" marks a logout you asked for.
	Detail string `json:"detail,omitempty"`
	// ExpireSeconds is the length of a temporary_ban.
	ExpireSeconds int `json:"expireSeconds,omitempty"`
	// RetryAfterSeconds is how long until the engine reconnects after a temporary_ban.
	RetryAfterSeconds int `json:"retryAfterSeconds,omitempty"`
}

// PairingMethod is how a pairing attempt links the device.
type PairingMethod string

const (
	// PairingQR is a scan of the QR code.
	PairingQR PairingMethod = "qr"
	// PairingCode is a pairing code typed on the phone ([Session.PairPhone]).
	PairingCode PairingMethod = "code"
)

// PairingFailureReason says why a pairing attempt ended without linking the
// number. The set is closed, but a newer server may add a value: decoding never
// fails on one, so treat a reason you do not know ([PairingFailureReason.Known])
// like [PairingError].
type PairingFailureReason string

const (
	// PairingTimeout means nobody read the QR or typed the code in time (the QR
	// channel lasts about 160 s).
	PairingTimeout PairingFailureReason = "timeout"
	// PairingRejected means WhatsApp refused the link.
	PairingRejected PairingFailureReason = "rejected"
	// PairingCancelled means disconnect or pause ended the attempt.
	PairingCancelled PairingFailureReason = "cancelled"
	// PairingClientOutdated means WhatsApp refuses the connector's version.
	PairingClientOutdated PairingFailureReason = "client_outdated"
	// PairingError is any other failure.
	PairingError PairingFailureReason = "error"
)

// PairingFailureReasons returns every reason the server publishes today.
func PairingFailureReasons() []PairingFailureReason {
	return []PairingFailureReason{
		PairingTimeout, PairingRejected, PairingCancelled, PairingClientOutdated, PairingError,
	}
}

// Known reports whether r is one of [PairingFailureReasons].
func (r PairingFailureReason) Known() bool {
	for _, k := range PairingFailureReasons() {
		if r == k {
			return true
		}
	}
	return false
}

// PairingFailed is the data payload for [EventPairingFailed]: how a pairing
// attempt ended without linking the number. The same result is on the session as
// lastPairing.
type PairingFailed struct {
	Method PairingMethod        `json:"method"`
	Reason PairingFailureReason `json:"reason"`
	// At is the end of the attempt, ISO 8601.
	At string `json:"at"`
}

// PairingSucceeded is the data payload for [EventPairingSucceeded], sent just
// before [EventConnected].
type PairingSucceeded struct {
	Method PairingMethod `json:"method"`
	// At is the end of the attempt, ISO 8601.
	At string `json:"at"`
}

// ReceiptStatus is the normalized delivery state of a [Receipt].
type ReceiptStatus string

const (
	ReceiptDelivered ReceiptStatus = "delivered"
	ReceiptRead      ReceiptStatus = "read"
	ReceiptPlayed    ReceiptStatus = "played"
)

// Receipt is the data payload for [EventReceipt]. One receipt can acknowledge
// several messages (MessageIDs); in groups there is one receipt per participant.
type Receipt struct {
	// Status is empty for receipt types that say nothing about delivery
	// (retry, server-error, inactive, peer_msg, hist_sync).
	Status ReceiptStatus `json:"status,omitempty"`
	// RawType is the raw WhatsApp type ("" delivered, sender, read, read-self,
	// played, played-self, ...). Type carries the same value.
	RawType    string   `json:"rawType"`
	Type       string   `json:"type"`
	Chat       string   `json:"chat"`
	Sender     string   `json:"sender"`
	MessageIDs []string `json:"messageIds"`
	// Timestamp is unix seconds.
	Timestamp int64 `json:"timestamp"`
	// IsFromMe: the receipt came from another device of your own account.
	IsFromMe bool `json:"isFromMe"`
	// MessageFromMe: the acknowledged messages were sent by you; only these
	// update the status of sent messages.
	MessageFromMe bool   `json:"messageFromMe"`
	MessageSender string `json:"messageSender,omitempty"`
}

type QuotedMessage struct {
	ID   string `json:"id"`
	From string `json:"from"`
	Text string `json:"text,omitempty"`
}
