FROM node:22-alpine@sha256:bec0ea49c2333c429b62e74e91f8ba1201b060110745c3a12ff957cd51b363c6

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
