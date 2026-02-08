FROM node:25-alpine@sha256:c8d96e95e88f08f814af06415db9cfd5ab4ebcdf40721327ff2172ff25cfb997

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
