#  Distributed Video Transcoding Engine

A scalable, distributed, asynchronous video transcoding pipeline built in **Go** using **Gin**, **RabbitMQ**, **AWS SQS**, **PostgreSQL (GORM)**, and **AWS S3**, powered by **FFmpeg**.

It converts uploaded videos into multiple streaming-optimized resolutions (360p, 480p, 720p) using a fault-tolerant worker-based architecture.

Designed with production-grade patterns: event-driven ingestion, queue buffering, horizontal worker scaling, retry handling, and asynchronous job tracking.

---

##  High-Level Architecture

This system is built to decouple **video upload**, **event ingestion**, and **CPU-heavy transcoding workloads**.

###  Processing Flow

1. **Client Upload Request**
   - Client calls `POST /upload`
   - Backend generates:
     - `job_id`
     - S3 presigned upload URL
   - Client uploads video directly to **S3 input bucket**

2. **S3 Event Trigger**
   - S3 upload event triggers **AWS SQS message**

3. **Event Consumer Layer**
   - SQS consumer:
     - Validates event
     - Creates DB record (`pending`)
     - Pushes job into **RabbitMQ quorum queue**

4. **Transcoding Workers**
   - Worker pool consumes RabbitMQ messages
   - Executes pipeline:
     - Validate file using `ffprobe`
     - Transcode using `ffmpeg`
     - Generate multiple renditions:
       - 360p
       - 480p
       - 720p
     - Upload outputs to **S3 output bucket**
     - Update job status in PostgreSQL

---

## System Design Goals

- Fully asynchronous video processing
-  Queue-based buffering for reliability
-  Separation of ingestion and compute layers
-  Horizontally scalable worker pool
-  Fault tolerance with retry + DLQ strategy
-  Cloud-native storage using AWS S3
-  Trackable job lifecycle via PostgreSQL

---

##  Tech Stack

- **Language:** Go (1.25+)
- **Web Framework:** Gin
- **Database:** PostgreSQL (via GORM)
- **Queue Systems:**
  - AWS SQS (event ingestion)
  - RabbitMQ (quorum queue job processing)
- **Object Storage:** AWS S3
- **Media Processing:** FFmpeg + FFprobe
- **Testing:** k6 load testing

---

##  Project Structure

```text
├── cmd/
│   └── main.go                 # Entry point
│
├── internal/
│   ├── configuration/          # Viper config loader
│   ├── database/               # PostgreSQL + migrations
│   ├── handler/                # HTTP handlers (upload, status, health)
│   ├── model/                  # DB models (Job schema)
│   ├── pipeline/               # Orchestrates transcoding pipeline
│   ├── rabbitmq/               # Queue publisher/consumer logic
│   ├── sqs/                    # AWS SQS event listener
│   ├── storage/                # AWS S3 client wrapper
│   ├── transcoder/             # FFmpeg abstraction layer
│   └── worker/                # Worker pool implementation
│
├── docker-compose.yml          # Local infra (Postgres + RabbitMQ)
├── test.js                     # k6 load testing script
├── .env.example                # Environment configuration template
└── README.md

```

##  Getting Started

### 1. Start Infrastructure
```bash
docker-compose up -d
```

### 2. Verify Dependencies
Ensure you have FFmpeg installed locally for the application workers:

```Bash
ffmpeg -version
ffprobe -version
```
### 3. Run Application
```Bash
go mod download
go run cmd/main.go

```
## API Reference
### 1. Generate Upload URL
Creates a presigned S3 URL for direct client upload.

Endpoint: POST /upload

Response:

```JSON
{
  "message": "Upload successfully",
  "job_id": "uuid",
  "upload_url": "https://s3-presigned-url...",
  "key": "originals/uuid.mp4"
}
```

### 2. Get Job Status
Endpoint: GET /status/:job_id

Responses by State:
Pending:

```JSON
{
  "job_id": "uuid",
  "status": "pending"
}
```
Completed:

```JSON
{
  "job_id": "uuid",
  "status": "completed",
  "outputs": {
    "360p": "https://s3/..._360p.mp4",
    "480p": "https://s3/..._480p.mp4",
    "720p": "https://s3/..._720p.mp4"
  }
}
```
Failed:

```JSON
{
  "job_id": "uuid",
  "status": "failed",
  "error": "ffmpeg encoding failed"
}
3. Health Check
Endpoint: GET /health
```

Response:

```JSON
{
  "message": "Server is healthy"
}
```
## Internal Design Notes
 ### 1) Worker Model
Concurrent Worker Pool: Configurable pool size running in Go.
Isolation: Each worker pulls a job from RabbitMQ, executes an isolated OS-level FFmpeg process, and guarantees per-job temporary workspace cleanup upon completion or failure.

### 2) Queue Strategy
AWS SQS: Acts as the ingestion buffer to decouple direct S3 event notifications.
RabbitMQ (Quorum Queues): Handles internal high-availability job processing, durable message persistence, and built-in poison message handling.

 ### 3) Failure Handling
Retry Policy: Automatic retry mechanism utilizing exponential backoff up to a MAX_RETRIES threshold.
Dead-Lettering: Built-in DLQ-ready architecture for unresolvable processing faults.
State Machine Transitions: State is strictly managed in the database via atomic transitions:

### Storage Strategy
Input Bucket: Dedicated to raw user uploads.
Output Bucket: Stores the successfully transcoded video variants.

Naming Convention:

Plaintext
outputs/{job_id}_360p.mp4
outputs/{job_id}_480p.mp4
outputs/{job_id}_720p.mp4

### Load Testing
The system utilizes k6 to simulate high-concurrency uploads and stress-test the pipeline end-to-end.

### Requirements
Place a sample file named test.mp4 into the project root directory.

Run Test
```Bash
k6 run test.js
```

### Scalability Strategy
This system architecture is built to scale horizontally out of the box:

Horizontal Worker Scaling: Add more independent RabbitMQ consumer instances/pods to handle higher transcoding loads.

Storage Throughput: S3 naturally scales to handle massive concurrent read/write throughput.

Database Optimization: Read replicas can be introduced to offload telemetry and status check traffic.

Backpressure Management: The queue architecture acts as a natural buffer, ensuring spikes in video uploads never overwhelm the underlying compute layer.

### Future Improvements
1) Hardware Acceleration: GPU-based transcoding integration via NVENC.
2) Adaptive Bitrate Streaming: Implementation of HLS/DASH segmentation and manifest generation.
3) Edge Distribution: CDN integration (e.g., AWS CloudFront) for optimal video playback delivery.
4) Per-Title Encoding: Smart, content-aware bitrate optimization.
5) Resilient Ingestion: Chunked upload and resume support for large files.
6) Priority Queuing: Multi-tier job priority scheduling.
7) AI-Assisted Processing: Automatic thumbnail generation and content tagging.


***
