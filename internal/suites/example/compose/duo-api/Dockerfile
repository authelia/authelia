FROM node:25-alpine@sha256:809972647175c30a4c7763d3e6cc064dec588972af57e540e5a6f27442bb0845

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
