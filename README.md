# PulseGuard

Hafif Filo İzleme & Olay Toplama Sistemi.

## Bileşenler

### **Agent**

The PulseGuard Agent is a high-performance, fault-tolerant metrics collection unit designed to run at the edge. Built with Go's powerful concurrency primitives, it operates asynchronously without causing overhead on the host system.

#### Key Features

*   **Concurrent Sensors (Goroutines):** CPU, RAM, and Disk usage metrics are continuously monitored by independent, non-blocking threads (Goroutines).
*   **Zero Data Loss Architecture (At-Least-Once Delivery):** To protect against network outages, collected data is safely persisted to a local disk using **bbolt (BoltDB)** (ACID compliant) before network transmission. Records are only deleted from the local vault upon receiving a confirmed `200 OK` response from the Collector.
*   **Batch Processing:** To optimize network bandwidth and reduce overhead, events are accumulated and transmitted in batches rather than individually.
*   **Cryptographic Security (HMAC-SHA256):** To ensure payload integrity and prevent data spoofing between the Agent and the Collector, every outgoing payload is cryptographically signed using a predefined Secret Key. The signature is sent via the `X-Signature` HTTP header.
*   **Network Resilience (Timeouts):** Strict timeout policies are enforced on the HTTP client to prevent Goroutine memory leaks in the event of server hangs or unresponsiveness.

#### Configuration (`config.yaml`)

The agent's behavior is dynamically managed via the `config.yaml` file:

```yaml
agent:
  host_id: "server-01"
  interval_seconds: 5
  secret_key: "my-super-secret-key"

collector:
  url: "http://localhost:8080/api/events"

checks:
  cpu:
    enabled: true
    threshold_warning: 80.0
    threshold_error: 95.0

  ram:
    enabled: true
    threshold_warning: 85.0
    threshold_error: 98.0

  disk:
    enabled: true
    path: "/"
    threshold_warning: 85.0
    threshold_error: 92.0
```

To run the agent locally, navigate to the agent directory and execute:

**go run .**

To compile and run a standalone binary:

**go build -o agent**

**./agent**

- **Collector:** Olayları alan, doğrulayan ve REST API sunan Go servisi.
- **Dashboard:** React ve TypeScript ile yazılmış filo izleme arayüzü.
- **pulsectl:** Ajanı ve konfigürasyonunu yöneten CLI aracı.

## Kurulum ve Çalıştırma
*(Bu bölüm ilerleyen aşamalarda doldurulacaktır.)*