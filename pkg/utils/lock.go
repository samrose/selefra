package utils

import (
	"github.com/selefra/selefra-utils/pkg/id_util"
	"os"
)

// BuildLockOwnerId The current host name is placed in the owner of the lock so that it is easy to identify who is holding the lock
// This place is mainly used for database locks
func BuildLockOwnerId() string {
	hostname, err := os.Hostname()
	id := id_util.RandomId()
	if err != nil {
		return "unknown-hostname-" + id
	} else {
		return hostname + "-" + id
	}
}
