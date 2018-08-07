docker run -ti --rm -p 8080:8080 -e "token=arkady" hub

#CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hub .
#только с образом scratch
