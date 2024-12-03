GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/twitter-publish cmd/twitter-publish/main.go

debug:
	go run -mod $(GOMOD)  -ldflags="$(LDFLAGS)" \
		cmd/twitter-publish/main.go \
		-reader-uri repo:///usr/local/data/sfomuseum-data-socialmedia-twitter \
		-writer-uri repo:///usr/local/data/sfomuseum-data-socialmedia-twitter \
		-iterator-uri directory:// \
		-iterator-source /usr/local/data/sfomuseum-data-socialmedia-twitter/data \
		-tweets-uri $(TWEETS)

docker-tweets:
	docker buildx build --platform=linux/amd64 --no-cache=true -f Dockerfile -t sfomuseum-twitter-publish .	
