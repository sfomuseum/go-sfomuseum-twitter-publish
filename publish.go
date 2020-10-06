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
	"time"
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

	tweet_id := id_rsp.Int()

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

		created_rsp := gjson.GetBytes(tweet_body, "created")

		if !created_rsp.Exists() {
			return errors.New("Missing created timestamp")
		}

		created_ts := created_rsp.Int()
		created_t := time.Unix(created_ts, 0)

		edtf_t := created_t.Format(time.RFC3339)

		wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:inception", edtf_t)

		if err != nil {
			return err
		}

		wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:cessation", edtf_t)

		if err != nil {
			return err
		}
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:name", "FIX ME")

	if err != nil {
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:concordances.twitter:id", tweet_id)

	if err != nil {
		return err
	}

	var tw interface{}

	err = json.Unmarshal(tweet_body, &tw)

	if err != nil {
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.twitter:tweet", tw)

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

	// Null Terminal
	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-architecture/master/data/115/916/086/9/1159160869.geojson

	lat := 37.616356
	lon := -122.386166

	feature := map[string]interface{}{
		"type": "Feature",
		"properties": map[string]interface{}{
			"sfomuseum:placetype": "tweet",
			"src:geom":            "sfomuseum",
			"wof:country":         "US",
			"wof:parent_id":       1159160869,
			"wof:placetype":       "custom",
			"wof:repo":            "sfomuseum-data-twitter",
		},
		"geometry": map[string]interface{}{
			"type":        "Point",
			"coordinates": [2]float64{lon, lat},
		},
	}

	return json.Marshal(feature)
}
