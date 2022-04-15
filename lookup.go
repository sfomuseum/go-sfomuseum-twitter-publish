package publish

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"io"
	"io/ioutil"
	_ "log"
	"sync"
	"sync/atomic"
)

func BuildLookup(ctx context.Context, indexer_uri string, indexer_path string) (*sync.Map, error) {

	lookup := new(sync.Map)
	count := int32(0)

	indexer_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
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
		return nil, err
	}

	err = iter.IterateURIs(ctx, indexer_path)

	if err != nil {
		return nil, err
	}

	return lookup, nil
}
