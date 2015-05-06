package verdb

import (
	"reflect"
	"strings"
)

func getVal(rdoc map[string]interface{}, ks []string) interface{} {
	for i, k := range ks {
		if len(ks)-1 == i {
			return rdoc[k]
		}

		_, ok := rdoc[k]
		if !ok {
			break
		}
		tdoc, ok := rdoc[k].(map[string]interface{})
		if !ok {
			break
		}
		rdoc = tdoc
	}
	return nil
}

// compare two docs
func isChanged(ldoc, rdoc map[string]interface{}, keys []string) bool {
	for _, key := range keys {
		parts := strings.Split(key, ".")
		va := SelectVals(ldoc, parts)
		vb := SelectVals(rdoc, parts)
		if !reflect.DeepEqual(va, vb) {
			return true
		}
	}
	return false
}

func SelectVals(m map[string]interface{}, ks []string) []interface{} {
	var nexts = []interface{}{m}
	var vals []interface{}
	for i, k := range ks {
		var nnexts []interface{}
		for _, next := range nexts {
			switch tnext := next.(type) {
			case map[string]interface{}:
				if tnext[k] == nil {
					continue
				}
				if i == len(ks)-1 {
					vals = append(vals, tnext[k])
				} else {
					nnexts = append(nnexts, tnext[k])
				}

			case []interface{}:
				for _, v := range tnext {
					if tm, ok := v.(map[string]interface{}); ok {
						if tm[k] == nil {
							continue
						}
						if i == len(ks)-1 {
							vals = append(vals, tm[k])
						} else {
							nnexts = append(nnexts, tm[k])
						}
					}
				}
			}
		}
		if len(nnexts) == 0 {
			break
		}
		nexts = nnexts
	}
	return vals
}
