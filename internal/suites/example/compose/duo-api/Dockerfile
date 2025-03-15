FROM node:23-alpine@sha256:6eae672406a2bc8ed93eab6f9f76a02eb247e06ba82b2f5032c0a4ae07e825ba

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
