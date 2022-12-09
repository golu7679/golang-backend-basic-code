# FROM golang:latest
# RUN mkdir /build
# WORKDIR /build
# RUN export GO111MODULE=on
# RUN go get https://github.com/golu7679/golang-backend-basic-code.git
# RUN cd /build && git clone https://github.com/golu7679/golang-backend-basic-code.git
# RUN cd /build/golang-backend-basic-code/without-framework && go build
# EXPOSE 8080
# ENTRYPOINT ["/build/golang-backend-basic-code/without-framework/main"]

# FROM golang:1.12.0-alpine3.9
# RUN mkdir /app
# ADD . /app
# WORKDIR /app
# RUN go build -o main .
# cmd ["/app/main"]


FROM golang:1.18.1-buster
WORKDIR /app
COPY . .
# copy go.mod ./
RUN go mod download
RUN go build -o /godocker
EXPOSE 8080
cmd ["/godocker"]