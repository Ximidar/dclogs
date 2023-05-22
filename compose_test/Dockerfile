FROM golang:latest

RUN mkdir /code
WORKDIR /code
COPY ./src .
RUN go mod tidy
# It's such a small program it's not worth compiling and extracting
ENTRYPOINT ["go", "run", "main.go"]