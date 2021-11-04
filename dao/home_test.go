package dao

import (
	"context"
	"testing"
)

func TestGetHomeDirList(t *testing.T) {
	testDao.Login(context.Background(), "", "")
	dirs, err := testDao.GetHomeDirList(context.Background(), 1, 60, "ASC")
	if err != nil {
		t.Fatalf("testDao.GetHomeDirList() error(%v)", err)
	}
	t.Logf("dir(%+v)", dirs[1])
}
