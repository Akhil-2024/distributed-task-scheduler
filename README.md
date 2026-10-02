# Distributed Task Scheduler & Fault-Tolerant Job Processing Platform

A backend/distributed-systems project built in **Go** to learn and demonstrate job scheduling, worker coordination, retries, fault tolerance, message queues, caching, distributed locking, containerization, orchestration, and monitoring.

> **Current status:** Core REST API, PostgreSQL persistence, worker execution, priority scheduling, safe concurrent job claiming, retry logic, exponential backoff, and DEAD-state handling are implemented. Kafka, Redis, Docker, Kubernetes, monitoring, and several production-hardening features are still planned.

---

## 1. Project Goal

Build a scalable and fault-tolerant platform where clients can submit background jobs and workers process them asynchronously.

Example flow:

```text
Client
  |
  v
Go REST API
  |
  v
PostgreSQL
  |
  v
Job Queue / Scheduler
  |
  +------> Worker 1
  +------> Worker 2
  +------> Worker 3
  |
  v
Retry / Failure Handling
```

The final version will use:

- Go
- PostgreSQL
- Kafka
- Redis
- REST
- gRPC
- Docker
- Kubernetes
- Prometheus
- Grafana

---

## 2. Current Architecture

```text
Postman / Client
      |
      v
+------------------+
|     Go API       |
+------------------+
      |
      v
+------------------+
|   PostgreSQL     |
|    jobs table    |
+------------------+
      ^
      |
+------------------+
|   Go Worker      |
+------------------+
```

The worker currently polls PostgreSQL and safely claims queued jobs using:

```sql
FOR UPDATE SKIP LOCKED
```

This prevents two workers from processing the same queued job at the same time.

---

## 3. Features Completed

### REST API
- [x] `GET /health`
- [x] `POST /jobs`
- [x] `GET /jobs`
- [x] `GET /jobs/{id}`

### Job Model
- [x] Unique job ID
- [x] Job type
- [x] JSON payload
- [x] Status
- [x] Priority
- [x] Retry count
- [x] Maximum retries
- [x] Created timestamp
- [x] Updated timestamp

### PostgreSQL
- [x] Persistent job storage
- [x] JSONB payload
- [x] Job status tracking
- [x] Retry metadata
- [x] `last_error`
- [x] `next_retry_at`

### Worker
- [x] Worker process
- [x] Priority-based job selection
- [x] Job status transition
- [x] `QUEUED -> PROCESSING -> SUCCESS`
- [x] Multiple-worker-safe claiming
- [x] PostgreSQL row locking
- [x] `FOR UPDATE SKIP LOCKED`

### Fault Tolerance
- [x] Simulated job failures
- [x] Retry count tracking
- [x] Exponential backoff
- [x] Maximum retry limit
- [x] `DEAD` status after max retries
- [x] Error storage in PostgreSQL

---

## 4. Job Lifecycle

Successful job:

```text
QUEUED
   |
   v
PROCESSING
   |
   v
SUCCESS
```

Failed job with retries:

```text
QUEUED
   |
   v
PROCESSING
   |
   v
FAILED
   |
   v
QUEUED (retry scheduled)
   |
   v
PROCESSING
   |
   v
...
   |
   v
DEAD
```

---

## 5. Priority Scheduling

Jobs are selected using:

```sql
ORDER BY priority DESC, created_at ASC
```

Example:

```text
Priority 10 -> processed first
Priority 5  -> processed next
Priority 2  -> processed later
```

When jobs have the same priority, the older job is selected first.

---

## 6. Safe Concurrent Job Claiming

To support multiple workers safely, PostgreSQL row locking is used:

```sql
SELECT id, job_type, retry_count, max_retries
FROM jobs
WHERE status = 'QUEUED'
ORDER BY priority DESC, created_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

### Why this matters

Without row locking:

```text
Worker 1 -> Job A
Worker 2 -> Job A
```

With `FOR UPDATE SKIP LOCKED`:

```text
Worker 1 -> Job A
Worker 2 -> Job B
Worker 3 -> Job C
```

---

## 7. Retry Strategy

A failed job is retried using exponential backoff.

Example:

```text
Retry 1 -> wait 2 seconds
Retry 2 -> wait 4 seconds
Retry 3 -> max retries reached
```

Current formula:

```text
delay = 2 ^ retry_count
```

After the retry limit is reached:

```text
status = DEAD
```

---

## 8. Database Schema

Current `jobs` table:

```sql
CREATE TABLE jobs (
    id VARCHAR(50) PRIMARY KEY,
    job_type VARCHAR(100) NOT NULL,
    payload JSONB,
    status VARCHAR(30) NOT NULL,
    priority INT DEFAULT 0,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

Additional fields:

```sql
ALTER TABLE jobs
ADD COLUMN last_error TEXT,
ADD COLUMN next_retry_at TIMESTAMPTZ;
```

---

## 9. Project Structure

```text
distributed-task-scheduler/
|
|-- cmd/
|   |-- api/
|   |   `-- main.go
|   |
|   `-- worker/
|       `-- main.go
|
|-- internal/
|   |-- database/
|   |   `-- db.go
|   |
|   |-- handler/
|   |   `-- job_handler.go
|   |
|   |-- model/
|   |   `-- job.go
|   |
|   |-- repository/
|   |   `-- job_repository.go
|   |
|   `-- service/
|       `-- job_service.go
|
|-- go.mod
|-- go.sum
|-- .gitignore
`-- README.md
```

---

## 10. Local Setup

### Requirements

Install:

- Go
- PostgreSQL
- pgAdmin 4
- Git
- Postman

### Create database

Create:

```text
task_scheduler
```

### Environment variable

Set the PostgreSQL connection string before starting the API or worker.

Windows Command Prompt:

```bat
set DATABASE_URL=postgres://postgres:<URL_ENCODED_PASSWORD>@localhost:5432/task_scheduler
```

> Never commit your real database password to GitHub.

### Start API

```bat
go run ./cmd/api
```

Expected:

```text
Connected to PostgreSQL successfully
Distributed Task Scheduler
Server running on http://localhost:8080
```

### Start worker

Open another terminal:

```bat
go run ./cmd/worker
```

Expected:

```text
Worker connected to PostgreSQL
Worker started...
```

---

## 11. API Examples

### Health Check

```http
GET /health
```

### Create Job

```http
POST /jobs
```

Example:

```json
{
  "type": "send_email",
  "payload": {
    "email": "test@example.com"
  },
  "priority": 10,
  "max_retries": 3
}
```

### List Jobs

```http
GET /jobs
```

### Get Job by ID

```http
GET /jobs/{id}
```

---

# 12. Remaining Roadmap

## Phase 1 - Improve Current API and Worker

- [ ] Expose `last_error` in API responses
- [ ] Expose `next_retry_at` in API responses
- [ ] Add `FAILED` state if needed before retrying
- [ ] Add `POST /jobs/{id}/retry`
- [ ] Add `DELETE /jobs/{id}` / cancel job
- [ ] Add `CANCELLED` state
- [ ] Validate allowed job types
- [ ] Validate payload structure
- [ ] Add request size limits
- [ ] Add graceful API shutdown
- [ ] Add graceful worker shutdown
- [ ] Add context timeouts
- [ ] Add per-job execution timeout
- [ ] Add database indexes for frequently queried columns
- [ ] Add database migrations instead of running SQL manually
- [ ] Add structured logging

---

## Phase 2 - Worker Improvements

- [ ] Build a worker pool using goroutines
- [ ] Use Go channels
- [ ] Configure worker concurrency
- [ ] Add worker IDs
- [ ] Track which worker owns a job
- [ ] Add worker heartbeat
- [ ] Detect dead workers
- [ ] Add job lease / lease expiration
- [ ] Requeue jobs abandoned by crashed workers
- [ ] Add worker health endpoint
- [ ] Add clean shutdown while finishing active jobs

---

## Phase 3 - Kafka

Replace PostgreSQL polling as the primary job-delivery mechanism.

Planned architecture:

```text
Client
  |
  v
Go API
  |
  +----> PostgreSQL
  |
  v
Kafka Producer
  |
  v
Kafka Topic
  |
  +----> Worker 1
  +----> Worker 2
  +----> Worker 3
```

Tasks:

- [ ] Install/run Kafka locally
- [ ] Add Kafka producer
- [ ] Publish job after creation
- [ ] Add Kafka consumer
- [ ] Add consumer groups
- [ ] Add multiple partitions
- [ ] Understand offset commits
- [ ] Handle consumer rebalance
- [ ] Handle duplicate delivery
- [ ] Decide commit-before/after-processing semantics
- [ ] Add retry topics
- [ ] Add Dead Letter Queue topic
- [ ] Add Kafka failure handling

Concepts to understand:

- Producer
- Consumer
- Broker
- Topic
- Partition
- Offset
- Consumer group
- Replication
- At-most-once
- At-least-once
- Exactly-once concepts

---

## Phase 4 - Idempotency

Kafka commonly provides at-least-once delivery, so workers must safely handle duplicate messages.

Tasks:

- [ ] Add idempotency key
- [ ] Prevent duplicate execution
- [ ] Store execution state
- [ ] Handle duplicate Kafka messages
- [ ] Test duplicate job delivery

Example:

```text
job-123 arrives twice
        |
        v
only one execution should happen
```

---

## Phase 5 - Redis

Planned Redis use cases:

- [ ] Job caching
- [ ] Idempotency keys
- [ ] Distributed locks
- [ ] Worker heartbeat
- [ ] Rate limiting
- [ ] Temporary execution metadata

Example cache flow:

```text
GET /jobs/{id}
      |
      v
    Redis
   /     \
hit      miss
 |         |
return   PostgreSQL
           |
           v
         Redis
```

---

## Phase 6 - Dead Letter Queue

Current `DEAD` jobs are stored in PostgreSQL.

Planned improvement:

- [ ] Create Kafka DLQ topic
- [ ] Publish permanently failed jobs to DLQ
- [ ] Store failure reason
- [ ] Add DLQ inspection API
- [ ] Add manual replay/retry from DLQ

---

## Phase 7 - Priority Scheduling with Kafka

Kafka is not a native priority queue.

Possible implementation:

```text
jobs-high
jobs-medium
jobs-low
```

Tasks:

- [ ] Create priority topics
- [ ] Route jobs based on priority
- [ ] Define fair scheduling policy
- [ ] Avoid starvation of low-priority jobs
- [ ] Document the trade-off

---

## Phase 8 - REST and gRPC

REST will remain the external API.

Planned gRPC usage:

- [ ] Define Protocol Buffer files
- [ ] Generate Go gRPC code
- [ ] Add internal worker/scheduler service
- [ ] Implement unary RPC
- [ ] Learn streaming RPC
- [ ] Compare REST vs gRPC

---

## Phase 9 - Docker

Containerize:

- [ ] API
- [ ] Worker
- [ ] PostgreSQL
- [ ] Redis
- [ ] Kafka

Tasks:

- [ ] Create API Dockerfile
- [ ] Create worker Dockerfile
- [ ] Create `docker-compose.yml`
- [ ] Configure environment variables
- [ ] Add health checks
- [ ] Start complete local stack with one command

Target:

```bash
docker compose up
```

---

## Phase 10 - Kubernetes

Deploy:

- [ ] API Deployment
- [ ] API Service
- [ ] Worker Deployment
- [ ] ConfigMap
- [ ] Secret
- [ ] Readiness probe
- [ ] Liveness probe
- [ ] Rolling updates
- [ ] Multiple replicas
- [ ] Horizontal Pod Autoscaler

Target architecture:

```text
               Load Balancer
                    |
              +-----v------+
              | API Service |
              +-----+------+
                    |
             +------+------+
             |             |
          API Pod 1     API Pod 2

                  Kafka
                    |
          +---------+---------+
          |         |         |
       Worker 1  Worker 2  Worker 3
```

---

## Phase 11 - Autoscaling

- [ ] Measure Kafka consumer lag
- [ ] Scale workers based on workload
- [ ] Test worker scale-up
- [ ] Test worker scale-down
- [ ] Document partition-count limitations

---

## Phase 12 - Monitoring

Use:

- Prometheus
- Grafana

Metrics:

- [ ] `jobs_created_total`
- [ ] `jobs_completed_total`
- [ ] `jobs_failed_total`
- [ ] `jobs_dead_total`
- [ ] `job_processing_duration`
- [ ] API request latency
- [ ] Worker count
- [ ] Active jobs
- [ ] Queue size
- [ ] Kafka consumer lag
- [ ] Retry count

Grafana dashboards:

- [ ] Jobs per second
- [ ] Success/failure rate
- [ ] P50 latency
- [ ] P95 latency
- [ ] P99 latency
- [ ] Worker utilization
- [ ] Kafka lag

---

## Phase 13 - Testing

### Unit Tests

- [ ] Job service
- [ ] Validation
- [ ] Retry logic
- [ ] Priority rules
- [ ] Backoff calculation
- [ ] Job state transitions

### Integration Tests

- [ ] API -> PostgreSQL
- [ ] API -> Kafka
- [ ] Kafka -> Worker
- [ ] Worker -> PostgreSQL
- [ ] Worker -> Redis

### Failure Tests

- [ ] Worker crash
- [ ] API restart
- [ ] PostgreSQL unavailable
- [ ] Redis unavailable
- [ ] Kafka unavailable
- [ ] Duplicate job
- [ ] Invalid payload
- [ ] Job timeout
- [ ] Network interruption

---

## Phase 14 - Load Testing

Use a load-testing tool such as k6.

Tests:

- [ ] 100 jobs
- [ ] 1,000 jobs
- [ ] 10,000 jobs
- [ ] Multiple API clients
- [ ] Multiple workers

Measure:

- Requests/sec
- Jobs/sec
- P50 latency
- P95 latency
- P99 latency
- Failure rate
- Queue delay
- End-to-end processing latency

Only publish performance numbers that were actually measured.

---

## Phase 15 - Security and Production Hardening

- [ ] Do not commit passwords/secrets
- [ ] Use environment variables
- [ ] Kubernetes Secrets
- [ ] Authentication
- [ ] Authorization
- [ ] API rate limiting
- [ ] Input validation
- [ ] SQL injection protection
- [ ] TLS
- [ ] CORS policy
- [ ] Audit logging
- [ ] Database connection pool tuning
- [ ] Resource limits
- [ ] Job payload size limits

---

## Phase 16 - CI/CD

- [ ] GitHub Actions
- [ ] Run `go test ./...`
- [ ] Run formatting checks
- [ ] Run static analysis
- [ ] Build Docker images
- [ ] Push images
- [ ] Optional automated deployment

---

# 13. Final Target Architecture

```text
                            CLIENT
                               |
                               | REST
                               v
                       +---------------+
                       |    Go API     |
                       +-------+-------+
                               |
                  +------------+------------+
                  |                         |
                  v                         v
            +-----------+             +---------+
            | PostgreSQL|             |  Redis  |
            | Job State |             | Cache / |
            +-----------+             | Locks   |
                  |                   +---------+
                  |
                  v
              +-------+
              | Kafka |
              +---+---+
                  |
       +----------+----------+
       |          |          |
       v          v          v
    Worker 1   Worker 2   Worker 3
       |          |          |
       +----------+----------+
                  |
                  v
           Retry / DLQ

       +----------------------+
       | Prometheus + Grafana |
       +----------------------+

              Kubernetes
```

---

# 14. Interview Topics Covered by This Project

This project is intended to demonstrate understanding of:

### Backend Engineering
- REST APIs
- Layered architecture
- Validation
- Database persistence
- Connection pooling
- Error handling

### Go
- Goroutines
- Channels
- Context
- Worker pools
- Concurrency
- Graceful shutdown

### Databases
- PostgreSQL
- Transactions
- Row locking
- JSONB
- Indexes
- ACID
- Isolation
- `FOR UPDATE SKIP LOCKED`

### Distributed Systems
- Message queues
- Kafka
- Consumer groups
- Partitions
- Retries
- Exponential backoff
- Dead Letter Queues
- Idempotency
- Distributed locks
- Worker failure recovery
- Horizontal scaling

### DevOps
- Docker
- Docker Compose
- Kubernetes
- Autoscaling
- Prometheus
- Grafana
- CI/CD

---

# 15. Key Interview Questions to Prepare

Before putting this project strongly on a resume, be able to answer:

1. Why did you build a distributed task scheduler?
2. Why Go?
3. Why PostgreSQL?
4. Why Kafka instead of polling PostgreSQL?
5. Why Redis?
6. What does `FOR UPDATE SKIP LOCKED` do?
7. How do two workers avoid processing the same job?
8. What happens if a worker crashes during execution?
9. How do retries work?
10. Why use exponential backoff?
11. What is a Dead Letter Queue?
12. What is idempotency?
13. What happens if Kafka delivers the same message twice?
14. At-most-once vs at-least-once delivery?
15. Why can't Kafka directly provide strict priority scheduling?
16. How would you process one million jobs?
17. What is the system bottleneck?
18. How would you horizontally scale workers?
19. What happens when PostgreSQL goes down?
20. How would you monitor the system?
21. REST vs gRPC?
22. How would Kubernetes help?
23. How would you prevent low-priority job starvation?
24. How would you safely cancel a running job?
25. How would you measure throughput and latency?

---

# 16. Resume Goal

After the full project is implemented and benchmarked, a resume bullet can be based on real measured results, for example:

> Built a fault-tolerant distributed task-processing platform in Go using Kafka, Redis, PostgreSQL, and Kubernetes, implementing priority scheduling, concurrent workers, retries, idempotency, failure recovery, REST/gRPC APIs, and horizontal scaling.

Do **not** add throughput, latency, availability, or scale numbers until they have actually been measured.

---

## Current Development Status

```text
Core API                    COMPLETE
PostgreSQL Persistence      COMPLETE
Priority Worker             COMPLETE
Safe Multi-Worker Claiming  COMPLETE
Retry Logic                 COMPLETE
Exponential Backoff         COMPLETE
DEAD Job Handling           COMPLETE

Kafka                       TODO
Redis                       TODO
Idempotency                 TODO
Dead Letter Queue           TODO
Worker Heartbeats           TODO
Crash Recovery              TODO
gRPC                        TODO
Docker                      TODO
Kubernetes                  TODO
Prometheus/Grafana          TODO
Testing                     TODO
Load Testing                TODO
CI/CD                       TODO
```

---

## License

Add a license before public/reusable distribution if desired.
