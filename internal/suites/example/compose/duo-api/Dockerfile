FROM node:25-alpine@sha256:08e94d88b58859f44d89e0597b416385840e69a1ec56c800c810dfb8c8d0c1bf

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
