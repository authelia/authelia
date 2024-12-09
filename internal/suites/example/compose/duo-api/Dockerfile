FROM node:23-alpine@sha256:3e58d4a77778684d17ab30ba8192515bd8078083df29a2d90fe7481f78f30093

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
