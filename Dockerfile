# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.20.3@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a

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
