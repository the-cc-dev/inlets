dist:
	CGO_ENABLED=0 GOOS=darwin go build -a -installsuffix cgo --ldflags "-s -w" -o inlets-darwin && \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo --ldflags "-s -w" -o inlets && \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -a -installsuffix cgo --ldflags "-s -w" -o inlets-armhf

