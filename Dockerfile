FROM node:20-alpine AS client-build
WORKDIR /app/client
COPY client/package*.json ./
RUN npm ci
COPY client/ .
RUN npm run build

FROM golang:1.22-alpine AS server-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=client-build /app/client/dist ./client/dist
RUN go build -o server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=server-build /app/server ./server
EXPOSE 8080
CMD ["./server"]
