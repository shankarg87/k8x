FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY k8x .

CMD ["./k8x"]
