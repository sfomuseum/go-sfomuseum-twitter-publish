package publish

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sfomuseum/go-sfomuseum-twitter/document"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/exporter"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io/ioutil"
	"log"
	"sync"
)

type PublishOptions struct {
	Lookup   *sync.Map
	Reader   reader.Reader
	Writer   writer.Writer
	Exporter exporter.Exporter
}

func PublishTweet(ctx context.Context, opts *PublishOptions, body []byte) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	id_rsp := gjson.GetBytes(body, "id")

	if !id_rsp.Exists() {
		return errors.New("Missing 'id' property")
	}

	tweet_id := id_rsp.String()

	tweet_body, err := document.AppendCreatedAtTimestamp(ctx, body)

	if err != nil {
		return nil
	}

	pointer, ok := opts.Lookup.Load(tweet_id)

	var wof_record []byte

	if ok {

		wof_id := pointer.(int64)
		rel_path, err := uri.Id2RelPath(wof_id)

		if err != nil {
			return err
		}

		wof_fh, err := opts.Reader.Read(ctx, rel_path)

		if err != nil {
			return err
		}

		defer wof_fh.Close()

		wof_body, err := ioutil.ReadAll(wof_fh)

		if err != nil {
			return err
		}

		wof_record = wof_body

	} else {

		new_record, err := newWOFRecord(ctx)

		if err != nil {
			return err
		}

		wof_record = new_record
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:name", "FIX ME")

	if err != nil {
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.concordances.twitter:id", tweet_id)

	if err != nil {
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.tweet", tweet_body)

	if err != nil {
		return err
	}

	wof_record, err = opts.Exporter.Export(wof_record)

	if err != nil {
		return err
	}

	log.Println(string(tweet_body))
	return nil
}

func newWOFRecord(ctx context.Context) ([]byte, error) {

	feature := map[string]interface{}{
		"type": "Feature",
		"properties": map[string]interface{}{
			"sfomuseum:placetype": "tweet",
			"wof:country":         "US",
			"wof:placetype":       "custom",
		},
		"geometry": map[string]interface{}{
			"type":        "Point",
			"coordinates": [2]float64{0.0, 0.0},
		},
	}

	return json.Marshal(feature)
}
