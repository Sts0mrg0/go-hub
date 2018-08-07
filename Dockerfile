FROM ubuntu:16.04
WORKDIR /app
EXPOSE 8080
ADD hub /app/

CMD ["./hub"]
