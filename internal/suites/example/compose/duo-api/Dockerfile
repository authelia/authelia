FROM node:23-alpine@sha256:5672de224fa400409f65e9b890254c1caf93a3d84665737960626f9419ae4f77

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
