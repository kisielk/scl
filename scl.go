// Package scl is for reading and writing files in the Scala scale format.
// See http://www.huygens-fokker.org/scala/scl_format.html for more details about the file format.
package scl

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Scale struct {
	Description string
	Pitches     []Pitch
}

type Pitch interface {
	String() string
}

type RatioPitch struct {
	N, D int64
}

func (p RatioPitch) String() string {
	return fmt.Sprintf("%d/%d", p.N, p.D)
}

type CentsPitch float64

func (p CentsPitch) String() string {
	return fmt.Sprintf("%f", p)
}

func Read(r io.Reader) (Scale, error) {
	var (
		scale      Scale
		readDesc   bool
		readNum    bool
		numPitches int64
		err        error
	)
	s := bufio.NewScanner(r)
	for i := 1; s.Scan(); i++ {
		line := s.Text()
		if strings.HasPrefix(line, "!") {
			continue
		} else if !readDesc {
			scale.Description = line
			readDesc = true
		} else if !readNum {
			line = strings.TrimSpace(line)
			numPitches, err = strconv.ParseInt(line, 10, 64)
			if err != nil {
				return scale, fmt.Errorf("malformed number of pitches: %s", line)
			}
			readNum = true
		} else {
			pitch, err := parsePitch(line)
			if err != nil {
				fmt.Errorf("Parse error on line %d: %s", i, err)
			}
			scale.Pitches = append(scale.Pitches, pitch)
		}
	}
	if len(scale.Pitches) != int(numPitches) {
		return scale, fmt.Errorf("read %d pitches but expected %d", len(scale.Pitches), numPitches)
	}
	return scale, s.Err()
}

func parsePitch(s string) (Pitch, error) {
	s = strings.Fields(s)[0]
	if strings.ContainsRune(s, '.') {
		v, err := strconv.ParseFloat(s, 64)
		return CentsPitch(v), err
	}
	var (
		p     = RatioPitch{1, 1}
		parts = strings.Split(s, "/")
		err   error
	)
	switch len(parts) {
	case 2:
		d, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return p, err
		}
		p.D = d
		fallthrough
	case 1:
		n, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return p, err
		}
		p.N = n
	default:
		err = fmt.Errorf("malformed pitch ratio: %s", s)
	}
	if p.N <= 0 || p.D <= 0 {
		err = fmt.Errorf("malformed pitch ratio: %s", s)
	}
	return p, err
}
