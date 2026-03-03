# sykell-crawler

A small full-stack URL crawler application.

## Overview

This project consists of:

- **Backend (Go)** -- REST API + background crawler workers\
- **Frontend (React + Vite)** -- Simple UI to manage URLs\
- **MySQL** -- Persistence layer (via Docker Compose)

------------------------------------------------------------------------

## Features

- Create URLs in bulk
- Start / Stop crawling
- List URLs with pagination
- Get URL details
- Background scheduler + worker pool
- Bearer token authentication (except `/healthz`)

------------------------------------------------------------------------

## URL Lifecycle

`created → queued → running → done | failed | expired | stopped`

Supported statuses:

- created
- queued
- running
- done
- failed
- stopped
- expired

------------------------------------------------------------------------

## Prerequisites

- Go
- Node.js + npm
- Docker + Docker Compose

------------------------------------------------------------------------

## Quick Start

### 1️⃣ Start Infrastructure (MySQL)

``` bash
make infra-up
```

### 2️⃣ Run Migrations

``` bash
make migrate
```

### 3️⃣ Run Backend

``` bash
make run-backend
```

Backend runs on:

    http://localhost:8080

### 4️⃣ Run Frontend

``` bash
make run-frontend
```

------------------------------------------------------------------------

## Tests

### Backend tests

From the repository root:

```bash
cd backend
go test ./... -v

------------------------------------------------------------------------

## Configuration

### Backend Environment Variables

  Variable         Default
  ---------------- --------------
  HTTP_ADDR        0.0.0.0:8080
  API_TOKEN        dev-token
  MYSQL_HOST       localhost
  MYSQL_PORT       3306
  MYSQL_USER       crawler_user
  MYSQL_PASSWORD   supersecret
  MYSQL_DATABASE   crawler

Crawler tuning:

- `CRAWL_SCHEDULE_INTERVAL` (default: 10s)
- `CRAWL_SCHEDULE_BATCH_SIZE` (default: 10)

------------------------------------------------------------------------

## Frontend Environment

Create `.env` in `frontend/`:

``` env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TOKEN=dev-token
```

------------------------------------------------------------------------

## API Examples

### Health

``` bash
curl http://localhost:8080/healthz
```

### Create URLs

``` bash
curl -X POST http://localhost:8080/urls   -H "Authorization: Bearer dev-token"   -H "Content-Type: application/json"   -d '{"urls":["https://example.com"]}'
```

### List URLs

``` bash
curl http://localhost:8080/urls?page=1&limit=20   -H "Authorization: Bearer dev-token"
```

### Start Crawling

``` bash
curl -X POST http://localhost:8080/urls/start   -H "Authorization: Bearer dev-token"   -H "Content-Type: application/json"   -d '{"ids":[1,2]}'
```

### Stop Crawling

``` bash
curl -X POST http://localhost:8080/urls/stop   -H "Authorization: Bearer dev-token"   -H "Content-Type: application/json"   -d '{"ids":[1,2]}'
```

------------------------------------------------------------------------

## Development Commands

``` bash
make tidy
make fmt
make infra-down
```
