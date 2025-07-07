FROM node:24-alpine@sha256:31214211871e7c1fed53d111995ec5096f8501697f3206b9e4093202ade151ea

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
