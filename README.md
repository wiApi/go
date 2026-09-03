# wi-api — Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/wi-api/go.svg)](https://pkg.go.dev/github.com/wi-api/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wi-api/go?style=flat-square)](https://goreportcard.com/report/github.com/wi-api/go)
[![license](https://img.shields.io/github/license/wi-api/go?style=flat-square&color=0d9373)](LICENSE)

Official Go SDK for the [wi-api](https://wi.api.br) WhatsApp platform.

- Zero runtime dependencies — stdlib only
- Context-aware API
- Functional options pattern
- Idiomatic `error` types via `*wi.Error`

---

## Install

```bash
go get github.com/wi-api/go
```

Requires Go 1.22+.

---

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    wi "github.com/wi-api/go"
)

func main() {
    client := wi.New(os.Getenv("WI_API_KEY"))
    session := client.Session("my-instance")

    msg, err := session.SendText(context.Background(), wi.SendTextParams{
        To:   "5511999999999",
        Text: "Hello from wi-api",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("sent:", msg.MessageID)
}
```

---

## Sessions

```go
ctx := context.Background()

// Connect — triggers QR or pairphone flow
err := session.Connect(ctx)

// Get QR code (base64 PNG)
qr, err := session.QR(ctx)
fmt.Println(qr.QR) // base64 string

// Pair by phone number
result, err := session.PairPhone(ctx, "5511999999999")
fmt.Println(result.PairCode) // 8-char code

// Status
status, err := session.Status(ctx)
fmt.Println(status.Connected, status.Phone)

// Disconnect / logout
err = session.Disconnect(ctx)
err = session.Logout(ctx)
```

---

## Sending messages

```go
// Text
session.SendText(ctx, wi.SendTextParams{
    To:   "5511999999999",
    Text: "Hello!",
})

// Image
session.SendImage(ctx, wi.SendImageParams{
    To:      "5511999999999",
    URL:     "https://example.com/photo.jpg",
    Caption: "Check this out",
})

// Voice note
session.SendAudio(ctx, wi.SendAudioParams{
    To:  "5511999999999",
    URL: "https://example.com/audio.ogg",
    PTT: true,
})

// Document
session.SendDocument(ctx, wi.SendDocumentParams{
    To:       "5511999999999",
    URL:      "https://example.com/report.pdf",
    Filename: "Q3-report.pdf",
})

// Location
session.SendLocation(ctx, wi.SendLocationParams{
    To:        "5511999999999",
    Latitude:  -23.5505,
    Longitude: -46.6333,
    Title:     "São Paulo",
})

// Reaction
session.React(ctx, wi.SendReactionParams{
    To:        "5511999999999",
    MessageID: "MESSAGE_ID",
    Emoji:     "👍",
})

// Poll
session.SendPoll(ctx, wi.SendPollParams{
    To:       "5511999999999",
    Question: "Best stack?",
    Options:  []string{"Go + Fiber", "Bun + Elysia", "Rust + Axum"},
})
```

---

## Webhooks

```go
import (
    wi "github.com/wi-api/go"
    "encoding/json"
)

http.Handle("/webhook", wi.WebhookHandler(os.Getenv("WI_WEBHOOK_SECRET"),
    func(event *wi.Event) error {
        switch event.Event {
        case wi.EventMessage:
            var msg wi.IncomingMessage
            if err := json.Unmarshal(event.Data, &msg); err != nil {
                return err
            }
            fmt.Printf("[%s] %s: %s\n", msg.Chat, msg.From, msg.Text)

        case wi.EventConnected:
            fmt.Println("session connected:", event.SessionID)
        }
        return nil
    },
))
```

### Verify manually

```go
body, _ := io.ReadAll(r.Body)
sig := r.Header.Get("x-wi-signature")

if !wi.VerifySignature(body, sig, []byte(secret)) {
    http.Error(w, "invalid signature", http.StatusUnauthorized)
    return
}

event, err := wi.ParseEvent(body)
```

---

## Error handling

```go
msg, err := session.SendText(ctx, wi.SendTextParams{To: "...", Text: "..."})
if err != nil {
    var wiErr *wi.Error
    if errors.As(err, &wiErr) {
        fmt.Println(wiErr.Status, wiErr.Message, wiErr.Code)
    }
    return err
}
```

---

## Custom HTTP client

```go
client := wi.New(apiKey,
    wi.WithBaseURL("https://endpoint.wi.api.br"),
    wi.WithTimeout(10 * time.Second),
    wi.WithHTTPClient(&http.Client{
        Transport: myTransport,
    }),
)
```

---

## License

MIT — [wi.api.br](https://wi.api.br)
