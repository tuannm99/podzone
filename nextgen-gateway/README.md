## Customer PROXY for tenants matching route

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
      │  Middleware Layers  │ (auth, rate-limit, breaker, tenant-match)
      └─────────────────────┘
                │
                ▼
        [ Proxy to backend ]



+-------------+
REQUEST -> API GW -> TENANT-MATCHING -> K8s Ingress/GW -> Pod

```
