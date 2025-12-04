#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Building Authelia Locally${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

echo -e "${GREEN}Step 1/3: Building frontend...${NC}"
cd web
pnpm install --frozen-lockfile --ignore-scripts
NODE_ENV=production pnpm build --mode production
cd ..

echo ""
echo -e "${GREEN}Step 2/3: Building Go binary...${NC}"
cp -r api internal/server/public_html/api
GOEXPERIMENT="nosynchashtriemap" \
CGO_ENABLED=1 \
GOMEMLIMIT=1GiB \
go build -p 1 \
    -tags dev \
    -ldflags "-s -w" \
    -trimpath \
    -o authelia \
    ./cmd/authelia

echo ""
echo -e "${GREEN}Step 3/3: Building minimal Docker image...${NC}"
cat > Dockerfile.local <<'EOF'
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY authelia LICENSE entrypoint.sh healthcheck.sh .healthcheck.env ./
RUN chmod +x /app/authelia /app/entrypoint.sh /app/healthcheck.sh && chmod 666 /app/.healthcheck.env

ENV PATH="/app:${PATH}" \
    PUID=0 \
    PGID=0 \
    X_AUTHELIA_CONFIG="/config/configuration.yml"

EXPOSE 9091
ENTRYPOINT ["/app/entrypoint.sh"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=1m CMD /app/healthcheck.sh
EOF

docker build -f Dockerfile.local -t ghcr.io/deep-jiwan/mypackages/authelia/customtheme:autobuild .
docker tag ghcr.io/deep-jiwan/mypackages/authelia/customtheme:autobuild ghcr.io/deep-jiwan/mypackages/authelia/customtheme:latest

echo ""
echo -e "${GREEN}✓ Build completed!${NC}"
echo "Image: ghcr.io/deep-jiwan/mypackages/authelia/customtheme:autobuild"
