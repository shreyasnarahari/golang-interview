# Real Company Go Interview Questions & Solutions Index 🏢

This directory contains company-specific Go interview question banks, architecture deep dives, gotcha code snippets, and complete working solutions organized by top engineering companies.

> **Scope & attribution**: These are **company-flavored** question banks. Questions are inspired by each company's public engineering stack (open-source projects, engineering blogs, conference talks) and by **candidate reports** of interview loops — they are *not* verbatim transcripts of actual interviews, and inclusion here does not mean a company endorses or currently asks a given question. Use them to practice the domains each company is known for, not as leaked question lists.

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
| 🐶 **Datadog** | Datadog Agent 7+, metrics processing pipelines, APM tracing | Telemetry Agents, Atomic Metrics, Memory Monitoring | [datadog.md](datadog.md) |
| 💳 **Stripe** | Payment gateways, API proxies, idempotency engines, ledgering | Idempotency, Transaction Errors, Retries | [stripe.md](stripe.md) |
| 🎵 **ByteDance / TikTok** | Kitex RPC framework, Netpoll epoll library, video transcoding | Map Internals, Interface Boxing, Allocation Elimination | [bytedance.md](bytedance.md) |
| 🔷 **Meta** | Developer tooling, network probing agents, container engines | Struct Memory Layout, Alignment, Padding | [meta.md](meta.md) |
| 📦 **Amazon / AWS** | AWS Lambda Go runtime, ECS agents, AWS SDK for Go | Exponential Backoff with Jitter, Retries | [amazon.md](amazon.md) |
| 💻 **Microsoft / Azure** | Azure Kubernetes (AKS), Dapr runtime, Azure SDK | Table-Driven Testing, Mock Interfaces | [microsoft.md](microsoft.md) |
| 🪙 **Coinbase** | Crypto trading engine, order book matching, wallet infra | Order Book Engine, Concurrent Matching | [coinbase.md](coinbase.md) |
| 🟣 **Twitch** | Massive live chat WebSocket infrastructure, video ingestion | WebSocket Fan-Out, Chat Room Broadcaster | [twitch.md](twitch.md) |
| 🚗 **DoorDash** | Delivery dispatch engine, driver tracking, geofence routing | Circuit Breakers, Delivery State Machine | [doordash.md](doordash.md) |

---

## 🏗️ System Design Walkthroughs

Go-centric system-design walkthroughs for senior rounds — requirements, architecture, scaling, and the exact Go primitives to name-drop, cross-referenced to the `core-topics/` modules and `capstones/` projects.

| Design Problem | Core Trade-off | Walkthrough |
|---|---|---|
| **Distributed Rate Limiter** | Exactness (Redis/Lua) vs latency & availability (in-memory + gossip) | [system-design/distributed-rate-limiter.md](system-design/distributed-rate-limiter.md) |
| **Distributed Job Scheduler** | At-least-once delivery vs duplicate work; throughput vs ordering | [system-design/job-scheduler.md](system-design/job-scheduler.md) |
| **LRU Cache Service** | Hit ratio vs memory ceiling; lock contention vs consistency | [system-design/lru-cache-service.md](system-design/lru-cache-service.md) |

See [system-design/README.md](system-design/README.md) for the full index and interview strategy.

---

## 📊 80/20 Priority & Frequency Matrix

| Priority | Core Topic | Frequency | Target Roles |
|---|---|---|---|
| 🔴 **P1 (Do First)** | Data Types, Slice/Map Internals, Structs & Interfaces, Defer/Panic | 60% of all Go screens | Software Engineer, Backend, Infrastructure |
| 🟡 **P2 (Week 2)** | Goroutines & Channels, `sync` Package, Context, Errors, Generics | 25% of senior screens | Distributed Systems, Cloud-Native, Platform |
| 🟢 **P3 (Senior)** | GMP Scheduler Model, Tri-Color GC, Escape Analysis, `pprof`, `unsafe` | 15% specialized rounds | Staff Engineer, Systems Architect, HFT |
