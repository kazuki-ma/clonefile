package clonefile

import (
	"crypto"
	_ "crypto/sha256"
	"encoding/hex"
	"github.com/kazuki-ma/clonefile/gosrc/clonefile"
	"io"
	"log"
	"os"
	"path/filepath"
)

var wd, _ = os.Getwd()

func main() {
	args := os.Args[1:]

	sizeMap := make(map[int64][]string)

	for _, dir := range args {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if v, ok := sizeMap[info.Size()]; ok {
				sizeMap[info.Size()] = append(v, path)
			} else {
				sizeMap[info.Size()] = []string{path}
			}

			return nil
		})
	}

	for k, v := range sizeMap {
		if len(v) <= 1 {
			delete(sizeMap, k)
			continue
		}

		if err := dedupMaybeSame(v); err != nil {
			panic(err)
		}
	}
}

func dedupMaybeSame(files []string) error {
	hashMap := make(map[string][]string)

	for _, v := range files {
		h, err := hash(v)
		if err != nil {
			log.Printf("%s %+v", v, err)
			continue
		}

		if g, ok := hashMap[h]; ok {
			hashMap[h] = append(g, v)
		} else {
			hashMap[h] = []string{v}
		}
	}

	for _, v := range hashMap {
		if len(v) <= 1 {
			continue
		}

		err := dedup(v) // dedup same file hash
		if err != nil {
			return err
		}
	}

	return nil
}

func dedup(files []string) error {
	src := files[0]

	for _, dst := range files[1:] {
		log.Printf("clonefile %s > %s", src, dst)
		if _, err := clonefile.ByPath(src, dst); err != nil {
			return err
		}
	}

	return nil
}

func hash(p string) (hash string, err error) {
	sha256 := crypto.SHA256.New()

	file, err := os.Open(p)
	if err != nil {
		return
	}

	_, err = io.Copy(sha256, file)
	if err != nil {
		return
	}

	return hex.EncodeToString(sha256.Sum(nil)), nil
}
