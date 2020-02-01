FROM docker.pkg.github.com/dacci/juno-go/base

WORKDIR /root

COPY socks.json .

EXPOSE 1080

CMD ["/usr/local/bin/juno-go", "-c", "socks.json"]
