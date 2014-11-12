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

// A Scale is a sequence of pitches that can be applied relative to a base frequency.
type Scale struct {
	Description string
	Pitches     []Pitch
}

// Freqs returns one octave of frequencies in the scale, starting at and including
// the given base frequency.
func (s Scale) Freqs(base float64) []float64 {
	f := make([]float64, len(s.Pitches)+1)
	f[0] = base
	for i, p := range s.Pitches {
		f[i+1] = p.Freq(base)
	}
	return f
}

// Pitch represents one scale pitch.
type Pitch interface {
	// Freq returns the pitch frequency relative to the given base.
	Freq(base float64) float64

	String() string
}

// A RatioPitch is a pitch which is the ratio of two integers.
type RatioPitch struct {
	N, D int64
}

func (p RatioPitch) String() string {
	return fmt.Sprintf("%d/%d", p.N, p.D)
}

func (p RatioPitch) Freq(f float64) float64 {
	return float64(p.N) * f / float64(p.D)
}

// A CentsPitch represents a pitch given in units of cents.
type CentsPitch float64

func (p CentsPitch) String() string {
	return fmt.Sprintf("%f", p)
}

func (p CentsPitch) Freq(f float64) float64 {
	return f * math.Exp2(float64(p)/1200.0)
}

// Read reads a Scale from the given reader.
// The input is assumed to be a file in scl format and is consumed until EOF is reached.
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

// Write writes a scale to the given writer in scl format.
// If name is a non-empty string it is written in a comment at the beginning of the output,
// as is customary in .scl files.
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
