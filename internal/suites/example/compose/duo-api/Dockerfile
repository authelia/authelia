FROM node:22-alpine@sha256:8f6be98e48e57c0ebad20c67b16377a96491ef6d10a8a77b6def82ced1c1c287

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
