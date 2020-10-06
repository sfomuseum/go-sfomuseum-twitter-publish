package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	wof_name := fmt.Sprintf("Twitter message #%d", tweet_id)
	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:name", wof_name)

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

	// Null Terminal - please read these details from source...
	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-architecture/master/data/115/916/086/9/1159160869.geojson

	parent_id := 1159160869
	
	lat := 37.616356
	lon := -122.386166

	hier := map[string]interface{}{
				"building_id":1159160869,
				"campus_id":102527513,
				"continent_id":102191575,
				"country_id":85633793,
				"county_id":102087579,
				"locality_id":85922583,
				"neighbourhood_id":-1,
				"region_id":85688637,
	}
	
	geom := map[string]interface{}{
			"type":        "Point",
			"coordinates": [2]float64{lon, lat},
	}
	
	feature := map[string]interface{}{
		"type": "Feature",
		"properties": map[string]interface{}{
			"sfomuseum:placetype": "tweet",
			"src:geom":            "sfomuseum",
			"wof:country":         "US",
			"wof:parent_id":       parent_id,
			"wof:placetype":       "custom",
			"wof:repo":            "sfomuseum-data-twitter",
			"wof:hierarchy": hier,
		},
		"geometry": geom,
	}

	return json.Marshal(feature)
}
