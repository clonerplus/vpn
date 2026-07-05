# vpn-methods

VPN infrastructure for outi.ir, deployable on Kubernetes with Kustomize.

## Architecture

```
Client ──► VLESS/TCP ──► vless-server ──► Internet
         ──► VLESS/Reality ──► vless-reality ──► Internet
         ──► SS/WS ──► ss-server ──► Internet
         ──► WireGuard ──► vpn-gateway ──► Internet
```

## Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| vless-server | 443 | TCP | VLESS proxy with UUID auth |
| vless-reality | 443 | TCP | VLESS with Reality camouflage |
| ss-server | 80, 9050 | WS | Shadowsocks over WebSocket |
| vpn-gateway | 51820 | UDP | WireGuard VPN gateway |
| vpn-api | 8080 | HTTP | User management API |

## Directory Structure

```
vpn-methods/
├── vless-server/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── ingress.yml
│   │   ├── certificate.yml
│   │   └── config/
│   │       ├── kustomization.yml
│   │       └── config.json
│   └── overlays/stage/
├── vless-reality/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── ingress.yml
│   │   └── config/
│   │       ├── kustomization.yml
│   │       └── config.json
│   └── overlays/stage/
├── ss-server/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── ingress.yml
│   │   └── config/
│   │       ├── kustomization.yml
│   │       └── config.json
│   └── overlays/stage/
├── vpn-gateway/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   └── configmap.yml
│   └── overlays/stage/
├── vpn-api/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── ingress.yml
│   │   └── config/
│   │       └── config.json
│   └── overlays/stage/
├── cmd/
│   └── vpn-server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── vless/
│   │   └── server.go
│   └── wireguard/
├── Dockerfile
├── go.mod
├── go.sum
└── .github/workflows/
    ├── deploy.yml
    ├── rollout.yml
    └── status.yml
```

## Deployment

```bash
kubectl apply -k vless-server/overlays/stage
kubectl apply -k vless-reality/overlays/stage
kubectl apply -k ss-server/overlays/stage
kubectl apply -k vpn-gateway/overlays/stage
kubectl apply -k vpn-api/overlays/stage
```

## Configuration

### VLESS Server

Edit `vless-server/base/config/config.json`:

```json
{
    "vless": {
        "enabled": true,
        "port": 443,
        "uuid": "your-uuid-here",
        "flow": "",
        "security": "none"
    },
    "users": [
        {
            "uuid": "user-uuid",
            "name": "username"
        }
    ]
}
```

### VLESS Reality

Edit `vless-reality/base/config/config.json` and update:
- `privateKey` with your Reality private key
- `serverNames` with your domain
- `shortIds` with your short IDs

### Shadowsocks

Edit `ss-server/base/config/config.json`:
- `method`: cipher (default `chacha20-ietf-poly1305`)
- `password`: shared secret
- `ports`: 80 (main), 9050 (tor)

### WireGuard

Generate keys:
```bash
wg genkey | tee privatekey | wg pubkey > publickey
```

Update `vpn-gateway/base/configmap.yml` with your keys.

## Client Connection

### VLESS
```
Server: vless.outi.ir
Port: 443
UUID: <your-uuid>
Transport: TCP
```

### VLESS Reality
```
Server: <your-reality-domain>
Port: 443
UUID: <your-uuid>
Flow: xtls-rprx-vision
Security: Reality
SNI: www.cloudflare.com
```

### Shadowsocks
```
Server: ss.outi.ir
Port: 80
Password: <your-password>
Method: chacha20-ietf-poly1305
Plugin: v2ray-plugin (WebSocket)
```

### WireGuard
```
Endpoint: vpn.outi.ir:51820
Public Key: <server-public-key>
AllowedIPs: 0.0.0.0/0
```

## CI/CD

- **deploy.yml**: Auto-deploys on push to master
- **rollout.yml**: Manual restart of specific service
- **status.yml**: Manual check of all resources
