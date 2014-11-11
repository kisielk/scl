package scl

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var corpus string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&corpus, "corpus", filepath.Join(wd, "scales/"), "Directory containing scales corpus")
	flag.Parse()
}

const exampleMeanquar = `! meanquar.scl
!
1/4-comma meantone scale. Pietro Aaron's temperament (1523)
 12
!
 76.04900
 193.15686
 310.26471
 5/4
 503.42157
 579.47057
 696.57843
 25/16
 889.73529
 1006.84314
 1082.89214
 2/1
`

func TestReadMeanquar(t *testing.T) {
	f := bytes.NewBufferString(exampleMeanquar)
	s, err := Read(f)
	if err != nil {
		t.Fatal(err)
	}
	if want := "1/4-comma meantone scale. Pietro Aaron's temperament (1523)"; s.Description != want {
		t.Fatalf("Bad decsription, got %q want %q", s.Description, want)
	}
}

func TestCorpus(t *testing.T) {
	dir, err := os.Open(corpus)
	if err != nil {
		t.Fatal(err)
	}
	names, err := dir.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		f, err := os.Open(filepath.Join(corpus, name))
		if err != nil {
			t.Error(err)
			continue
		}
		scale, err := Read(f)
		if err != nil {
			t.Errorf("Couldn't read %s: %s", name, err)
		}
		if len(scale.Description) == 0 {
			t.Errorf("%s: 0 length description", name)
		}
		f.Close()
	}
}
