make: 
	go build -o bin/mnemo mnemo.go

run:
	bin/mnemo

tag:
	git tag -af v0.0.22 -m "mnemo v0.0.22" &&\
	git push --tags

publish-proxy:
	GOPROXY=proxy.golang.org go list -m github.com/snburman/mnemo@v0.0.22

lookup:
	curl https://sum.golang.org/lookup/github.com/snburman/mnemo@v0.0.22

publish:
	git tag -af v0.0.22 -m "mnemo v0.0.22" &&\
	git push --tags &&\
	GOPROXY=proxy.golang.org go list -m github.com/snburman/mnemo@v0.0.22 &&\
	curl https://sum.golang.org/lookup/github.com/snburman/mnemo@v0.0.22