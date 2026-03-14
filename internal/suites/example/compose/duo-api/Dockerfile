FROM node:25-alpine@sha256:5209bcaca9836eb3448b650396213dbe9d9a34d31840c2ae1f206cb2986a8543

WORKDIR /usr/app/src

COPY --link package.json pnpm-lock.yaml ./

RUN yarn global add pnpm && pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
