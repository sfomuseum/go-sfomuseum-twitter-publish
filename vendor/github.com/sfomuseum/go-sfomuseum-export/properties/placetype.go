package properties

import (
	"errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsurePlacetype(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "properties.sfomuseum:placetype")

	if !rsp.Exists() {
		return feature, errors.New("missing sfomuseum:placetype")
	}

	return sjson.SetBytes(feature, "wof:placetype_alt", rsp.String())
}
