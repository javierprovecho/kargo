FROM alpine

COPY run.sh /run.sh
RUN apk --no-cache add wget ca-certificates && chmod +x /run.sh

ENTRYPOINT ["/run.sh"]
