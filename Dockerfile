# ===================================
# ===== Authelia official image =====
# ===================================
ARG BASE="authelia/base:latest"

FROM ${BASE}

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

ENV \
	PATH="/app:${PATH}" \
	PUID=0 \
	PGID=0 \
	X_AUTHELIA_CONFIG="/config/configuration.yml"

COPY --link authelia-${TARGETOS}-${TARGETARCH}/authelia LICENSE entrypoint.sh healthcheck.sh ./

COPY --link --chmod=666 .healthcheck.env ./

EXPOSE 9091

ENTRYPOINT ["/app/entrypoint.sh"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=1m CMD /app/healthcheck.sh
