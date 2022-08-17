FROM alpine:latest 

RUN mkdir /app 

COPY logPose /app

CMD ["/app/logPose"]