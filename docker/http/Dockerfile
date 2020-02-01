FROM docker.pkg.github.com/dacci/juno-go/base

WORKDIR /root

COPY http.json .

EXPOSE 8080

CMD ["/usr/local/bin/juno-go", "-c", "http.json"]
