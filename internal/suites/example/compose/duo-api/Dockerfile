FROM node:25-alpine@sha256:4e0711d09436971f759692f1437b2634bdd989cacf833c9e620c75aa04fd9f7a

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
