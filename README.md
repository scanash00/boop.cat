# boop.cat codebase

This is boop.cat's codebase, and it was released under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Features

- **Instant Deployment**: Connect any public or private Git repository.
- **Auto-Deploy**: Automatically triggers a new build on every `push` to your main branch. (for GitHub only, use API for other platforms)
- **Edge Delivery**: Powered by Cloudflare Workers for global caching and low latency.
- **Managed SSL**: Automatic HTTPS for every site and custom domain.
- **Environment Variables**: Full support for build-time environment variables.
- **Clean API**: Manage your sites and deployments programmatically.

## Tech Stack

- **Backend**: Go (chi router, SQLite)
- **Frontend**: React (Vite, Lucide Icons)
- **Storage**: Backblaze B2 (Object storage)
- **Delivery**: Cloudflare Workers (Edge Computing) & KV (Metadata)
- **Database**: SQLite (Local file-based database)

## Getting Started

### 1. Prerequisites

- [Go](https://go.dev/) 1.23+
- [Node.js](https://nodejs.org/) 22+ (with `pnpm` or `bun`)
- [Cloudflare Account](https://dash.cloudflare.com/) (Token, Account ID, Zone ID)
- [Backblaze B2 Account](https://www.backblaze.com/b2/cloud-storage.html) (Key ID, App Key, Bucket)

### 2. Installation

```bash
# Clone the repository
git clone https://tangled.org/scanash.com/boombox
cd boombox

# Install frontend dependencies
npm install
```

### 3. Configuration

Copy the example environment file and fill in your credentials:

```bash
cp .env.example .env
```

Key variables to configure:

- `SESSION_SECRET`: Random string for sessions.
- `FSD_DATA_DIR`: Path where the SQLite database will be stored.
- `CF_*`: Your Cloudflare API credentials.
- `B2_*`: Your Backblaze B2 storage credentials.

### 4. Running Locally

**Start the Go Backend:**

```bash
cd backend-go
go run main.go
```

The backend serves the frontend from `client/dist`. For development, you can run the Vite dev server separately:

**Start Vite Dev Server:**

```bash
cd client
npm run dev
```

## Docker Deployment

The project includes a multi-stage `Dockerfile` that builds both the React frontend and Go backend into a single production-ready image.

```bash
docker build -t boop-cat .
docker run -p 8788:8788 --env-file .env boop-cat
```

## API Documentation

The platform provides a REST API for managing sites. See the **API Documentation** page within the dashboard for details and examples.

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
