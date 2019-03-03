FROM alpine:3.9

RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*

COPY ./kubernetes-tagger /bin/kubernetes-tagger

ENTRYPOINT [ "/bin/kubernetes-tagger" ]
