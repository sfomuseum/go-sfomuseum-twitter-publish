package main

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-sfomuseum-twitter"
	"github.com/sfomuseum/go-sfomuseum-twitter-publish"
	"github.com/sfomuseum/go-sfomuseum-twitter/walk"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-writer"
	_ "gocloud.dev/blob/fileblob"
	"log"
	"sync"
)

func main() {

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	reader_uri := flag.String("reader-uri", "", "A valid whosonfirst/go-reader URI")
	writer_uri := flag.String("reader-uri", "", "A valid whosonfirst/go-writer URI")

	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	open_opts := &twitter.OpenTweetsOptions{
		TrimPrefix: *trim_prefix,
	}

	tweets_fh, err := twitter.OpenTweets(ctx, *tweets_uri, open_opts)

	if err != nil {
		log.Fatalf("Failed to open %s, %v", *tweets_uri, err)
	}

	defer tweets_fh.Close()

	lookup := new(sync.Map)

	rdr, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatal(err)
	}

	wrtr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatal(err)
	}

	publish_opts := &publish.PublishOptions{
		Lookup: lookup,
		Reader: rdr,
		Writer: wrtr,
	}

	cb := func(ctx context.Context, body []byte) error {
		return publish.PublishTweet(ctx, publish_opts, body)
	}

	err = walk.WalkTweetsWithCallback(ctx, tweets_fh, cb)

	if err != nil {
		log.Fatal(err)
	}

}
