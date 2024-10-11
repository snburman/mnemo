make: 
	go build -o bin/mnemo mnemo.go

run:
	bin/mnemo

tag:
	git tag -af v0.0.2-beta -m "mnemo v0.0.2-beta"

publish-proxy:
	GOPROXY=proxy.golang.org go list -m github.com/snburman/mnemo@v0.0.2-beta

lookup:
	curl https://sum.golang.org/lookup/github.com/snburman/mnemo@v0.0.2-beta