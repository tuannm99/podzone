package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/tuannm99/podzone/services/storefront"
)

func main() {
	component := storefront.Hello("John")
	// _ = component.Render(context.Background(), os.Stdout)

	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on :3000")
	_ = http.ListenAndServe(":3000", nil)
}
