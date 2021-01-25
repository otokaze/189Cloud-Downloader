package dao

import (
	"context"
	"testing"
)

func TestDownload(t *testing.T) {
	url := "https://cloud.189.cn/downloadFile.action?fileStr=055BFCD4BAD9FFC956DEEC776D7CA34D4C8C52BB381C3E1D0910DB3FEA97E336F2B4B5E88FD3A291D1DF532FC87F7E23E019FD82B29FA0D0BD5CF47C&downloadType=1"
	if err := testDao.Download(context.Background(), url, "/tmp", 10); err != nil {
		t.Errorf("testDao.Download(%s) error(%v)", url, err)
	}
}
