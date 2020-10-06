package main

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-sfomuseum-twitter"
	"github.com/sfomuseum/go-sfomuseum-twitter-publish"
	"github.com/sfomuseum/go-sfomuseum-twitter/walk"
	_ "gocloud.dev/blob/fileblob"
	"log"
)

func main() {

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

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

	publish_opts := &publish.PublishOptions{}

	cb := func(ctx context.Context, body []byte) error {
		return publish.PublishTweet(ctx, publish_opts, body)
	}

	err = walk.WalkTweetsWithCallback(ctx, tweets_fh, cb)

	if err != nil {
		log.Fatal(err)
	}

}
