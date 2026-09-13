package generic

import "gosalusa.com/database/dialects"

func group(r dialects.RawQuery, err error) (dialects.RawQuery, error) {
	r.SQL = "(" + r.SQL + ")"
	return r, err
}
