package properties

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func EnsureIsSFO(feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, "properties.sfomuseum:is_sfo")

	if rsp.Exists() {
		return feature, nil
	}

	return sjson.SetBytes(feature, "sfomuseum:is_sfo", -1)
}
