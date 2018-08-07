FROM scratch
WORKDIR /app
EXPOSE 8080
ADD hub /app/

CMD ["./hub"]
