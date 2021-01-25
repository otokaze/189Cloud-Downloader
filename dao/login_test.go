package dao

import (
	"context"
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {
	info, err := testDao.Login(context.Background(), "", "")
	if err != nil {
		t.Errorf("testDao.Login error(%v)", err)
	}
	fmt.Printf("=======info(%+v)", info)
}
