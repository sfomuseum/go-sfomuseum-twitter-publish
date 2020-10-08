cli:
	go build -mod vendor -o bin/publish cmd/publish/main.go

debug:
	go run -mod vendor cmd/publish/main.go -reader-uri fs:///usr/local/data/sfomuseum-data-twitter/data -writer-uri fs:///usr/local/data/sfomuseum-data-twitter/data -indexer-uri directory -indexer-path /usr/local/data/sfomuseum-data-twitter/data -tweets-uri file:///usr/local/data/twitter/data/tweet.js
