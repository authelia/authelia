# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.20.3@sha256:a8f120106f5549715aa966fd7cefaf3b7045f6414fed428684de62fec8c2ca4b

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
