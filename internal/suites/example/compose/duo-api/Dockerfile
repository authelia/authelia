FROM node:22-alpine@sha256:0ed6a8d9b031a649d2aae65986df03f7dc4df9e0d3addc03e3a6aa4684ca32f9

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
