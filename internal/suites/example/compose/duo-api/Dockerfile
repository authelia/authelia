FROM node:25-alpine@sha256:ad82ecad30371c43f4057aaa4800a8ed88f9446553a2d21323710c7b937177fc

WORKDIR /usr/app/src

COPY --link package.json pnpm-lock.yaml ./

RUN yarn global add pnpm && pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
