# End-to-End Real-Time Financial Data Platform

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Powered-blue)

An end-to-end, event-driven data platform built in Go. This system ingests live financial data from a WebSocket, streams it through a Kafka message bus, processes and stores it in MongoDB, and exposes it via a clean, well-tested REST API.
## System Architecture

The project is a multi-microservice application, fully containerised with Docker, demonstrating a decoupled, scalable, and resilient architecture.

```mermaid
graph TD
    subgraph External
        A[Live Finnhub WebSocket]
    end

    subgraph "Financial Data Platform (Containerized via Docker)"
        subgraph "Ingestion Service"
            B(Go Ingestor)
        end
        subgraph "Message Bus"
            C[Apache Kafka]
        end
        subgraph "Processing Service"
            D(Go Processor)
        end
        subgraph "Data Storage"
            E[MongoDB]
        end
        subgraph "API Service"
            F(Go REST API)
        end
    end

    subgraph Client
        G[User / Frontend]
    end

    A -- Real-Time Trade Data --> B
    B -- Publishes Raw Messages --> C
    D -- Subscribes to Topic --> C
    D -- Persists Data --> E
    F -- Queries Data --> E
    G -- Makes HTTP Requests --> F

    style A fill:#f9f,stroke:#333
    style C fill:#fda,stroke:#333
    style E fill:#aaf,stroke:#333
```

## Key Features & Technical Highlights

This project demonstrates several advanced, production-ready engineering concepts:

*   **Decoupled & Resilient Architecture**: The Ingestion and Processing services are separated. **Apache Kafka** acts as a durable buffer, ensuring that if the database is slow or temporarily unavailable, no incoming real-time data is lost.

*   **Idempotent Data Processing**: The Go Processor is designed to be idempotent. It generates a **deterministic key** from Kafka message metadata (`topic-partition-offset-index`) for each trade. This allows MongoDB's unique index to automatically reject duplicate writes, guaranteeing that data is processed **exactly once**, even if a message is redelivered after a crash.

*   **Concurrent-Safe Metadata Updates**: To avoid race conditions when scaling the processor, all metadata updates are performed within a single `UpdateOne` command. **MongoDB guarantees atomicity for operations that modify a single document.** By using atomic operators like `$inc` and `$max` together in one operation, we ensure that `tradeCount` and ``lastTradeAt`` are updated as a single, uninterruptible unit, making data more consistent without complex application-level locking.

*   **Clean Architecture API**: The REST API is built following Clean Architecture principles, separating concerns into `Handler`, `Usecase (Service)`, and `Repository` layers. This makes the code organized, maintainable, and highly testable.

*   **Comprehensive Testing Strategy**: The project is validated with a multi-layered testing approach, demonstrating a commitment to code quality and reliability:
    *   **Unit Tests** for core business logic, achieving **100% statement coverage** on the API's Usecase layer and **82% coverage** on the Processor's critical data transformation logic.
     *  **Unit and Integration Tests for the API's Request Layer**, validating the complete interaction between the Handler and Usecase and the Middleware chain. These tests achieved **96% coverage** on the middleware package and confirmed the system's resilience by verifying robust error handling and request timeout behavior.
    *   **Integration Tests** for the Repository layer, achieving **82% coverage** by running against a real test database to verify MongoDB query behaviour. This includes a specific test to validate the **idempotency strategy**, proving that the unique `message_key` index successfully prevents duplicate data insertion on message re-delivery.

## Tech Stack

| Category      | Technology                                             |
|---------------|--------------------------------------------------------|
| **Language**  | Go                                                     |
| **Pipeline**  | Apache Kafka, Gorilla WebSocket                        |
| **API**       | Gin (Web Framework)                                    |
| **Database**  | MongoDB (with Time Series Collections)                 |
| **Testing**   | Go Standard Library, `testify`, `mockery`              |
| **DevOps**    | Docker, Docker Compose                                 |

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
        { "symbol": "AAPL", "trade_count": 12345, "last_trade_at": "2025-11-20T15:00:01Z" },
        { "symbol": "TSLA", "trade_count": 67890, "last_trade_at": "2025-11-20T15:00:00Z" }
      ]
    }
  }
  ```

#### Get Latest Trades for a Symbol
- **Endpoint**: `GET /api/v1/trades/:symbol/latest`
- **Description**: Returns a paginated list of the most recent trades for a symbol using efficient cursor-based pagination.
- **Query Parameters**: `limit` (int), `before` (Unix ms timestamp)
- **Example Response**:
  ```json
  {
    "data": {
      "data": [
        { "timestamp": "2025-11-20T15:00:01.123Z", "price": "200.05", "volume": "50.0" }
      ],
      "pagination": {
        "nextCursor": 1763664001123
      }
    }
  }
  ```

## Getting Started

### Prerequisites
*   Go (1.21+)
*   Docker & Docker Compose

### 1. Configuration

Create a `config/config.yml` file in the project root with the following structure, filling in your credentials:

```yml
finnhub:
  token: YOUR_FINNHUB_TOKEN

kafka:
  broker_url: "kafka:29092"
  topic: "raw_stock_ticks"

mongodb:
  url: "mongodb+srv://USER:PASS@your_cluster.mongodb.net/"
  databaseName: "StockFeedDatabase"
  collectionName: "finnhub_trades"
  symbolsCollectionName: "symbols"

api:
  port: "8000"

timeouts:
  api_request: "5s"
  processor_db_operation: "15s"
  shutdown: "5s"

subscribed_symbols:
  - "AAPL"
  - "GOOGL"
  - "TSLA"
  - "MSFT"
  - "NVDA"
  - "AMD"
```

### 2. Run the Application

From the project root, start the entire platform with a single command:
```bash
docker compose up --build
```
The API will be available at `http://localhost:8000`.

## Running Tests

The project includes a comprehensive test suite. To run all tests, you first need to provide a connection string for a test database in a `.env` file.

1.  **Create a `.env` file** in the project root:
    ```env
    # .env
    MONGO_URL_TEST="mongodb+srv://USER:PASS@your_cluster.mongodb.net/"
    ```

2.  **Run the tests** with the coverage flag:
    ```bash
    go test ./... --cover
    ```

## Future Improvements

*   **Data Reconciliation Service**: Build a background cron job to periodically recalculate the `symbols` metadata from the raw `finnhub_trades` data, ensuring long-term consistency in the eventually consistent model.

