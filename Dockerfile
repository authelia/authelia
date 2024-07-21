# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.20.1@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Set environment variables
ENV PATH="/app:${PATH}" \
    PUID=0 \
    PGID=0 \
    X_AUTHELIA_CONFIG="/config/configuration.yml"

RUN \
	apk --no-cache add ca-certificates su-exec tzdata wget

COPY LICENSE .healthcheck.env entrypoint.sh healthcheck.sh ./

RUN \
	chmod 0666 /app/.healthcheck.env

COPY authelia-${TARGETOS}-${TARGETARCH}-musl ./authelia

EXPOSE 9091

VOLUME /config

ENTRYPOINT ["/app/entrypoint.sh"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=1m CMD /app/healthcheck.sh
