package publish

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index-git"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"io"
	"io/ioutil"
	"log"
	"sync"
)

func BuildLookup(ctx context.Context, indexer_uri string, indexer_path string) (*sync.Map, error) {

	lookup := new(sync.Map)

	indexer_cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			return err
		}

		wof_rsp := gjson.GetBytes(body, "properties.wof:id")

		if !wof_rsp.Exists() {
			return fmt.Errorf("Missing WOF ID")
		}

		wof_id := wof_rsp.Int()

		tweet_rsp := gjson.GetBytes(body, "properties.wof:concordances.twitter:id")

		if !tweet_rsp.Exists() {
			return fmt.Errorf("Missing Twitter ID for record %d", wof_id)
		}

		tweet_id := tweet_rsp.Int()

		log.Println("Store", tweet_id, wof_id)
		lookup.Store(tweet_id, wof_id)
		return nil
	}

	indexer, err := index.NewIndexer(indexer_uri, indexer_cb)

	if err != nil {
		return nil, err
	}

	err = indexer.IndexPath(indexer_path)

	if err != nil {
		return nil, err
	}

	return lookup, nil
}
