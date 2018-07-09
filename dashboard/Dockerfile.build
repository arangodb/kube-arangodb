FROM node:10.6-alpine

RUN mkdir -p /usr/code
ADD package-lock.json /usr/code/
ADD package.json /usr/code/

RUN cd /usr/code/ && npm install

VOLUME /usr/code/build
VOLUME /usr/code/public
VOLUME /usr/code/src

WORKDIR /usr/code/
CMD ["npm", "run-script", "build"]
