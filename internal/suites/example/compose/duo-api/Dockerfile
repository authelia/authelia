FROM node:24-alpine@sha256:6ff78d6d45f2614fe0da54756b44a7c529a15ebcaf9832fab8df036b1d466e73

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
