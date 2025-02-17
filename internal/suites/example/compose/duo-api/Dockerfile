FROM node:23-alpine@sha256:0a29d8a047149079bf48dab598f287815e39867d547e8f3935cfdece0a4379cc

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
