# End-to-End Real-Time Financial Data Platform

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![Python Version](https://img.shields.io/badge/Python-3.11+-yellow.svg)
![Docker](https://img.shields.io/badge/Docker-Powered-blue)

![](docs/client_demo.gif)
> **Demo:** Real-time terminal dashboard (TUI) visualising high-frequency market data and **$O(1)$ Rolling VWAP** metrics, streamed via **TCP Sockets**.


An end-to-end, event-driven data platform built in **Go** and **Python**. This system ingests live financial data from WebSockets and streams it through **Apache Kafka** for two uses:

1. **Go Service:** stores time-series data in **MongoDB** and exposes it via a clean, well-tested REST API.
2. **Python Analytics Engine:** leverages **AsyncIO** to calculate **rolling variation of VWAP** in $O(1)$ time, broadcasting real-time metrics to clients via **TCP Sockets.**

## Table of Contents
- [Live Demo](#live-demo-of-api)
- [System Architecture](#system-architecture)
- [Key Features & Technical Highlights](#key-features--technical-highlights)
  - [1. Availability & Scalability](#1-availability--scalability)
  - [2. Data Consistency & Integrity](#2-data-consistency--integrity)
  - [3. API Performance & Design](#3-api-performance--design)
  - [4. Real-Time Analytics Engine (Python)](#4-real-time-analytics-engine-python)
  - [5. DevOps & Quality Assurance](#5-devops--quality-assurance)
- [Tech Stack](#tech-stack)
- [API Endpoints](#api-endpoints)
- [Getting Started](#getting-started)
  - [Configuration](#1-configuration)
  - [Run the Application](#2-run-the-application)
  - [Run the Analytics Client](#3-run-the-real-time-analytics-client)
- [Running Tests](#running-tests)
- [Production Deployment (AWS)](#production-deployment-aws)
- [Horizontal Scalability](#demonstrating-horizontal-scalability)
- [Future Improvements](#future-improvements)

## Live Demo of API
**Base URL:** [http://3.26.92.41/api/v1/trades/AAPL](http://3.26.92.41/api/v1/trades/AAPL)
* (Deployed on AWS EC2. Returns live JSON data from the cloud database.)
* See the [API Endpoints](#api-endpoints) section below for full documentation on how to filter and paginate trades.

## System Architecture

The project is a multi-microservice application, fully containerised with Docker, demonstrating a decoupled, scalable, and resilient architecture.

```mermaid
graph TD
    subgraph External
        A[Live Finnhub WebSocket]
    end

    subgraph "Financial Data Platform"
        B(Go Ingestor)
        subgraph "Message Bus"
            C[Apache Kafka]
        end
        subgraph "Go Service"
            subgraph "Horizontally Scalable"
                D(Go Processor)
            end
            subgraph "Data Storage"
                E[MongoDB]
            end
            subgraph "Cloud API Service"
                F(AWS REST API)
            end
        end
        subgraph "Python Analytics Layer"
            H(Python Analytics Engine)
            I(TCP Server :8888)
        end
    end

    subgraph Client
        G[User / Frontend / Postman]
        J["Python TCP Terminal Client (TUI)"]
    end

    A -- Real-Time Trade Data --> B
    B -- Publishes Raw Messages --> C
    C -- Consumer Group A --> D
    D -- Persists Data --> E
    F -- Queries Data --> E
    G -- Makes HTTP Requests --> F

    C -- Consumer Group B --> H
    H -- O(1) Rolling VWAP --> H
    H -- Broadcasts Stream --> I
    J -- TCP Connect --> I
```

## Key Features & Technical Highlights

This project was architected to satisfy many Non-Functional Requirements (NFRs).
### 1. Availability & Scalability
*Designed to handle growth and maintain high availability.*
*   **Decoupled & Resilient Architecture**: The Ingestion and Processing services are fully decoupled using **Apache Kafka**. This acts as a durable buffer, ensuring that if the database is slow or temporarily unavailable, no incoming real-time data is lost.
*   **Horizontal Scalability via Consumer Groups**: The processing service is designed to be scaled out. Kafka's **consumer group** model guarantees that each message is delivered to exactly one processor instance, enabling safe, parallel processing of the data stream without duplication.
*   **Hybrid Cloud Deployment (Cost & Performance)**: To optimise resource usage, the system employs a hybrid strategy. The memory-intensive **Data Pipeline** (Ingestor, Kafka, Processor) runs on local infrastructure but writes directly to a centralised **MongoDB Atlas** cloud database. The **API Service** is deployed to a lightweight **AWS EC2 instance**, connecting to that same cloud database. This decouples the heavy processing from the query layer, ensuring the API remains available 24/7 via the public internet, accessible from anywhere, regardless of the state of the local ingestion pipeline.
### 2. Data Consistency & Integrity
*Ensuring data is durable, accurate, and safe during failures.*

*   **Deliberate Pivot to Eventual Consistency**: The initial design aimed for perfect atomicity using transactions. However, discovering that **MongoDB's Time Series engine does not support inserts within transactions** forced a deliberate architectural pivot. The system now prioritises the absolute durability of the raw trade data, updating aggregated metadata on a best-effort, eventually consistent basis.
*   **Idempotent Processing for Crash Recovery**: To prevent data duplication if the processor crashes and re-reads a message, the system generates a **deterministic idempotency key** from Kafka metadata (`topic-partition-offset--symbol-timestamp-index`). MongoDB's unique index rejects duplicate writes, guaranteeing **exactly-once** persistence logic.
*   **Graceful Shutdown**: The stateful `go-processor` catches `SIGINT` or `SIGTERM` signals. It finishes processing its in-flight Kafka message and commits the offset before exiting, ensuring **at-least-once** delivery is handled cleanly during deployments.
*   **Concurrent-Safe Metadata Updates**: To support horizontal scaling, the system handles concurrent writes to the same symbol metadata.
    1.   **`$inc`**: Used for the trade count to ensure every trade is counted, even if multiple processors update the same symbol simultaneously.
    2.  **`$max`**: Used for the `lastTradeAt` timestamp. This solves the "out-of-order write" race condition, ensuring the timestamp only moves forward to a later time and never regresses, even if an older message is processed last.
    3.   **Atomicity**: Because MongoDB guarantees atomicity for single-document updates, combining these into one operation guarantees concurrency safety.
### 3. API Performance & Design
*Optimising latency and developer experience.*

*   **Cursor-Based Pagination**: To efficiently paginate high-frequency time-series data, the `/trades` endpoint avoids slow `limit/offset` queries. Instead, it implements **cursor-based pagination**, which offers two critical advantages:
    1.  **Index-Optimised Performance**: By leveraging MongoDB's **Time Series index** on `symbol` (metadata field) and `time` (time field), the database allows for a fast "seek" to the correct position in the dataset.
    2.  **Data Stability**: In a live system where new trades are constantly inserted at the top of the list, traditional pagination causes "page drift" (skipping or repeating records). Using a timestamp cursor ensures that historical pages remain stable and consistent for the client, regardless of incoming live traffic.
*   **Robust Middleware Chain**: The API is protected by custom Gin middleware. A **Timeout Middleware** actively races handlers against a timer to prevent slow DB queries from exhausting server resources. A **Centralised Error Middleware** captures all failures (including timeouts) to ensure the API always returns a consistent, predictable JSON error response.
*   **Clean Architecture**: The codebase follows Clean Architecture principles (`Handler` -> `Usecase` -> `Repository`), separating concerns to ensure the system is maintainable and easy to extend.

### 4. Real-Time Analytics Engine (Python)
*Demonstrating network programming and algorithmic efficiency.*

- **Polyglot Microservices Pattern**: Integrated a Python service alongside the Go architecture. It consumes the same Kafka `raw_stock_ticks` topic using a separate Consumer Group, enabling parallel, decoupled processing of market data without affecting the primary Go pipeline.
- **AsyncIO & TCP Networking**: Built a TCP server using Python's `asyncio` library. It manages concurrent client connections via an event loop and broadcasts real-time metrics via **TCP Sockets** to a **Live Terminal Dashboard (TUI)** built with `rich`.
* **O(1) Rolling VWAP Algorithm**: Implemented a Rolling variation of the **Volume Weighted Average Price**, calculating the average price over a fixed window of recent **ticks** rather than resetting at the start of a trading session.
    *   **Data Structure**: Utilised `collections.deque` to maintain the window state.
    *   **Efficiency**: Achieved constant time ($O(1)$) complexity by maintaining running totals—adding the incoming tick's value and subtracting the evicted tick's value—eliminating the need to iterate over the dataset on every update.
- **Exchange Simulation Mode**: Engineered a standalone **Mock Engine** to simulate high-frequency market conditions (synthetic ticker generation). This allows for offline development and testing of clients without being constrained by external API rate limits.

### 5. DevOps & Quality Assurance
*Ensuring reliability through automation.*

*   **Automated CI Pipeline**: A **GitHub Actions** workflow serves as a strict quality gate, automatically running the full test suite on every push to `main`. This prevents regression and ensures that code works as intended on a neutral environment (as opposed to just local environment.)
*   **Comprehensive Testing Strategy**: The project is validated with a multi-layered strategy:
    1.  **Unit Tests**: 100% coverage on API business logic (Usecase) and 82% on Processor transformation logic.
    2.  **Request Layer Integration**: Validates the full interaction chain between Middleware, Handler and Usecase (96% middleware coverage).
    3.  **Repository Integration**: Runs against a **live MongoDB test database** to verify query behaviour and validate the **idempotency strategy** (82% repository coverage).
    4. **Simulation Testing**: Utilises the Python **Mock Engine** to generate deterministic, high-frequency synthetic data streams, verifying that the TUI remains responsive and flicker-free under rapid update cycles.

## Tech Stack

| Category           | Technology                                            |
| ------------------ | ----------------------------------------------------- |
| **Language**       | Go (1.24), Python (3.11)                              |
| **Pipeline**       | Apache Kafka, Gorilla WebSocket, AsyncIO (Python)     |
| **API**            | Gin (REST), TCP Sockets                           |
| **Database**       | MongoDB (with Time Series Collections)                |
| **Testing**        | Go: Std Lib, `testify`, `mockery`; Python: `unittest` |
| **CI/CD**          | GitHub Actions (Automated Testing)                    |
| **Infrastructure** | Docker, Docker Compose, AWS (EC2)                     |

## API Endpoints

The API provides the following MVP endpoints for data consumption:

#### Get All Tracked Symbols
- **Endpoint**: `GET /api/v1/symbols`
- **Description**: Returns metadata for all symbols the system has processed.
- **Example Response**:
  ```json
  {
      "data": {
          "available": [
              {
                  "symbol": "AMD",
                  "trade_count": 8,
                  "last_trade_at": "2025-11-20T12:42:48.093Z"
              },
              {
                  "symbol": "NVDA",
                  "trade_count": 7,
                  "last_trade_at": "2025-11-20T12:39:40.831Z"
              },
              {
                  "symbol": "TSLA",
                  "trade_count": 5,
                  "last_trade_at": "2025-11-20T12:41:40.175Z"
              }
          ]
      },
      "error": null,
      "message": null
  }
  ```

#### Get Latest Trades for a Symbol
- **Endpoint**: `GET /api/v1/trades/:symbol`
- **Description**: Returns a paginated list of the most recent trades for a symbol using efficient cursor-based pagination.
- **Query Parameters**: `limit` (int), `before` (Unix ms timestamp)
- **Example Response**:
  ```json
  {
      "data": {
          "data": [
              {
                  "timestamp": "2025-11-20T13:33:18.585Z",
                  "price": "196.38",
                  "volume": "265"
              },
              {
                  "timestamp": "2025-11-20T13:33:12.457Z",
                  "price": "196.29",
                  "volume": "100"
              },
              {
                  "timestamp": "2025-11-20T13:33:08.27Z",
                  "price": "196.39",
                  "volume": "115"
              },
              {
                  "timestamp": "2025-11-20T13:33:06.456Z",
                  "price": "196.49",
                  "volume": "100"
              },
              {
                  "timestamp": "2025-11-20T13:33:03.544Z",
                  "price": "196.6",
                  "volume": "300"
              }
          ],
          "pagination": {
              "next_cursor": 1763645583544
          }
      },
      "error": null,
      "message": null
  }
  ```

## Getting Started

### Prerequisites
*   Go (1.21+)
*   Docker & Docker Compose

### 1. Configuration

Create a `config/config.yml` file in the project root with the following structure, filling in your credentials:

```yml
api_port: "8000"

finnhub:
  token: YOUR_FINNHUB_TOKEN

kafka:standard address.
  broker_url: "kafka:29092" 
  topic: "raw_stock_ticks"

subscribed_symbols:
  - "AAPL"
  - "GOOGL"
  - "TSLA"
  - "MSFT"
  - "NVDA"
  - "AMD"

mongodb:
  url: YOUR_MONGODB_ATLAS_URL
  database_name: "financialDataDatabase"
  collection_name: "finnhub_trades"
  symbols_collection_name: "symbols"

timeouts:
  # For user-facing API requests. Should be short.
  api_request: "5s"
  # For background network operations (Kafka writes, DB writes). Can be longer.
  background_operation: "15s"
  # For graceful shutdown of services.
  shutdown: "5s"
```

### 2. Run the Application

From the project root, start the entire platform with a single command:
```bash
docker compose up --build
```
The API will be available at `http://localhost:8000` (if run locally.)

### 3. Run the Real-Time Analytics Client
While `docker compose up` starts the backend microservices, the Python **TCP Client** is designed to run interactively in your terminal to monitor the data stream.

1.  Ensure the Docker stack (at least the Go Ingestor and Python analytics engine) is running.
2.  Open a new terminal window.
3.  Launch the terminal-based client to visualise live price feeds and VWAP calculations.
    ```bash
    # Requires Python 3 installed locally
    python python-analytics/client.py
    ```
    *You will see a live feed of trades and calculated VWAP streaming from the engine.*

#### Quick Start (Simulation Mode)
Want to see the TUI in action without setting up Kafka or Docker? Run the included mock engine:

1.  **Start the Mock Engine:**
    ```bash
    python python-analytics/mock_engine.py
    ```
2.  **Start the Client:**
    ```bash
    python python-analytics/client.py
    ```

## Running Tests

The project includes a comprehensive test suite. To run all tests, you first need to provide a connection string for a test database in a `.env` file.

1.  **Create a `.env` file** in the project root:
    ```env
    # .env
    MONGO_URL_TEST="..."
    ```

2.  **Run the tests** with the coverage flag:
    ```bash
    go test ./... --cover
    ```

## Production Deployment (AWS)
The API service is designed to be deployed to an AWS EC2 instance. 

**Note on t3.micro:**  
Currently, we cannot compile the Go binary on the server itself (it will crash). Instead, we **build the Docker image locally**, push it to a registry (Docker Hub), and simply pull/run it on the server.

### Prerequisites
1.  **Docker Hub Account.**
2.  **Compose File:** Update your `docker-compose.prod.yml` to use `image:` instead of `build:`:
    ```yaml
    services:
      go-api-service:
        image: <YOUR_DOCKERHUB_USERNAME>/go-api-service:latest
        # build: ... (Remove build section)
    ```

---

### 1. Build and Push (Local Machine)
Run these commands from your project root:

```bash
# 1. Build the image
# Replace <YOUR_USER> with your Docker Hub username
docker build --platform linux/amd64 -t <YOUR_USER>/go-api-service:latest -f cmd/go-api-service/Dockerfile .

# 2. Push image to Docker Hub
docker push <YOUR_USER>/go-api-service:latest
```

### 2. Provision Infrastructure (AWS)
1.  Launch a `t3.micro` instance (Ubuntu).
2.  Configure **Security Group** (Inbound Rules):
    *   **SSH:** Your IP Address.
    *   **HTTP:** `0.0.0.0/0`.
3.  Download your PEM key (e.g., `...pem`).

### 3. Server Setup (Remote)
Connect to the instance and set up the environment.

```bash
chmod 400 ...pem

# SSH into server
ssh -i "...pem" ubuntu@<NEW_IP_ADDRESS>
```

Once inside the server, run the following to install Docker:

```bash
# add 2G of swap space
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make swap file permanent (but backup fstab first)
sudo cp /etc/fstab /etc/fstab.bak
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# --- 2. Install Docker & Docker Compose ---
sudo apt-get update -y
sudo apt-get install -y docker.io docker-compose
```

---

### 4. Deploy Configuration (Local Machine)
Instead of copying the source code, we only copy the **Compose file** and the **Config file**. Run this from your local computer:

```bash
# 1. Create the application directory on the server
ssh -i "...pem" ubuntu@<NEW_IP_ADDRESS> "mkdir -p ~/app/config"

# 2. Upload the Production Compose file
scp -i "...pem" docker-compose.prod.yml ubuntu@<NEW_IP_ADDRESS>:~/app/docker-compose.yml

# 3. Upload the Config file
scp -i "...pem" config/config.yml ubuntu@<NEW_IP_ADDRESS>:~/app/config/config.yml
```

---

### 5. Launch Service (Remote)
Connect back to the server to start the application.

```bash
ssh -i "...pem" ubuntu@<NEW_IP_ADDRESS>

# Navigate to app folder
cd ~/app

# (Optional) Login
# sudo docker login

# Run the service
# -d: Detached mode (background)
# Note: We do NOT use '--build' here
sudo docker-compose up -d
```
#### Critical: Verify Database Connection
After starting the service, you **must** check the logs to ensure the EC2 instance can connect to MongoDB Atlas. Atlas blocks unknown IPs by default.

1.  **Check the logs:**
    ```bash
    sudo docker logs -f go-api-service
    ```

2.  **If you see errors** like `Failed to connect to MongoDB:`:
    This likely means MongoDB Atlas is blocking your EC2 IP.

3.  **Get your EC2 IP Address:**
    Run this command in the terminal:
    ```bash
    curl ifconfig.me
    ```

4.  **Whitelist in MongoDB Atlas:**
    *   Log in to [MongoDB Atlas](https://cloud.mongodb.com/).
    *   Go to **Security** $\rightarrow$ **Database & Network Access**.
    *   Click **Add IP Address**.
    *   Paste the IP from step 3 and confirm.

5.  **Restart the Service:**
    Once the IP is active in Atlas, restart the container to reconnect:
    ```bash
    sudo docker restart go-api-service
    ```

## Demonstrating Horizontal Scalability

This architecture is designed to scale horizontally. You can run multiple instances of the `go-processor` to handle higher data loads, and Kafka's consumer group will automatically distribute the work.

#### 1. Prepare `docker-compose.yml`

First, edit your `docker-compose.yml` file. To run multiple instances, you must **remove or comment out the `container_name` line** from the `go-processor` service.

```yml
  go-processor:
    # container_name: go-processor  <-- Remove this line
    build: ...
```

#### 2. Run with the `--scale` Command

From your project root, start the application and scale the processor to three instances with this command:

```bash
docker compose up --build --scale go-processor=3
```

#### 3. Verify Load Balancing

To see the processors working in parallel, open a new terminal and follow their logs:

```bash
docker compose logs -f go-processor
```

You will see logs from different instances (e.g., `go-processor-1`, `go-processor-2`) processing different Kafka partitions, confirming that the load is being shared.

## Future Improvements
*   **Automated CD Pipeline**: Extend the GitHub Actions workflow to implement full Continuous Deployment, to AWS.
*   **Data Reconciliation Service**: Build a background cron job to periodically recalculate the `symbols` metadata from the raw `finnhub_trades` data, ensuring long-term consistency in the eventually consistent model.

