FROM node:23-alpine@sha256:3b34bf98664bfd5394e138b4d7956822aeffe8abed943ed4d2631eb216fef9e8

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
