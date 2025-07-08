# VoidDB

A lightweight, high-performance in-memory database written in Go with persistent transaction logging, TTL support, and priority-based storage.

## Features

- **In-Memory Storage**: Ultra-fast key-value storage with thread-safe operations
- **Persistent Transaction Log**: Automatic recovery from previous state on restart
- **TTL Support**: Automatic expiration of keys with configurable time-to-live
- **Priority System**: Store items with priority levels for advanced use cases
- **HTTP REST API**: Simple HTTP endpoints for database operations
- **Concurrent Safe**: Built-in mutex protection for thread-safe operations
- **Automatic Cleanup**: Background janitor process for expired key removal

## Quick Start

### Prerequisites

- Go 1.24.1 or higher

### Installation

```bash
git clone https://github.com/ivoidable/void-db.git
cd void-db
go mod tidy
```

### Running the Server

```bash
go run main.go
```

The server will start on port 6969 and load any existing data from the transaction log.

## API Endpoints

### GET /get
Retrieve a value by key.

**Parameters:**
- `key` (query parameter): The key to retrieve

**Example:**
```bash
curl "http://localhost:6969/get?key=mykey"
```

### POST /set
Store a key-value pair with optional TTL and priority.

**Request Body:**
```json
{
  "key": "mykey",
  "value": "myvalue",
  "ttl": 300,
  "priority": 1
}
```

**Fields:**
- `key` (string): The key to store
- `value` (any): The value to store
- `ttl` (int64, optional): Time-to-live in seconds
- `priority` (int, optional): Priority level for the item

**Example:**
```bash
curl -X POST http://localhost:6969/set \
  -H "Content-Type: application/json" \
  -d '{"key":"mykey","value":"myvalue","ttl":300,"priority":1}'
```

### DELETE /delete
Remove a key-value pair.

**Parameters:**
- `key` (query parameter): The key to delete

**Example:**
```bash
curl -X DELETE "http://localhost:6969/delete?key=mykey"
```

## Architecture

VoidDB consists of several key components:

- **In-Memory Store**: Thread-safe map-based storage with RWMutex protection
- **Transaction Logger**: Persistent logging system that records all operations
- **TTL Manager**: Background janitor process that removes expired keys
- **HTTP Server**: REST API layer for database operations

### Data Structures

```go
type Item struct {
    Value      interface{}
    Expiration int64
    Priority   int
}

type Transaction struct {
    Action string
    Key    string
    Item   Item
}
```

## Persistence

VoidDB maintains durability through a transaction log (`transaction.log`) that records all SET and DELETE operations. On startup, the database automatically replays the transaction log to restore the previous state.

## Performance Features

- **Concurrent Read/Write**: Optimized with RWMutex for concurrent read operations
- **Background Cleanup**: Efficient janitor process runs every second to remove expired keys
- **Memory Efficient**: Direct in-memory storage without unnecessary overhead
- **Fast Recovery**: Quick startup times with transaction log replay

## Use Cases

- **Caching Layer**: High-performance cache with TTL support
- **Session Storage**: Store user sessions with automatic expiration
- **Rate Limiting**: Use priority system for rate limiting implementations
- **Temporary Data**: Store temporary data with automatic cleanup
- **Configuration Cache**: Cache configuration data with priority levels

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Created by [ivoid](https://github.com/ivoidable)

---

**Note**: This is a lightweight database intended for development and small-scale production use. For high-availability production environments, consider using established database solutions.
