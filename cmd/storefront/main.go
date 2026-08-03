package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/tuannm99/podzone/internal/storefront"
)

func main() {
	component := storefront.Hello("John")
	// _ = component.Render(context.Background(), os.Stdout)

	http.Handle("/", templ.Handler(component))
	// PS10, docs/09-pzep/PZEP-0019-staging-k8s-survivability-track-a.md:
	// the staging manifest's probe already assumed this route existed -
	// it didn't, so uncommenting the probe as-is would have crashlooped.
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	fmt.Println("Listening on :3000")
	_ = http.ListenAndServe(":3000", nil)
}
