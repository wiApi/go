# wi-api Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/wiApi/go.svg)](https://pkg.go.dev/github.com/wiApi/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wiApi/go)](https://goreportcard.com/report/github.com/wiApi/go)
[![license](https://img.shields.io/github/license/wiApi/go?style=flat-square)](LICENSE)

Go SDK for [wi-api](https://wi.api.br). Send and receive WhatsApp messages from any Go application.

No runtime dependencies. Context-aware. Idiomatic error types.

## Install

```bash
go get github.com/wiApi/go
```

Requires Go 1.27+.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    wi "github.com/wiApi/go"
)

func main() {
    client := wi.New(os.Getenv("WI_API_KEY"))

    msg, err := client.Session("my-instance").SendText(context.Background(), wi.SendTextParams{
        To:   "5511999999999",
        Text: "Hello from wi-api",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("sent:", msg.MessageID)
}
```

## Sessions

```go
ctx := context.Background()
session := client.Session("my-instance")

// Start connection flow
err := session.Connect(ctx)

// Get QR code (base64 PNG)
qr, err := session.QR(ctx)
fmt.Println(qr.QR)

// Pair by phone number
result, err := session.PairPhone(ctx, "5511999999999")
fmt.Println(result.PairCode) // 8-character code

// Check status
status, err := session.Status(ctx)
fmt.Println(status.Connected, status.Phone)

session.Disconnect(ctx)
session.Logout(ctx)
```

## Sending messages

```go
// Text
session.SendText(ctx, wi.SendTextParams{To: "5511999999999", Text: "Hello!"})

// Image
session.SendImage(ctx, wi.SendImageParams{
    To:      "5511999999999",
    URL:     "https://example.com/photo.jpg",
    Caption: "Look at this",
})

// Voice note
session.SendAudio(ctx, wi.SendAudioParams{To: "5511999999999", URL: "https://example.com/audio.ogg", PTT: true})

// Document
session.SendDocument(ctx, wi.SendDocumentParams{
    To:       "5511999999999",
    URL:      "https://example.com/report.pdf",
    Filename: "report.pdf",
})

// Location
session.SendLocation(ctx, wi.SendLocationParams{
    To:        "5511999999999",
    Latitude:  -23.5505,
    Longitude: -46.6333,
    Title:     "São Paulo",
})

// Poll
session.SendPoll(ctx, wi.SendPollParams{
    To:       "5511999999999",
    Question: "Best stack?",
    Options:  []string{"Go", "Rust", "Node.js"},
})

// Reaction
session.React(ctx, wi.SendReactionParams{
    To:        "5511999999999",
    MessageID: messageID,
    Emoji:     "",
})
```

## Webhooks

```go
import (
    "encoding/json"
    "net/http"
    "os"

    wi "github.com/wiApi/go"
)

http.Handle("/webhook", wi.WebhookHandler(
    os.Getenv("WI_WEBHOOK_SECRET"),
    func(event *wi.Event) error {
        switch event.Event {
        case wi.EventMessage:
            var msg wi.IncomingMessage
            json.Unmarshal(event.Data, &msg)
            fmt.Println(msg.From, msg.Text)
        case wi.EventConnected:
            fmt.Println("connected:", event.SessionID)
        }
        return nil
    },
))
```

Verify manually:

```go
body, _ := io.ReadAll(r.Body)
if !wi.VerifySignature(body, r.Header.Get("x-wi-signature"), []byte(secret)) {
    http.Error(w, "invalid signature", http.StatusUnauthorized)
    return
}
event, err := wi.ParseEvent(body)
```

## Error handling

```go
msg, err := session.SendText(ctx, params)
if err != nil {
    var wiErr *wi.Error
    if errors.As(err, &wiErr) {
        fmt.Println(wiErr.Status, wiErr.Message)
    }
}
```

## Configuration

```go
client := wi.New(apiKey,
    wi.WithBaseURL("https://endpoint.wi.api.br"),
    wi.WithTimeout(10 * time.Second),
)
```

## Resources

- [pkg.go.dev](https://pkg.go.dev/github.com/wiApi/go)
- [Dashboard](https://wi.api.br)
- [Docs](https://docs.wi.api.br)
- [Changelog](https://github.com/wiApi/go/releases)
