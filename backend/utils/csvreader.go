package utils

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"golang.org/x/text/encoding"

	"encoding/csv"

	"github.com/gocarina/gocsv"
)

// CSVReader can read CSV files with header line.
// It is able to skip the garbage before the first valid CSV line (header).
// A valid header line contains at least MinimumHeaderSeparators column delimiters
// specified by Comma
type CSVReader struct {
	reader        *bufio.Reader
	csvReader     *csv.Reader
	lineBuffer    *bytes.Buffer
	garbageBuffer *bytes.Buffer
	headerParsed  bool
	// Comma specifies the delimiter used to separate columns (default is ',')
	Comma rune
	// MinimumHeaderSeparators is the least amount of Comma in a line to consider
	// the line being a header
	MinimumHeaderSeparators int
}

// NewCSVReaderWithEncoding creates a new CSV reader object with encoding interface.
// For encoding it is possible to use https://godoc.org/golang.org/x/text/encoding/charmap
func NewCSVReaderWithEncoding(r io.Reader, encoding encoding.Encoding) *CSVReader {
	return NewCSVReader(encoding.NewDecoder().Reader(r))
}

// NewCSVReader creates a new CSV reader object
func NewCSVReader(r io.Reader) *CSVReader {
	return &CSVReader{
		bufio.NewReader(r),
		nil,
		bytes.NewBuffer([]byte{}),
		bytes.NewBuffer([]byte{}),
		false,
		',',
		3,
	}
}

func (r *CSVReader) Read(dest []byte) (n int, err error) {
	if r.headerParsed {
		read := 0
		if r.lineBuffer != nil {
			// try read from line buffer first
			n, err := r.lineBuffer.Read(dest)
			if err == io.EOF {
				// clean line bufer memory
				r.lineBuffer = nil
			} else {
				read += n
			}
		}
		if read < len(dest) {
			// pass-through more
			moreRead, err := r.reader.Read(dest[read:])
			read += moreRead
			if err != nil {
				return read, err
			}
		}
		return read, nil
	}

	// parse header, skipping unnecessary lines until a header is found
	numSepsInLine := 0
	for {
		readRune, n, err := r.reader.ReadRune()
		if err != nil {
			return 0, err
		} else if n == 0 {
			return 0, errors.New("unable to read input or locate csv header (check Comma)")
		}

		if readRune == r.Comma {
			numSepsInLine++
		}

		_, err = r.lineBuffer.Write([]byte(string(readRune)))
		if err != nil {
			return 0, err
		}

		if readRune == '\n' || readRune == '\r' {
			// we have a complete line which is csv header
			if r.lineBuffer.Len() > 0 && numSepsInLine >= r.MinimumHeaderSeparators {
				r.headerParsed = true
				break
			} else {
				// store the garbage
				r.lineBuffer.WriteTo(r.garbageBuffer)
			}

			r.lineBuffer.Reset()
			numSepsInLine = 0
		}
	}

	if !r.headerParsed {
		return 0, errors.New("csv header not found (must have at least three ';' or ',' in a line)")
	}

	// run the method again, also reading from lineBuffer this time
	return r.Read(dest)
}

// Garbage returns the garbage bytes before the header line
func (r *CSVReader) Garbage() []byte {
	return r.garbageBuffer.Bytes()
}

// GarbageString returns the garbage before the header line as string
func (r *CSVReader) GarbageString() string {
	return r.garbageBuffer.String()
}

// GoCSVReader returns csv.Reader
func (r *CSVReader) GoCSVReader() *csv.Reader {
	if r.csvReader != nil {
		return r.csvReader
	}
	r.csvReader = csv.NewReader(r)
	r.csvReader.Comma = r.Comma
	return r.csvReader
}

// Unmarshal parses the CSV from the reader to the interface, skipping the garbage before the CSV header.
func (r *CSVReader) Unmarshal(out interface{}) error {
	rdr := r.GoCSVReader()
	return gocsv.UnmarshalCSV(rdr, out)
}
