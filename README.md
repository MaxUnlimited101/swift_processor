# Swift Codes Processor

Swift Codes Processor is a tool designed to process and store information about banks. This project is implemented in Golang.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Running the Project](#running-the-project)
- [Testing](#testing)

## Prerequisites

Before you begin, ensure you have the following installed:
- Go (version 1.18 or later)
- Git
- Postgres v17

OR:
- Git
- Docker

## Setup

1. Clone the repository:
    ```bash
    git clone https://github.com/your-username/swift_processor.git
    cd swift_processor
    ```

2. Install dependencies (if you don't want to use Docker):
    ```bash
    go mod tidy
    ```

## Running the Project
Before running the application, you need to have the csv file, with intial bank data. You have to put it as 'data.csv' in the data folder. Also, you have to specify the data in the .env file, using .env-example.
To start the application, run:
```bash
go run main.go
```

Alternatively, if you want to use Docker:
```bash
docker-compose up
```

## Testing

To run the tests, use:
```bash
go test ./...
```

This will execute all tests (unit and integration) and display the results.