FROM golang:alpine as build

WORKDIR $GOPATH/src/github.com/Ahton89/vacancies_scrapper/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN apk --no-cache add git bash
RUN go build -o /go/bin/vacancies_scrapper ./cmd/vacancies_scrapper

FROM alpine:3.18
WORKDIR /app
RUN apk --no-cache add ca-certificates git
COPY --from=build /go/bin/vacancies_scrapper /bin/

ENTRYPOINT ["vacancies_scrapper"]