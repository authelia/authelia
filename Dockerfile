# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.21.0@sha256:e323a465c03a31ad04374fc7239144d0fd4e2b92da6e3e0655580476d3a84621

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
