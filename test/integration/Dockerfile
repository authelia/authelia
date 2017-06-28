FROM node:7-alpine

WORKDIR /usr/src

CMD ["./node_modules/.bin/mocha", "--compilers", "ts:ts-node/register", "--recursive", "test/integration"]
