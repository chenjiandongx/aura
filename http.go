package aura

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *Registry) apiHealth(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok\n"))
}

func (r *Registry) apiMetadata(w http.ResponseWriter, req *http.Request) {
	mds := make([]*MetaData, 0)
	for _, md := range r.metadata {
		mds = append(mds, md)
	}

	bs, err := json.Marshal(mds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bs)
}

// todo
func (r *Registry) apiStats(w http.ResponseWriter, req *http.Request) {

}

func (r *Registry) Serve(address string) {
	http.HandleFunc("/-/health", r.apiHealth)
	http.HandleFunc("/-/metadata", r.apiMetadata)
	http.HandleFunc("/-/stats", r.apiStats)
	if err := http.ListenAndServe(address, nil); err != nil {
		panic(fmt.Sprintf("failed to start http server(%s): %+v", address, err))
	}
}
