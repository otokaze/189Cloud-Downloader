package dao

import (
	"context"
	"testing"
)

func TestGetShareInfo(t *testing.T) {
	t.Log(testDao.GetShareInfo(context.Background(), "UreAVrBJJfii", "107w"))
}
