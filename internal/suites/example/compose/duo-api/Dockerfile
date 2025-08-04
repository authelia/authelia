FROM node:24-alpine@sha256:6a61dbdf1dd6adbfcf0c09237f8f0f7a101b479b0491f2e49b114e55f027ba16

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
