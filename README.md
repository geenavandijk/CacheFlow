# CacheFlow

**Invest. Grow. Retire.**

CacheFlow is a web-based investment platform that builds personalized stock portfolios based on user preferences, risk tolerance, and real-world financial data.

---

# Tech Stack

## Backend
- Go (Golang)

## Database
- MongoDB (NoSQL document store)

## Frontend
- React + Vite

## Cloud / Secrets Management
- Google Cloud Platform (GCP)
- Google Cloud CLI (`gcloud`)
- GCP Secret Manager

---

# Prerequisites

Before running the project, install:

### 1. Go
https://go.dev/dl/

Verify installation:
```
go version
```

### 2. Node.js + npm
https://nodejs.org/

Verify installation:
```
node -v
npm -v
```

### 3. Google Cloud CLI
https://cloud.google.com/sdk/docs/install

Verify installation:
```
gcloud version
```

---

# Running CacheFlow Locally (First Time Setup)

You must run BOTH the backend and frontend.

---

# Backend Setup

### 1. Open a terminal

Navigate to the backend directory:

```
cd backend-go/code.cacheflow.internal
```

### 2. Download dependencies

```
go mod download
```

### 3. Authenticate with Google Cloud

```
gcloud auth login
```

Log in with the Gmail account associated with the CacheFlow project.

### 4. Run the backend server

```
go run main.go
```

If successful, the server will start on the configured port (usually `localhost:8080` unless changed).

Leave this terminal running.

---

# Frontend Setup

Open a NEW terminal window.

### 1. Navigate to client folder

```
cd client
```

### 2. Install dependencies

```
npm install
```

If you receive dependency errors:

```
npm install --legacy-peer-deps
```

### 3. Start the development server

```
npm run dev
```

You should see a local URL such as:

```
http://localhost:5173
```

Open that URL in your browser.

---

# Running After First Setup

After dependencies are installed, future runs are simpler.

## Backend
```
cd backend-go/code.cacheflow.internal
go run main.go
```

## Frontend
```
cd client
npm run dev
```

---

# Environment Variables

The backend relies on environment variables such as:

- `PORT`
- `MONGODB_URI`
- `JWT_SECRET`
- `STOCK_API_KEY`
- `GCP_PROJECT_ID`

⚠️ Do NOT commit secrets to GitHub.

Use:
- `.env` locally
- Google Cloud Secret Manager in production

---

# How the System Works

### MongoDB
MongoDB is a document store (NOT relational like SQL).

Instead of tables and rows, Mongo uses:

- Collections
- Documents (JSON format)

Example document:

```json
{
  "first_name": "John",
  "last_name": "Apple"
}
```

---

# MVP Features

- User Signup
- User Login
- Authentication (JWT-based)
- Protected Routes
- Basic Dashboard
- MongoDB persistence

---

# Troubleshooting

### Backend won’t start
- Make sure Go is installed
- Ensure `go mod download` was run
- Ensure `gcloud auth login` was completed

### Frontend won’t start
- Run `npm install`
- Try `npm install --legacy-peer-deps`

### Port already in use
Kill the existing process or change the configured port.

---

# Branching Strategy

- `main` → Stable releases only
- `develop` → Active integration
- `feature/*` → New work
- `hotfix/*` → Urgent fixes

No direct commits to `main`.

---

# Developer Notes

- All API responses should follow consistent JSON format.
- Use middleware for logging and authentication.
- Centralize error handling.
- Never expose secrets to the frontend.
