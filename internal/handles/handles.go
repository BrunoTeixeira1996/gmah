package handles

import (
	"fmt"
	"net/http"
)

// Handles "/"
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}
