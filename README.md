# Kafka Log Aggregation System

A real-time log aggregation system built with **Apache Kafka** and **Go** that demonstrates core distributed streaming concepts. This project simulates multiple applications generating logs and aggregating them through Kafka topics for centralized processing.

## 🎯 Project Overview

This system consists of:

- **Log Producer**: Simulates multiple applications (UserService, DatabaseService, AuthService, PaymentService) generating realistic log messages
- **Log Consumer**: Processes and displays logs in real-time with color-coded output based on log levels
- **Kafka Infrastructure**: Handles message queuing, partitioning, and reliable delivery

### Features

- ✅ **Structured Logging**: JSON-formatted log messages with timestamps, application names, and log levels
- ✅ **Real-time Processing**: Immediate log consumption and display
- ✅ **Partitioned Topics**: Uses application names as partition keys for ordered processing
- ✅ **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL with color-coded display
- ✅ **Fault Tolerant**: Handles connection retries and graceful shutdowns
- ✅ **Scalable Design**: Multiple producers and consumers can run concurrently

## 🏗️ Architecture

```
[Applications] → [Log Producers] → [Kafka Topic: raw-logs] → [Log Consumers] → [Display/Processing]
     │                │                    │                      │               │
  UserService    ─────┼────────────────────┼──────────────────────┼───────────── Terminal 1
  AuthService    ─────┼────────────────────┼──────────────────────┼───────────── Terminal 2
  PaymentService ─────┼────────────────────┼──────────────────────┼───────────── Terminal 3
  DatabaseService─────┘                    │                      └───────────── ...more consumers
                                           │
                                    [Partitions 0,1,2]
                                  (ordered by app name)
```

## 📋 Prerequisites

- **Docker Desktop** (running)
- **Go 1.19+** ([Download here](https://golang.org/dl/))
- **Windows PowerShell** or **Command Prompt**

## 🚀 Quick Start

### Step 1: Clone and Setup Project

```powershell
# Create project directory
mkdir kafka-logging-system
cd kafka-logging-system

# Initialize Go module
go mod init kafka-logging-system

# Install dependencies
go get github.com/IBM/sarama

# Create project structure
mkdir -p cmd\producer cmd\consumer internal\models
```

### Step 2: Start Kafka Infrastructure

Create `docker-compose.yml`:

```yaml
version: "3.8"
services:
  kafka:
    image: bitnami/kafka:latest
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      # KRaft settings
      KAFKA_CFG_NODE_ID: 0
      KAFKA_CFG_PROCESS_ROLES: controller,broker
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 0@kafka:9093
      # Listeners
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      # Additional settings
      KAFKA_CFG_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_CFG_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_CFG_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: true
      # Disable SASL for simplicity
      ALLOW_PLAINTEXT_LISTENER: yes
    volumes:
      - kafka_data:/bitnami/kafka

volumes:
  kafka_data:
    driver: local
```

**Start Kafka:**

```powershell
# Start Kafka container
docker-compose up -d

# Wait for Kafka to be ready (important!)
Start-Sleep 30

# Verify Kafka is running
docker-compose ps
```

### Step 3: Create Kafka Topic

```powershell
# Create the raw-logs topic with 3 partitions
docker exec kafka kafka-topics.sh --create --topic raw-logs --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1

# Verify topic creation
docker exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092
```

You should see `raw-logs` in the output.

### Step 4: Build the Applications

```powershell
# Create bin directory
New-Item -ItemType Directory -Path "bin" -Force

# Build producer
go build -o "bin\producer.exe" "cmd\producer\main.go"

# Build consumer
go build -o "bin\consumer.exe" "cmd\consumer\main.go"
```

### Step 5: Run the System

**Terminal 1 - Start Consumer:**

```powershell
.\bin\consumer.exe
```

You should see:

```
2025/09/22 01:45:00 Consumer group log-consumer-group started, consuming topics: [raw-logs]
Ctrl-C to stop...
```

**Terminal 2 - Start Producer:**

```powershell
.\bin\producer.exe
```

You should see logs being generated and consumed in real-time:

```
[01:45:15] [userService] [INFO] User logged in successfully (p:2, o:5)
[01:45:18] [DatabaseService] [ERROR] Database Connection Failed (p:0, o:3)
[01:45:21] [AuthService] [WARN] High Memory usage (p:1, o:7)
```

## 🎮 Advanced Usage

### Running Multiple Producers

```powershell
# Terminal 3
.\bin\producer.exe

# Terminal 4
.\bin\producer.exe

# Each will randomly select from: UserService, DatabaseService, AuthService, PaymentService
```

### Running Multiple Consumers

```powershell
# Terminal 5 - Consumer with different group (will get all messages)
go run cmd\consumer\main.go different-consumer-group

# Terminal 6 - Consumer in same group (will share load)
go run cmd\consumer\main.go log-consumer-group
```

### Testing with Kafka Console Tools

```powershell
# Send test message
docker exec -it kafka kafka-console-producer.sh --topic raw-logs --bootstrap-server localhost:9092

# Read messages directly
docker exec -it kafka kafka-console-consumer.sh --topic raw-logs --bootstrap-server localhost:9092 --from-beginning
```

## 📊 Sample Output

```
[15:23:45] [UserService] [INFO] User logged in successfully (p:2, o:42)
[15:23:47] [DatabaseService] [ERROR] Database Connection Failed (p:0, o:23)
[15:23:49] [AuthService] [WARN] Password failed for user (p:1, o:15)
[15:23:52] [PaymentService] [INFO] Request completed (p:2, o:43)
[15:23:54] [UserService] [ERROR] Invalid user credentials (p:2, o:44)
```

**Color Coding:**

- 🔵 **DEBUG** - Cyan
- 🟢 **INFO** - Green
- 🟡 **WARN** - Yellow
- 🔴 **ERROR** - Red
- 🟣 **FATAL** - Magenta

## 🔧 Troubleshooting

### Kafka Won't Start

```powershell
# Check Docker is running
docker info

# Check container logs
docker logs kafka

# Restart Docker Desktop if needed
```

### Connection Refused Error

```powershell
# Ensure Kafka is listening on port 9092
netstat -ano | findstr :9092

# Wait longer for Kafka startup (can take 30-60 seconds)
Start-Sleep 45
```

### Topic Doesn't Exist Error

```powershell
# Manually create topic again
docker exec kafka kafka-topics.sh --create --topic raw-logs --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1

# Check topic exists
docker exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092
```

### Build Errors

```powershell
# Ensure Go is installed
go version

# Verify module dependencies
go mod tidy

# Check import paths match your project structure
```

## 🧹 Cleanup

**Stop Applications:**

- Press `Ctrl+C` in all terminal windows running producers/consumers

**Stop Kafka:**

```powershell
docker-compose down

# Optional: Remove volumes
docker-compose down -v
```

**Clean Build Files:**

```powershell
Remove-Item -Recurse -Force bin
```

## 🎓 Learning Outcomes

This project demonstrates:

1. **Kafka Fundamentals**: Topics, partitions, producers, consumers, consumer groups
2. **Go Concurrency**: Goroutines, channels, graceful shutdowns
3. **Message Serialization**: JSON encoding/decoding
4. **Distributed Systems**: Partitioning strategies, load balancing
5. **Real-time Processing**: Stream processing patterns
6. **Error Handling**: Connection retries, malformed message handling

## 🚀 Next Steps

- **Step 2**: Add log routing (ERROR logs → error-logs topic)
- **Step 3**: Implement specialized consumers (metrics, alerts, search indexing)
- **Step 4**: Add dead letter queues for failed messages
- **Step 5**: Build a web dashboard for real-time monitoring

## 📝 Project Structure

```
kafka-logging-system/
├── cmd/
│   ├── producer/
│   │   └── main.go          # Log producer application
│   └── consumer/
│       └── main.go          # Log consumer application
├── internal/
│   └── models/
│       └── log.go           # Log data structures
├── bin/                     # Built executables
├── docker-compose.yml         # Kafka infrastructure (Bitnami)
└── README.md               # This file
```

---

For questions or issues, check the troubleshooting section or review the Kafka logs with `docker logs kafka`.
