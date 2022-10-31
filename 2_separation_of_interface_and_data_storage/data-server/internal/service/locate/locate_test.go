package locate

import (
	"testing"
)

func TestLocate(t *testing.T) {
	if !locate("locate.go") {
		t.Error("locate.go should be found")
	}
	if !locate("../dataServer.go") {
		t.Error("../dataServer.go should be found")
	}
	if locate("not-exists.go") {
		t.Error("not-exists.go should not be found")
	}
}
