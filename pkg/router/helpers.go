package router

import (
	"db_explorer/pkg/router/routertrie"
	"net/http"
)

func PathValue(r *http.Request, key string) string {
	v := r.Context().Value(routertrie.CtxParamKey(key))

	str, ok := v.(string)
	if ok {
		return str
	}

	return ""
}
