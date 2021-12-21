package coordinator

import "net/http"

type Coordinator struct {
	Handler *http.Handler
}

func New() {
}
