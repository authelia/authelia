FROM node:24-alpine@sha256:d20f4872a6fbbf88b62df40944e635d70f70437217d88c039c5835b10d008158

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
