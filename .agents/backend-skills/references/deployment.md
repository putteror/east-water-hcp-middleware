# CENT ACCESS - Deployment & Build Reference

This document outlines the standard build and deployment workflows for CENT ACCESS services.

---

## 1. Build & Push Workflow (Dev Machine)

All backend microservices and the frontend can be cross-compiled and pushed to Docker Hub using the root [build-and-push.sh](file:///Users/putter/Putteror/Developer/Platform/CentAccess/App/build-and-push.sh) script.

### Single Service Build
To build and push an individual service quickly:

```bash
./build-and-push.sh <service_name> [version]
```

**Available service names:**
- `core` -> `putteror/centaccess-app-service-core`
- `system` -> `putteror/centaccess-app-service-system`
- `visitor` -> `putteror/centaccess-app-service-visitor`
- `register` -> `putteror/centaccess-app-service-register`
- `attendance` -> `putteror/centaccess-app-service-attendance`
- `automation` -> `putteror/centaccess-app-service-automation`
- `frontend` -> `putteror/centaccess-app-frontend`

**Examples:**
```bash
# Build & push core service only (defaults to :latest)
./build-and-push.sh core

# Build & push frontend only
./build-and-push.sh frontend

# Build & push with custom tag version
./build-and-push.sh core v1.2.0
```

### Interactive Menu / Build All
```bash
# Interactive menu
./build-and-push.sh

# Build all services sequentially
./build-and-push.sh all
```

---

## 2. Production Deployment Workflow (Production Server)

On the production server, manage deployments using `docker-compose.deploy.yml`.

### Updating a Single Service (Zero Downtime for Other Services)
To update only a specific container without restarting database or other microservices:

```bash
# Pull latest image for target service and recreate its container
docker compose -f docker-compose.deploy.yml pull <service_name> && \
docker compose -f docker-compose.deploy.yml up -d --no-deps --force-recreate <service_name>
```

**Examples:**
```bash
# Update core service only
docker compose -f docker-compose.deploy.yml pull core && \
docker compose -f docker-compose.deploy.yml up -d --no-deps --force-recreate core

# Update frontend only
docker compose -f docker-compose.deploy.yml pull frontend && \
docker compose -f docker-compose.deploy.yml up -d --no-deps --force-recreate frontend
```

### Updating All Services
```bash
docker compose -f docker-compose.deploy.yml pull && \
docker compose -f docker-compose.deploy.yml up -d --force-recreate
```

### Inspecting Logs & Health Check
```bash
# Tail logs for a specific service
docker compose -f docker-compose.deploy.yml logs -f <service_name>

# Check running status
docker compose -f docker-compose.deploy.yml ps
```
