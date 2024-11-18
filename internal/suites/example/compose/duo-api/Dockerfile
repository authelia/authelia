FROM node:23-alpine@sha256:7f1b1f0e5098f6684258398d58c6c03e5f29a50b405acf9d7be9d2c83bd6c24f

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
