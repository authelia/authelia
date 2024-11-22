FROM node:23-alpine@sha256:ea9ff8f16c31cb0260761281234b6b94246a37a9e56e6e3eb127db2b906a23b1

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
