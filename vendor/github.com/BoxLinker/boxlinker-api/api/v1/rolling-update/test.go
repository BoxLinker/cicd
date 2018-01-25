package rolling_update

import "net/http"

func (a *Api) Test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("rolling update test"))
}
