package main

import (
	"context"
	"github.com/sfomuseum/go-sfomuseum-twitter/walk"
	"github.com/sfomuseum/go-sfomuseum-twitter-publish"	
	_ "gocloud.dev/blob/fileblob"
	"flag"
	"log"
)

func main() {

	tweets_uri := flag.String("tweets", "", "")
	
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	publish_opts := &publish.PublishOptions{
	}
	
	err_ch := make(chan error)
	tweet_ch := make(chan []byte)
	done_ch := make(chan bool)

	walk_opts := &walk.WalkOptions{
		DoneChannel: done_ch,
		ErrorChannel: err_ch,
		TweetChannel: tweet_ch,
	}
	
	go walk.WalkTweets(ctx, walk_opts, *tweets_uri,)

	working := true
	
	for {
		select {
		case <- done_ch:
			working = false
		case err := <- err_ch:
			log.Println(err)
			cancel()
		case body := <- tweet_ch:

			err := publish.PublishTweet(ctx, publish_opts, body)

			if err != nil {
				err_ch <- err
			}
		}
		
		if !working {
			break
		}
	}
	
}
