# nginxctl

A CLI tool for setting up nginx as a reverse proxy and configuring SSL via Let's Encrypt. Runs as a guided wizard — no config file editing required.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/hardope/nginxctl/master/install.sh | sudo bash
```

Or download a binary directly from [Releases](https://github.com/hardope/nginxctl/releases):

```bash
sudo curl -fsSL https://github.com/hardope/nginxctl/releases/latest/download/nginxctl-linux-amd64 \
  -o /usr/local/bin/nginxctl && sudo chmod +x /usr/local/bin/nginxctl
```

> Use `nginxctl-linux-arm64` for ARM-based servers (AWS Graviton, etc.)

## Usage

All commands require root:

### Full setup wizard

Installs nginx if needed, writes a reverse-proxy config, and optionally sets up SSL.

```bash
sudo nginxctl setup
```

Prompts for:
- Upstream URL (e.g. `http://localhost:3000`)
- Domain(s) (e.g. `example.com www.example.com`)
- Max upload size (default: `100M`)

Shows the generated config for review before writing anything. Creates a timestamped backup of any existing config before overwriting.

### Add SSL to an existing config

```bash
sudo nginxctl ssl
```

Prompts for the path to your existing nginx config and your domain(s), verifies DNS A records point to this server, then runs certbot.

## What gets configured

The generated nginx config includes:

- WebSocket support (`Upgrade`, `Connection` headers)
- Standard forwarding headers (`X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`)
- Configurable `client_max_body_size`
- Long-lived connection timeout (`proxy_read_timeout 86400s`)
- `proxy_buffering off`
- HTTP → HTTPS redirect (after SSL is set up)

## DNS verification

Before running certbot, the tool resolves each domain's A record and compares it to the server's public IP. If any domain doesn't match, you'll see which ones failed and can choose to abort or proceed anyway.

## Supported systems

| Distro | Package manager |
|---|---|
| Ubuntu / Debian | `apt` |
| CentOS / RHEL / Rocky / AlmaLinux / Fedora | `dnf` |

## Building from source

Requires Go 1.21+.

```bash
git clone https://github.com/hardope/nginxctl.git
cd nginxctl
go build -o nginxctl .
sudo mv nginxctl /usr/local/bin/
```

## Releasing

Push a `v*` tag to trigger a GitHub Actions build that cross-compiles for `linux/amd64` and `linux/arm64` and publishes the binaries to GitHub Releases.

```bash
git tag v1.0.0
git push origin v1.0.0
```
