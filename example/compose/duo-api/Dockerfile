FROM node:8.7.0-alpine

WORKDIR /usr/app/src

ADD package.json package.json
RUN npm install --production --quiet

ADD duo_api.js duo_api.js

EXPOSE 3000

CMD ["node", "duo_api.js"]