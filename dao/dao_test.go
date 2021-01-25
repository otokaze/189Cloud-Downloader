package dao

import (
	"os"
	"testing"
)

var testDao *dao

func TestMain(m *testing.M) {
	testDao = New()
	os.Exit(m.Run())
}
