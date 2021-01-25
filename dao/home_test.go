package dao

import (
	"context"
	"testing"
)

func TestGetHomeDirList(t *testing.T) {
	testDao.Login(context.Background(), "", "")
	dirs, paths, err := testDao.GetHomeDirList(context.Background(), 1, 60, "ASC")
	if err != nil {
		t.Fatalf("testDao.GetHomeDirList() error(%v)", err)
	}
	t.Logf("=====dir(%+v)\n=======path(%+v)\n", dirs[1], paths)
}
