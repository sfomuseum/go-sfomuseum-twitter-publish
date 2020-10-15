package main

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-sfomuseum-export"
	"github.com/sfomuseum/go-sfomuseum-twitter"
	"github.com/sfomuseum/go-sfomuseum-twitter-publish"
	"github.com/sfomuseum/go-sfomuseum-twitter/walk"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-whosonfirst-export/options"
	"github.com/whosonfirst/go-writer"
	_ "gocloud.dev/blob/fileblob"

	"log"
)

func main() {

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	indexer_uri := flag.String("indexer-uri", "git://", "A valid whosonfirst/go-whosonfirst-index URI")
	indexer_path := flag.String("indexer-path", "git@github.com:sfomuseum-data/sfomuseum-data-socialmedia-twitter.git", "...")

	reader_uri := flag.String("reader-uri", "", "A valid whosonfirst/go-reader URI")
	writer_uri := flag.String("writer-uri", "", "A valid whosonfirst/go-writer URI")

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

	rdr, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatal(err)
	}

	wrtr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatal(err)
	}

	exprtr_opts, err := export.NewDefaultOptions()

	if err != nil {
		log.Fatal(err)
	}

	exprtr, err := export.NewSFOMuseumExporter(exprtr_opts)

	if err != nil {
		log.Fatal(err)
	}

	lookup, err := publish.BuildLookup(ctx, *indexer_uri, *indexer_path)

	if err != nil {
		log.Fatalf("Failed to build lookup, %v", err)
	}

	publish_opts := &publish.PublishOptions{
		Lookup:   lookup,
		Reader:   rdr,
		Writer:   wrtr,
		Exporter: exprtr,
	}

	max_procs := 10
	throttle := make(chan bool, max_procs)

	for i := 0; i < max_procs; i++ {
		throttle <- true
	}

	cb := func(ctx context.Context, body []byte) error {

		<-throttle

		defer func() {
			throttle <- true
		}()

		return publish.PublishTweet(ctx, publish_opts, body)
	}

	err = walk.WalkTweetsWithCallback(ctx, tweets_fh, cb)

	if err != nil {
		log.Fatal(err)
	}

}
