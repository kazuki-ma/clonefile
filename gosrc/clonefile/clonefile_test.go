package clonefile

import (
	"io/ioutil"
	"testing"
)

func TestName(t *testing.T) {
	src, err := ioutil.TempFile("", "src")
	if err != nil {
		t.Fatal(err)
	}
	dst, err := ioutil.TempFile("", "dst")
	if err != nil {
		t.Fatal(err)
	}

	success, err := ByFD(src, dst)
	if err != nil == success {
		t.Fatalf("err:%s success:%v", err, success)
	}
}
