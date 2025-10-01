FROM node:24-alpine@sha256:77f3c4d1f33c17dfa4af4b0add57d86957187873e313c2c37f52831d117645c8

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
