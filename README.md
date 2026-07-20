# vpn-methods

VPN proxy services on Kubernetes with Kustomize.

## Architecture

```
Client ──► VLESS/HTTPUpgrade/TLS ──► xray-core ──► Internet
Client ──► Mieru/TCP ───────────────► mita ────────► Internet
```

## Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| vless-http (CDN) | 443 | HTTPUpgrade + TLS | VLESS proxy via cdn.outi.ir |
| vless-http (Direct) | 8443 | HTTPUpgrade + TLS | VLESS proxy via static.outi.ir |
| vless-subscription | 80 | HTTP | Subscription endpoint with basic auth |
| mieru | 443 | Mieru TCP | Mieru proxy via mieru.outi.ir (NodePort 30043) |

## Directory Structure

```
vpn-methods/
├── vless-http/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   ├── ingress.yml
│   │   ├── certificate.yml
│   │   ├── subscription-configmap.yml
│   │   ├── subscription-deployment.yml
│   │   ├── subscription-service.yml
│   │   └── config/
│   │       └── config.json
│   └── overlays/stage/
│       ├── kustomization.yml
│       ├── cronjob.yml
│       └── cronjob-rbac.yml
├── mieru/
│   ├── base/
│   │   ├── kustomization.yml
│   │   ├── deployment.yml
│   │   ├── service.yml
│   │   └── config/
│   │       └── server.json
│   └── overlays/stage/
│       ├── kustomization.yml
│       ├── cronjob.yml
│       └── cronjob-rbac.yml
└── .github/workflows/
    ├── deploy.yml
    └── rollout.yml
```

## Deployment

```bash
# Deploy VLESS
kubectl apply -k vless-http/overlays/stage

# Deploy Mieru
kubectl apply -k mieru/overlays/stage
```

## Configuration

### VLESS

Edit `vless-http/base/config/config.json` to update:
- Client UUIDs
- Domain names (serverName, httpupgrade host)
- TLS certificate paths

### Mieru

Edit `mieru/base/config/server.json` to update:
- Port bindings and protocol (TCP/UDP)
- User credentials (name and password)
- MTU and logging level

After deploying, update the client config in `vless-http/base/subscription-configmap.yml` (`mieru-config.json` key) with the actual server IP.

### Subscription

Edit `vless-http/base/subscription-configmap.yml` to update:
- VLESS subscription links
- Mieru client config (server IP, credentials)
- Basic auth credentials
- Dashboard HTML
