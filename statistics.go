package hare

import (
	"fmt"
	"net/http"
)

/**
 * Prepares and dispatches a response via ResponseWriter for a stats API request
 */
func apiRequestStats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v", Statistics)
}

func statisticsIncrement(stat string) {
	Statistics[stat] = Statistics[stat] + 1
}
