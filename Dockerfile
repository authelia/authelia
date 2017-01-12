FROM node

WORKDIR /usr/src

COPY package.json /usr/src/package.json
RUN npm install

COPY src /usr/src

ENV PORT=80
EXPOSE 80

CMD ["node", "index.js"]
