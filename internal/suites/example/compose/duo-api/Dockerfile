FROM node:23-alpine@sha256:c88fbd39c466c1383f6d62e010bd6e6a57535973144fd561a1db52343ef11637

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
