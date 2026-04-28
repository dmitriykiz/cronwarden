# cronwarden

Lightweight cron job monitoring daemon with webhook alerts and a local SQLite log.

---

## Installation

```bash
go install github.com/yourname/cronwarden@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwarden.git && cd cronwarden && go build -o cronwarden .
```

---

## Usage

Start the daemon with a config file:

```bash
cronwarden --config /etc/cronwarden/config.yaml
```

Example `config.yaml`:

```yaml
database: /var/lib/cronwarden/jobs.db
webhook: https://hooks.example.com/alerts

jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert_after: 5m
  - name: sync-reports
    schedule: "*/15 * * * *"
    timeout: 5m
```

Ping a job on start and finish from your cron script:

```bash
# In your crontab
* * * * * curl -s http://localhost:8080/ping/sync-reports/start && /usr/local/bin/sync.sh && curl -s http://localhost:8080/ping/sync-reports/finish
```

View recent job history:

```bash
cronwarden logs --job daily-backup --limit 20
```

---

## How It Works

- Monitors registered jobs for missed runs, timeouts, and failures
- Stores all run history in a local SQLite database
- Fires webhook POST requests when a job is late, timed out, or fails

---

## License

MIT © 2024 yourname