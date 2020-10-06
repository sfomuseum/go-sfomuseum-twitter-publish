package publish

import (
	"context"
	"log"
)

type PublishOptions struct {
}

func PublishTweet(ctx context.Context, opts *PublishOptions, tweet []byte) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	log.Println(string(tweet))
	return nil
}
