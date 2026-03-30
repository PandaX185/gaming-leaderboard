FROM node:18-alpine AS builder
WORKDIR /app
RUN npm install -g http-server
COPY . .

FROM node:18-alpine
WORKDIR /app
RUN npm install -g http-server
COPY --from=builder /app .
EXPOSE 3001
CMD ["http-server", ".", "-p", "3001"]
