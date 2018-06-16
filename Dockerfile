FROM golang:1.10 as build

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x  /usr/local/bin/dep 

RUN mkdir -p /go/src/github.com/sunilkumarmohanty/tictactoe
WORKDIR /go/src/github.com/sunilkumarmohanty/tictactoe

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o app cmd/main.go


FROM alpine:3.7

RUN addgroup -S tictactoe && adduser -S -G tictactoe tictactoe
RUN mkdir -p /home/tictactoe
RUN chown tictactoe /home/tictactoe

WORKDIR /home/tictactoe

COPY --from=build /go/src/github.com/sunilkumarmohanty/tictactoe/repository/migrations /migrations
COPY --from=build /go/src/github.com/sunilkumarmohanty/tictactoe/app .

USER tictactoe

CMD ["./app"]