# Ganesh.provengo.io

![Ganesh Processing Messages](./assets/ganesh.provengo.io.webp)


## Overview
Ganesh.provengo.io is a backend project designed to showcase the integration and efficient use of key technologies, such as **NATS**, **Postgres 17**, and **Redis (or DragonflyDB)**, through the implementation of **channels** and **goroutines** in **Golang**. This project is part of my public portfolio on GitHub, serving as an educational example on how to design scalable systems capable of processing more than **20,000 messages per second**.

## Purpose
This project is designed for **learning and testing** purposes. It highlights the synergy between **message brokers**, **in-memory databases**, and **persistent databases** in a high-performance environment, demonstrating how backend services can efficiently handle massive message flows.

## Key Technologies
- **Golang:** Core language to implement business logic using channels and goroutines for concurrency.
- **NATS:** A lightweight and high-performance message broker used for asynchronous communication between services.
- **Postgres 17:** Relational database to store structured data persistently.
- **Redis / DragonflyDB:** In-memory key-value store optimized for low-latency, frequently accessed data.

## Project Architecture
1. **Message Queue with NATS:** Messages are produced and consumed asynchronously using NATS, demonstrating its capability to handle high-throughput event streams.
2. **Concurrent Processing with Goroutines and Channels:** Goroutines are utilized to process incoming messages in parallel, ensuring high efficiency and non-blocking execution.
3. **Data Caching with Redis/DragonflyDB:** Frequently accessed data is stored in Redis or DragonflyDB to reduce load on Postgres and provide ultra-fast access.
4. **Persistent Storage in Postgres 17:** Data that requires persistence and relational access is stored in Postgres.

## Features
- Asynchronous message publishing and consumption using NATS.
- High concurrency enabled by Golang's goroutines and channels.
- Efficient caching mechanism using Redis/DragonflyDB.
- Durable storage and advanced queries through Postgres 17.
- Capable of handling **20,000+ messages per second** in test environments.
- Redis driver singleton
- Postgres driver singleton

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/willianpsouza/ganesh.provengo.io.git
   cd ganesh.provengo.io
   ```
2. Install dependencies using **Go modules**:
   ```bash
   go mod tidy
   ```
3. Start required services using Docker Compose:
   ```bash
   docker-compose up -d
   ```
4. For more control inside "scripts/infrastructure" has scripts to starts each component.
```
├── scripts
│   ├── infrastructure           # INFRASTRUCTURE Directory
│   │   ├── postgres             # Postgres database interactions
│   │   ├── nats                 # NATS message broker
│   │   └── dragonfly            # Redis/DragonflyDB cache logic
   
```

## Configuration
Modify the `.env` file to configure:
- NATS connection details
- Postgres database URL
- Redis/DragonflyDB connection details

Example `.env`:
```
NATS_URL=nats://localhost:4222
POSTGRES_URL=postgresql://user:password@localhost:5432/ganeshdb
REDIS_URL=redis://localhost:6379
```

## Usage
1. Run the application:
   ```bash
   go run main.go
   ```
2. Publish test messages using the sample producer provided:
   ```bash
   go run producer/main.go
   ```
3. Monitor logs to observe message processing:
   ```bash
   tail -f logs/app.log
   ```

## Project Structure
```
.
├── cmd
│   └── main.go             # Entry point of the application
├── internal
│   ├── messaging           # NATS message handlers
│   ├── storage             # Postgres database interactions
│   └── cache               # Redis/DragonflyDB cache logic
├── producer
│   └── main.go             # Sample message producer
├── docker-compose.yml      # Service configuration
├── .env                    # Environment configuration
├── go.mod                  # Go modules
└── README.md               # Project documentation
```

## Performance Testing
The system is designed to handle over **20,000 messages per second** using:
- NATS for fast message propagation.
- Goroutines and channels for parallel processing.
- Redis/DragonflyDB for quick data access.
- Postgres for long-term data persistence.

To run performance tests:
```bash
go test -bench ./tests
```

## Future Enhancements
- Add monitoring using **Prometheus** and **Grafana**.
- Integrate **Jaeger** for distributed tracing.
- Explore scaling options using **Kubernetes**.

## License
This project is licensed under the MIT License.

## Author
**Provengo.io**  
Feel free to reach out with suggestions, feedback, or to discuss improvements!


