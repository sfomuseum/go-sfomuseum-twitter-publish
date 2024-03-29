GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/twitter-publish cmd/twitter-publish/main.go

debug:
	go run -mod $(GOMOD)  cmd/twitter-publish/main.go -reader-uri fs:///usr/local/data/sfomuseum-data-socialmedia-twitter/data -writer-uri fs:///usr/local/data/sfomuseum-data-socialmedia-twitter/data -iterator-uri directory:// -iterator-source /usr/local/data/sfomuseum-data-socialmedia-twitter/data -tweets-uri $(TWEETS)

docker-tweets:
	docker buildx build --platform=linux/amd64 --no-cache=true -f Dockerfile -t sfomuseum-twitter-publish .	
