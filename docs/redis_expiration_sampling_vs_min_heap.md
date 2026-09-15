---
title: "Redis Expiration: Hash Table Sampling vs Min-Heap"
date: 2026-09-15
project: "redis"
tool: "antigravity"
tags:
  - llm-generated
  - redis
status: unreviewed
---

## Raw output

# LLM Note: Systems Design Insight — Why Redis Uses a Hash Table + Sampling Over Min-Heap for Expiration

## 1. The Question & The Intuition
When thinking about key expirations (`TTL` / `EXPIRE`), the textbook computer science instinct is:
> *"Why use an unordered hash table (`db->expires`)? Why not maintain a min-heap, priority queue, or monotonic queue sorted by expiration timestamp, so Redis can simply peek at the head and pop expired keys in $O(1)$?"*

---

## 2. The Core Realization: "Physical Computation at Scale"
> **The Insight:**
> *"Even if you can remove things in the most theoretically efficient way, you have to consider doing millions of that operation all at once. Because it's computation, physically."*

In textbook algorithmic analysis, operations like popping the min element are treated as abstract mathematical steps. In real-world systems programming, freeing a key requires **physical hardware work**:
1. Invoking allocator `free()` calls (e.g. `jemalloc`).
2. Traversing and freeing nested structures (e.g., lists, hash tables, streams, zsets with thousands of elements).
3. Heap page coalescing and updating allocator arena bins.
4. Serializing `DEL` commands into replication buffers and AOF write buffers.

If 500,000 keys expire at the same instant (e.g. midnight `EXPIREAT` spikes), a priority queue would pop them all back-to-back synchronously. Doing that physical work in Redis's single-threaded event loop causes a **"stop-the-world" latency spike** of hundreds of milliseconds or seconds, stalling incoming traffic.

---

## 3. Why Redis Prefers Hash Map (`dict`) + Probabilistic Sampling

| Aspect | Min-Heap / Priority Queue | Redis Design (`dict` + Sampling) |
| :--- | :--- | :--- |
| **Write / Update Cost (`SET`, `EXPIRE`, `DEL`)** | $O(\log N)$ tree reshuffle, plus needs an auxiliary map for key-to-node indexing. | **$O(1)$** instant hash table insertion/deletion. |
| **Memory Overhead** | Pointers, array resizing, and back-references across tens of millions of keys. | Reuses key pointers directly from `db->keys`; zero extra pointer overhead. |
| **Batch Expiration Impact** | Synchronous cascade at expiry time spikes p99/p99.9 tail latency. | Work is strictly bounded and amortized over time. |

---

## 4. How Redis Solves It in Practice

Redis splits expiration into two cooperative mechanisms:

1. **Passive (Check-on-Access) — `expireIfNeeded()`:**
   - Every read/write command checks the key before touching it.
   - If expired, it is deleted on-demand.
   - Guarantees clients never observe expired data.

2. **Active (Adaptive Background Sampling) — `activeExpireCycle()`:**
   - Runs in `serverCron` (default 10Hz, every 100ms) and before sleeping in the event loop.
   - Samples 20 random keys with TTL.
   - Deletes expired keys found.
   - If $>25\%$ of sampled keys are expired, it repeats.
   - **Crucially: Enforces a strict CPU time quota (capped at ~25% CPU time or ~1–2ms per loop)**, ensuring incoming user traffic is never starved.

---

## 5. Broader Systems Philosophy: Bounded Latency > Textbook Optimality
Redis repeatedly applies this principle across its codebase:
- **LRU / LFU Eviction**: Samples 5 random keys rather than maintaining a global linked list requiring lock/pointer updates on every read.
- **Lazy Freeing (`UNLINK`)**: Offloads heavy deallocation to background worker threads to protect event loop responsiveness.
- **Predictable p99 Latency**: In high-throughput systems, an algorithm with bounded execution time beats a theoretically "faster" algorithm that introduces unpredictable latency spikes.
