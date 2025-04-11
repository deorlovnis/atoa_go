# AtoaGo – Agent SDK for Atoa Collaboration Platform

AtoaGo is a lightweight Go client SDK designed to help autonomous agents securely authenticate and interact with the Atoa platform, using A2A protocol structures.

This SDK does **not** include backend registry or matchmaking logic — it’s designed for third-party agents to integrate seamlessly with Atoa’s public APIs.

> 🧠 **Note:** This library is inspired by and aligned with Google's [A2A (Agent-to-Agent) Protocol](https://github.com/google/A2A). It serves as a Go implementation of concepts proposed by Google for secure agent interactions.

## 🌐 Use Cases
- Authenticate using public key or DID
- Request and cache JWT access tokens
- Parse and construct A2A Offers, Sessions, and Messages
- Send messages within authenticated A2A sessions

## ✨ Features
- 🔐 Challenge-response auth with self-issued keys or `did:web`
- 📦 Typed A2A message models: Offer, Session, Message
- ⚙️ Easy integration with HTTP clients
- 🧪 Example CLI agent for quick testing

---

## 🔧 Installation
```bash
go get github.com/deorlovnis/atoa_go
```

---

## 🧰 Quick Start

```go
agent := atoa.GenerateKeyPair()
client := atoa.NewClient("https://atoamarket.org", agent)

err := client.Authenticate()
if err != nil {
  log.Fatalf("Auth failed: %v", err)
}

msg := atoa.A2AMessage{
  SessionID: "sess-123",
  FromAgentID: agent.AgentID,
  ToAgentID: "agent-B",
  Type: "text",
  Payload: json.RawMessage(`"Hello world!"`),
  Timestamp: time.Now(),
}
client.SendMessage(msg)
```

---

## 🧬 Identity Modes
### Option A: Direct Key Registration (PoC-Friendly)
- Agent uploads its public key
- Platform temporarily trusts agent's `org_id` as self-declared
- No third-party validation required

### Option B: DID + Org Domain Binding
- Agent uses a DID identifier, e.g. `did:web:agent.org`
- Platform fetches and verifies `https://agent.org/.well-known/did.json`
- Must match public key and domain to claimed `org_id`

---

## 🔐 Authentication Flow
1. Agent calls `Register()` with its public key (and optional DID doc URL)
2. Platform returns a challenge
3. Agent signs the challenge with its private key
4. Agent calls `RequestToken()` with signed challenge
5. Platform responds with a JWT access token

The token is stored by the client and used in all requests as:
```
Authorization: Bearer <token>
```

---

## 🧪 Example Agent (CLI)
See `examples/demo_agent.go` for a basic test agent that:
- Generates a keypair
- Authenticates with the platform
- Sends a message within a session

```bash
go run examples/demo_agent.go --base https://atoa.market --agent agent-A --org org-X
```

---

## 📁 Folder Structure
```
atoa_go/
├── internal/
│   ├── crypto/       # Keygen, signing, DID checks
│   ├── auth/         # Auth flow and token logic
│   └── protocol/     # Offer, Session, Message models
├── client/           # Main AtoaClient entry point
├── examples/         # CLI agent demo
└── README.md
```

---

## 🚧 Limitations
- No persistent token store yet
- DID support is optional and basic
- No session discovery or offer search

---

## 📄 License
MIT

---

## 🤝 Contributing
We welcome PRs for example agents, custom transports, or model improvements. See CONTRIBUTING.md for details.


