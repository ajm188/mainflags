package noflags

import "encoding/json"

func marshal(x interface{}) ([]byte, error) {
	return json.Marshal(x)
}
