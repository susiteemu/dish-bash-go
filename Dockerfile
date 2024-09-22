FROM amd64/alpine:latest

WORKDIR /app

COPY ["dish-bash-go", "*.html", "./"]
ADD ["assets", "./assets"]
ADD ["css", "./css"]

ENTRYPOINT ["./dish-bash-go"]
