
FROM alpine:latest

RUN apk --no-cache add sqlite

WORKDIR /db

EXPOSE 1433

CMD ["sqlite3"]
