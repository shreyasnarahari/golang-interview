# Real Company Go Interview Questions & Solutions Index 🏢

This directory contains company-specific Go interview question banks, architecture deep dives, gotcha code snippets, and complete working solutions organized by top engineering companies.

---

## 🗺️ Company Question Directory

| Company | Key Go Systems & Use Cases | Target Domain | Question Bank |
|---|---|---|---|
| 🔵 **Google** | gRPC microservices, Kubernetes control planes, network automation | Distributed Systems, Cloud Infrastructure | [google.md](google.md) |
| 🔴 **Uber** | 4,000+ microservices, dispatching, geofencing, Jaeger tracing | High-Throughput Concurrency, Rate Limiting | [uber.md](uber.md) |
| ☁️ **Cloudflare** | Edge proxies, DNS infrastructure, DDoS mitigation, network probes | Zero-Copy I/O, Low-Latency Networking | [cloudflare.md](cloudflare.md) |
| ⚙️ **HashiCorp** | Terraform, Vault, Consul, Nomad, Packer, Raft consensus | Distributed State, Gossip Protocols, Plugins | [hashicorp.md](hashicorp.md) |
| ☸️ **Kubernetes / CNCF** | Kubernetes API Server, Custom Controllers, Etcd, Helm, Envoy | Controllers, Informers, WorkQueues, Cgroups | [kubernetes-cncf.md](kubernetes-cncf.md) |
| 🪳 **CockroachDB / TiDB** | Distributed SQL engines, Raft log compaction, MVCC storage | Database Engines, Binary Encoding, MVCC | [cockroachdb.md](cockroachdb.md) |
| 🐶 **Datadog** | Datadog Agent 7+, metrics processing pipelines, APM tracing | Telemetry Agents, Profiling (`pprof`), Metrics | [datadog.md](datadog.md) |
| 💳 **Stripe** | Payment gateways, API proxies, idempotency engines, ledgering | Idempotency, Transaction Errors, Retries | [stripe.md](stripe.md) |
| 🎵 **ByteDance / TikTok** | Kitex RPC framework, Netpoll epoll library, video transcoding | High Concurrency, epoll I/O, Worker Pools | [bytedance.md](bytedance.md) |
| 🔷 **Meta** | Developer tooling, network probing agents, container engines | Struct Memory Layout, Alignment, Padding | [meta.md](meta.md) |
| 📦 **Amazon / AWS** | AWS Lambda Go runtime, ECS agents, AWS SDK for Go | Exponential Backoff with Jitter, Retries | [amazon.md](amazon.md) |
| 💻 **Microsoft / Azure** | Azure Kubernetes (AKS), Dapr runtime, Azure SDK | Table-Driven Testing, Mock Interfaces | [microsoft.md](microsoft.md) |
| 🪙 **Coinbase** | Crypto trading engine, order book matching, wallet infra | Order Book Engine, Concurrent Matching | [coinbase.md](coinbase.md) |
| 🟣 **Twitch** | Massive live chat WebSocket infrastructure, video ingestion | WebSocket Fan-Out, Chat Room Broadcaster | [twitch.md](twitch.md) |
| 🚗 **DoorDash** | Delivery dispatch engine, driver tracking, geofence routing | Circuit Breakers, Delivery State Machine | [doordash.md](doordash.md) |

---

## 📊 80/20 Priority & Frequency Matrix

| Priority | Core Topic | Frequency | Target Roles |
|---|---|---|---|
| 🔴 **P1 (Do First)** | Data Types, Slice/Map Internals, Structs & Interfaces, Defer/Panic | 60% of all Go screens | Software Engineer, Backend, Infrastructure |
| 🟡 **P2 (Week 2)** | Goroutines & Channels, `sync` Package, Context, Errors, Generics | 25% of senior screens | Distributed Systems, Cloud-Native, Platform |
| 🟢 **P3 (Senior)** | GMP Scheduler Model, Tri-Color GC, Escape Analysis, `pprof`, `unsafe` | 15% specialized rounds | Staff Engineer, Systems Architect, HFT |
