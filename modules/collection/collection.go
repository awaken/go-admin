package collection

type Collection []map[string]interface{}

// Where filters the collection by a given key / value pair.
func (c Collection) Where(key string, values ...interface{}) Collection {
	var d []map[string]interface{}
	switch len(values) {
	case 0:
		for _, m := range c {
			if isTrue(m[key]) {
				d = append(d, m)
			}
		}
	case 1:
		v := values[0]
		for _, m := range c {
			if m[key] == v {
				d = append(d, m)
			}
		}
	default:
		_ = values[1]
		if values[0].(string) == "=" {
			v := values[1]
			for _, m := range c {
				if m[key] == v {
					d = append(d, m)
				}
			}
		}
	}
	return d
}

func (c Collection) Length() int {
	return len(c)
}

func (c Collection) FirstGet(key string) interface{} {
	return c[0][key]
}

func isTrue(a interface{}) bool {
	switch t := a.(type) {
	case uint:
		return t != 0
	case uint8:
		return t != 0
	case uint16:
		return t != 0
	case uint32:
		return t != 0
	case uint64:
		return t != 0
	case int:
		return t != 0
	case int8:
		return t != 0
	case int16:
		return t != 0
	case int32:
		return t != 0
	case int64:
		return t != 0
	case float32:
		return t != 0
	case float64:
		return t != 0
	case string:
		return t != ""
	case bool:
		return t
	default:
		return false
	}
}
