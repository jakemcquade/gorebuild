# GoRebuild

![](https://img.shields.io/github/repo-size/jakemcquade/gorebuild)

A small self-hosted webhook server that redeploys your projects on every push.
Point a GitHub webhook at it, and on a matching push it resets the repo to the pushed commit and rebuilds it with Docker Compose.

## How it works

1. GitHub sends a `push` webhook to `POST /webhook`.
2. The request signature is verified with HMAC-SHA256 against your secret.
3. The pushed repository is matched against a project in `config.yaml`
   (and, optionally, filtered to a single branch).
4. In that project's directory, GoRebuild runs:
   ```sh
   git fetch origin
   git reset --hard origin/<branch>   # deterministic - no merge, no local drift
   docker compose up -d --build
   ```

Rebuilds are **per-repo, latest-wins**: if a new push arrives while a rebuild is
still running, the in-flight rebuild is cancelled and replaced. Each rebuild is
bounded by a 10-minute timeout so a stuck command can't hang forever.

## Requirements

- Go 1.26+ (to build)
- `git` and `docker compose` available on the host
- Each project already cloned, with an `origin` remote and a `docker-compose.yml`

## Configuration

### `config.yaml`

```yaml
server:
  port: 8765          # optional, defaults to 8080

projects:
  disping:            # must match the GitHub repository name
    path: ~/bots/disping   # working directory the commands run in
    branch: main           # optional; only rebuild on pushes to this branch
```

- The project key (`disping`) must equal the repo name GitHub sends in the payload.
- `~` in `path` is expanded to the current user's home directory.
- Omit `branch` to rebuild on a push to any branch.

### Environment

| Variable          | Required | Default        | Description                                  |
| ----------------- | -------- | -------------- | -------------------------------------------- |
| `WEBHOOK_SECRET`  | yes      | -              | Shared secret used to verify webhook signatures. |
| `CONFIG_PATH`     | no       | `config.yaml`  | Path to the config file.                     |

Set `WEBHOOK_SECRET` in a `.env` file next to the binary, or in the real
environment:

```env
WEBHOOK_SECRET=your-long-random-string
```

The server refuses to start if `WEBHOOK_SECRET` is unset.

## Running

```sh
go build .
./gorebuild
```

The server logs the loaded projects and listens on the configured port.

## Endpoints

| Method | Path       | Description                                          |
| ------ | ---------- | ---------------------------------------------------- |
| `GET`  | `/`        | Health check - returns `Server up and running.`      |
| `POST` | `/webhook` | GitHub webhook receiver (signed requests only).      |

## GitHub webhook setup

In your repository: **Settings → Webhooks → Add webhook**

- **Payload URL:** `https://your-host/webhook`
- **Content type:** `application/json`
- **Secret:** the same value as `WEBHOOK_SECRET`
- **Events:** *Just the push event*

A malformed or unsigned request is rejected with `401`; unknown repos and
filtered-out branches are acknowledged with `200` and ignored.
