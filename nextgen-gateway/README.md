```python
+-------------+
Inbound →  |  Axum App   | ← API admin
           +-------------+
                │
                ▼
        [ Matching Engine ]
                │
                ▼
      ┌─────────────────────┐
      │  Middleware Layers  │ (auth, rate-limit, breaker)
      └─────────────────────┘
                │
                ▼
        [ Proxy to backend ]

```
