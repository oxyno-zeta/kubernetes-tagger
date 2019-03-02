FROM alpine:3.9

COPY ./kubernetes-tagger /bin/kubernetes-tagger

ENTRYPOINT [ "/bin/kubernetes-tagger" ]
