# Manga Scraper

A full-stack application for scraping, downloading, and managing manga from WeebCentral. Features automated scraping with Playwright, asynchronous task processing via Celery/RabbitMQ, archive creation (PDF/CBZ), and a modern HTMX-powered web dashboard.

**Architecture**: Go server handles HTTP requests and frontend (Templ templates, PostgreSQL, Redis), while Python workers process scraping tasks asynchronously. Fully containerized with Docker.

## Prerequisites

- Docker and Docker Compose
- Python 3.9+
- Go 1.25+
- Node.js (for frontend assets)
- Make

## 🛠️ Installation

### Using Docker (Recommended)

1. Clone the repository:

```bash
git clone https://github.com/pratham-ak2004/manga-scraper.git
cd manga-scraper
```

2. Create docker environment file:

```bash
cp .env.docker.example .env.docker
# Edit .env with your configuration
```

3. Start all services:

```bash
make compose
# or docker compose up -d
```

This will start:

- PostgreSQL (port 5432)
- RabbitMQ (ports 5672, 15672)
- Redis (port 6379)
- Go Server (port 8080)
- Python Worker

### Manual Setup

#### Server Setup

```bash
cd server
make install
make build
```

- `Note:` The application uses PostgreSQL with SQLC for type-safe SQL queries. Schemas are located in `server/db/schema.sql`. Execute the schema queries to generate tables.

#### Worker Setup

```bash
python -m venv venv # or with conda
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
make worker
```

### Access the Application

- **Web Dashboard**: http://localhost:8080
- **RabbitMQ Management**: http://localhost:15672
- **Prometheus Metrics (RAW)**: http://localhost:8080/metrics

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 5432, 5672, 6379, 8080, and 15672 are available
2. **Database connection**: Check DATABASE_URL in your .env file
3. **Worker not processing**: Verify RabbitMQ is running and accessible
4. **Playwright errors**: Install browser dependencies: `playwright install chromium`
5. **Dashboard Error**: Check if database has the tables defined at `server/db/schema.sql`

## Disclaimer

This tool is for educational purposes only. Please respect copyright laws and the terms of service of the websites you scrape. The authors are not responsible for any misuse of this software.

## Acknowledgments

- [WeebCentral](https://weebcentral.com) - Source website
- [Playwright](https://playwright.dev/) - Browser automation
- [HTMX](https://htmx.org/) - Dynamic HTML
- [Templ](https://templ.guide/) - Go templating
