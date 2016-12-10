FROM node

WORKDIR /usr/src

COPY app /usr/src

CMD ["node", "app.js"]
