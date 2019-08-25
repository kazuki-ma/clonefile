package clonefile

import (
	"io/ioutil"
	"testing"
)

func TestName(t *testing.T) {
	src, err := ioutil.TempFile("F:/test", "src.txt")
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024*1024)

	for i := 0; i < 1024*1024; i++ {
		buf[i] = byte('a' + i%16)
	}

	for i := 0; i < 1; i++ {
		src.Write(buf)
	}
	src.WriteString("TEST")
	src.Sync()

	dst, err := ioutil.TempFile("F:/test", "dst.txt")
	if err != nil {
		t.Fatal(err)
	}

	success, err := ByFD(src, dst)
	if err != nil {
		t.Fatalf("err:%+v success:%v", err, success)
	}
}
