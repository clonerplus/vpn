# vpn-methods

VLESS over HTTPUpgrade VPN on Kubernetes with Kustomize.

## Architecture

```
Client ──► VLESS/HTTPUpgrade/TLS ──► xray-core ──► Internet
```

## Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| vless-http (CDN) | 443 | HTTPUpgrade + TLS | VLESS proxy via cdn.outi.ir |
| vless-http (Direct) | 8443 | HTTPUpgrade + TLS | VLESS proxy via static.outi.ir |
| vless-subscription | 80 | HTTP | Subscription endpoint with basic auth |

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
└── .github/workflows/
    ├── deploy.yml
    └── rollout.yml
```

## Deployment

```bash
kubectl apply -k vless-http/overlays/stage
```

## Configuration

Edit `vless-http/base/config/config.json` to update:
- Client UUIDs
- Domain names (serverName, httpupgrade host)
- TLS certificate paths

Edit `vless-http/base/subscription-configmap.yml` to update:
- Subscription links
- Basic auth credentials
- Dashboard HTML
