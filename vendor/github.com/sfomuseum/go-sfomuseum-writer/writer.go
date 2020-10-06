package writer

import (
	"bytes"
	"context"
	"errors"
	"github.com/sfomuseum/go-sfomuseum-export"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	go_writer "github.com/whosonfirst/go-writer"
	"io/ioutil"
)

func WriteFeature(ctx context.Context, wr go_writer.Writer, f geojson.Feature) (int64, error) {
	return WriteFeatureBytes(ctx, wr, f.Bytes())
}

func WriteFeatureBytes(ctx context.Context, wr go_writer.Writer, body []byte) (int64, error) {

	ex_opts, err := export.NewDefaultOptions()

	if err != nil {
		return -1, err
	}

	exporter, err := export.NewSFOMuseumExporter(ex_opts)

	if err != nil {
		return -1, err
	}

	ex_body, err := exporter.Export(body)

	if err != nil {
		return -1, err
	}

	id_rsp := gjson.GetBytes(ex_body, "properties.wof:id")

	if !id_rsp.Exists() {
		return -1, errors.New("Missing 'properties.wof:id' property")
	}

	id := id_rsp.Int()

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return -1, err
	}

	br := bytes.NewReader(ex_body)
	fh := ioutil.NopCloser(br)

	err = wr.Write(ctx, rel_path, fh)

	if err != nil {
		return -1, err
	}

	return id, nil
}
