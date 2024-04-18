FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o followers

FROM alpine
COPY --from=build_container /app/followers /usr/bin
EXPOSE 8082
ENTRYPOINT ["followers"]
