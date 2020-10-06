package publish

import (
	"context"
	"encoding/json"
	"errors"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	"github.com/sfomuseum/go-sfomuseum-twitter/document"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/exporter"
	"github.com/whosonfirst/go-writer"
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

		wof_body, err := sfom_reader.LoadBytesFromID(ctx, opts.Reader, wof_id)

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

	id, err := sfom_writer.WriteFeatureBytes(ctx, opts.Writer, wof_record)

	if err != nil {
		return err
	}

	log.Printf("Wrote %d\n", id)
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
