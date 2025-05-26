FROM node:24-alpine@sha256:dfea0736e82fef246aba86b2082a5e86c4825470302692b841d097dd61253b79

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
