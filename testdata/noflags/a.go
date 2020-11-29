package noflags

import "encoding/json"

func Marshal(x interface{}) ([]byte, error) {
	return json.Marshal(x)
}
