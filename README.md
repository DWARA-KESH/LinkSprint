# LinkSprint: Distributed URL Shortener & Analytics

## Overview

LinkSprint is a high-performance, distributed URL shortening service built with a modern Go-based tech stack. It allows users to create concise short links for long URLs, tracks click analytics, and is designed with scalability and fault tolerance in mind, making it an excellent demonstration of distributed system concepts.

## Live Demo

You can try out the live version of LinkSprint:
* **Frontend:** [LinkSprint](https://linksprint-frontend.onrender.com)
* **Backend API:** [Backend](https://linksprint.onrender.com)

## Features

* **URL Shortening:** Generate unique short codes for long URLs.
* **Custom Short Codes:** Users can specify a personalized short code (e.g., `linksprint.com/my-promo`).
* **Redirection Engine:** Fast and reliable redirection from short URL to original long URL.
* **Click Tracking:** Basic analytics to count total clicks per short URL.
* **Distributed Architecture:** Leverages CockroachDB for distributed, fault-tolerant data storage and Redis for high-speed caching.
* **Containerized Deployment:** Uses Docker for consistent development and deployment environments.

## Tech Stack

* **Backend:** [Go](https://golang.org/) with [Fiber](https://gofiber.io/) web framework  
* **Database:** [CockroachDB](https://www.cockroachlabs.com/) (Distributed SQL, PostgreSQL-compatible)  
* **Caching:** [Redis](https://redis.io/) (In-memory data store)  
* **Containerization:** [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)  
* **Frontend:** Plain HTML, CSS, JavaScript  
* **Deployment:** [Render.com](https://render.com/) (for free-tier hosting)  

## Architecture

LinkSprint is structured as a simple microservice, with the Go application serving as the API gateway and business logic handler.

```text
+----------------+       +-------------------+
|    Frontend    |       |   LinkSprint API  |
| (index.html)   |------>| (Go + Fiber)      |
+----------------+       |                   |
                         |  - Shorten URL    |
                         |  - Redirect URL   |
                         |  - Increment Clicks|
                         +--------+----------+
                                  |
               +------------------+------------------+
               |                                     |
               V                                     V
        +------------+                        +-------------+
        |   Redis    |                        | CockroachDB |
        | (Cache)    |<---------------------->| (Persistent |
        +------------+                        |  Storage)   |
                                              +-------------+
```

### Components

- **Go API:** Handles incoming requests for URL shortening and redirection. It interacts with Redis for fast lookups and CockroachDB for persistent storage.
- **Redis:** Serves as a high-speed cache for shortCode -> longURL mappings to minimize database load on frequent redirects.
- **CockroachDB:** Provides a strongly consistent, fault-tolerant, and horizontally scalable SQL database for storing all short URL mappings and click counts.

---

## 🧪 Local Setup & Development

### Prerequisites

- [Git](https://git-scm.com/)
- [Go](https://golang.org/dl/) (v1.24.5 or later)
- [Docker Desktop](https://www.docker.com/products/docker-desktop)

### 1. Clone the Repository

```bash
git clone https://github.com/DWARA-KESH/LinkSprint.git
cd LinkSprint
```

### 2. Set up Local Databases (CockroachDB & Redis)

```bash
# Make sure Docker is running
docker-compose up -d
```

- CockroachDB will run on `localhost:26257`
- Redis will run on `localhost:6379`

### 3. Initialize CockroachDB Schema

```bash
docker exec -it linksprint-db ./cockroach sql --insecure --host=localhost
```

Then in the SQL shell:

```sql
CREATE DATABASE linksprint;
USE linksprint;
CREATE TABLE urls (
    short_code STRING PRIMARY KEY,
    original_url STRING NOT NULL,
    click_count INT DEFAULT 0,
    custom_slug STRING UNIQUE
);

```

### 4. Install Go Dependencies

```bash
go mod tidy
```

### 5. Run the Backend API

```bash
go run cmd/api/main.go
```

API now runs on: [http://localhost:3000](http://localhost:3000)

### 6. Use the Frontend

- Open `index.html` in a browser
- Ensure it points to your local backend in JavaScript:
```javascript
const response = await fetch('http://localhost:3000/shorten', { ... });
```

---

## 🚀 Deployment (Render.com)

### What Gets Deployed:

- **Dockerized Go App:** From your GitHub repo → Render Web Service
- **CockroachDB Serverless:** Free-tier DB
- **Redis (Render Free Redis):** Used for caching
- **Frontend:** Static site hosting using Render Static Site or Netlify/Vercel

### Important Environment Variables

Set these in your Render backend service:

| Variable         | Description                              |
|------------------|------------------------------------------|
| `DATABASE_URL`   | CockroachDB connection string            |
| `REDIS_ADDR`     | Redis connection string (internal)       |
| `SERVICE_BASE_URL` | Public Render backend URL             |
| `PORT`           | Usually 8080 or 10000 on Render          |

---

## 💻 Usage

- Access frontend via browser
- Enter long URL (and optional custom code)
- Click "Shorten URL"
- Receive and test shortened link

---

## 🗂 Project Structure

```bash
LinkSprint/
├── cmd/
│   └── api/             # Main application entry point (main.go)
├── internal/
│   ├── cache/           # Redis caching logic
│   │   └── url_cache.go
│   ├── handler/         # HTTP request handlers (Fiber routes)
│   │   └── url_handler.go
│   ├── model/           # Data structures (structs for requests/responses/DB models)
│   │   └── url.go
│   └── repository/      # Database (CockroachDB) access logic
│       └── url_repository.go
├── pkg/
│   └── utils/           # Utility functions (e.g., short code generator)
│       └── code.go
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── index.html           # Simple HTML/CSS/JS frontend
```

---

## 🔮 Future Enhancements

- [ ] **Detailed Analytics:** Track IP, timestamp, User-Agent, etc.
- [ ] **Admin Dashboard:** View analytics & manage links
- [ ] **User Authentication:** Personalized link management
- [ ] **API Keys:** Secure programmatic access
- [ ] **Rate Limiting:** Prevent abuse
- [ ] **Link Expiration (TTL):** Auto-remove expired links
- [ ] **Graceful Shutdown:** Ensure clean shutdown on exit
- [ ] **Custom Domains:** Support user-owned short domains (e.g., `go.yourdomain.com`)

---

## 🙌 Credits

Built with 💙 by [Dwarakeswaran S H](https://github.com/DWARA-KESH)
