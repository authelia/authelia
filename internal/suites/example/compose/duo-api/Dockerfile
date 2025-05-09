FROM node:24-alpine@sha256:37712740dc486f179b9540be1c6703cef3f805ea932573a007db748b71189afe

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
