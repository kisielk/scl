// Package scl is for reading and writing files in the Scala scale format.
// See http://www.huygens-fokker.org/scala/scl_format.html for more details about the file format.
package scl

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type Scale struct {
	Description string
	Pitches     []Pitch
}

type Pitch interface {
	Freq(base float64) float64
	String() string
}

type RatioPitch struct {
	N, D int64
}

func (p RatioPitch) String() string {
	return fmt.Sprintf("%d/%d", p.N, p.D)
}

func (p RatioPitch) Freq(f float64) float64 {
	return float64(p.N) * f / float64(p.D)
}

type CentsPitch float64

func (p CentsPitch) String() string {
	return fmt.Sprintf("%f", p)
}

func (p CentsPitch) Freq(f float64) float64 {
	return f * math.Exp2(float64(p)/1200.0)
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

func Write(w io.Writer, s Scale, name string) error {
	if name != "" {
		_, err := fmt.Fprintf(w, "! %s\n!\n", name)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, s.Description)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, " %d\n", len(s.Pitches))
	if err != nil {
		return err
	}
	for _, p := range s.Pitches {
		_, err := fmt.Fprintf(w, " %s\n", p)
		if err != nil {
			return err
		}
	}
	return nil
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
