FROM node

WORKDIR /usr/src

COPY package.json /usr/src/package.json
RUN npm install

COPY src /usr/src

CMD ["node", "index.js"]
