FROM node:22-alpine@sha256:c06bea602e410a3321622c7782eb35b0afb7899d9e28300937ebf2e521902555

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
