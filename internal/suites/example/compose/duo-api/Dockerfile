FROM node:25-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4

WORKDIR /usr/app/src

COPY --link package.json pnpm-lock.yaml ./

RUN yarn global add pnpm && pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
