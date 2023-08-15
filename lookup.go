package publish

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func BuildLookup(ctx context.Context, indexer_uri string, indexer_path string) (*sync.Map, error) {

	lookup := new(sync.Map)
	count := int32(0)

	indexer_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		wof_rsp := gjson.GetBytes(body, "properties.wof:id")

		if !wof_rsp.Exists() {
			return fmt.Errorf("Missing WOF ID")
		}

		wof_id := wof_rsp.Int()

		tweet_rsp := gjson.GetBytes(body, "properties.twitter:tweet.id")

		if !tweet_rsp.Exists() {
			return fmt.Errorf("Missing Twitter ID for record %d", wof_id)
		}

		tweet_id := tweet_rsp.Int()

		lookup.Store(tweet_id, wof_id)

		atomic.AddInt32(&count, 1)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, indexer_uri, indexer_cb)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new iterator for %s, %w", indexer_uri, err)
	}

	err = iter.IterateURIs(ctx, indexer_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to iterate URIs, %w", err)
	}

	return lookup, nil
}
