// Network connectivity is required to get the page.
// Tests are using one page using one get or an available file in output directory
// The page is updated by replacing one word which is available using buffer or a file.
package testingfiles

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	techName = "Google"
	myTech   = "MyTech"
	wantf    = "originalpage.html"
	updatedf = "updatedpage.html"
)

// Only one read on the network or filled with the existing want file
var wantb []byte

func TestMain(m *testing.M) {
	resp, err := http.Get("https://about.google/intl/en_be/")

	OutputDir("output")
	if err == nil {
		defer resp.Body.Close()
		if _, err = os.Stat(wantf); os.IsNotExist(err) {
			// File missing, create it
			log.Printf("creating %s file\n", wantf)
			err = ReadCloserToFile(wantf, resp.Body)
			if err != nil {
				log.Fatalf("create want file failed with %v", err)
			}
		}
		// File updates will occur in the tests
		wantb, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		// No network mainly... Let us fill the buffer with the file
		// TODO Check permissions
		f, err := os.OpenFile(wantf, os.O_RDONLY, 777)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		fs, err := os.Stat(wantf)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		wantb = make([]byte, fs.Size())
		n, err := f.Read(wantb)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if n != len(wantb) {
			log.Printf("page is trucated by %d\n", len(wantb)-n)
		}
	}
	os.Exit(m.Run())
}

// Buffer is used as a string and produces a file
// The check is using FileCompare to detect an error
// The error is used for the test and this method by the Benchmark
func GetPageStringToFile(name string) error {
	// got file is identical to want file - no page update
	StringToFile(name, wantb)
	return FileCompare(name, wantf) // second element is the func name
}

// Test creation of a new file with an updated content. Error must be returned by comparison.
func TestPageStringToFile(t *testing.T) {
	var err error
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	// TODO First run fails when file is created. CI fix
	if err := GetPageStringToFile(t.Name()); err != nil && err != io.EOF {

		t.Error(err)
	}
	if err := os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

// Comparing a file to itelf must return nil
func TestFileCompare(t *testing.T) {
	if err := FileCompare(wantf, wantf); err != nil {
		t.Error(err)
	}
}

// Buffer to file, iso String. Then comparing files.
func GetPageBufferToFile(name string) error {
	// got file is rewritten with the updated page
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	BufferToFile(name, wantbuf)
	return FileCompare(name, wantf)
}

// Create a file from a buffer
func TestBufferToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	BufferToFile(t.Name(), b)
	if err := os.Remove(t.Name()); err != nil {
		log.Println(err)
	}

}

func TestBufferCompareNoDiff(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err = BufferCompare(wantbuf, wantf); err == nil {
		t.Error("no difference found")
	}
}

func TestBufferCompare(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err = BufferCompare(wantbuf, t.Name()); err != nil {
		t.Errorf("difference found. %v", err)
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

// Create a file from a ReadCloser (r.Body)
func TestReadCloserToFile(t *testing.T) {
	b := new(bytes.Buffer)
	b.Write(wantb)
	if err := ReadCloserToFile("gotbuffer.html", ioutil.NopCloser(b)); err != nil {
		t.Error(err)
	}
}

func TestReadCloserCompareNoDiff(t *testing.T) {
	var err error
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err := ReadCloserCompare(ioutil.NopCloser(wantbuf), wantf); err == nil {
		t.Error("no difference found")
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

func TestReadCloserCompare(t *testing.T) {
	// Replaces techname to get a different page. A reference file is created.
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	var err error
	if err = os.Remove(t.Name()); err != nil && !os.IsNotExist(err) {
		log.Println(err)
	}
	BufferToFile(t.Name(), wantbuf)
	if err := ReadCloserCompare(ioutil.NopCloser(wantbuf), t.Name()); err != nil {
		t.Errorf("difference found: %v", err)
	}
	if err = os.Remove(t.Name()); err != nil {
		log.Println(err)
	}
}

// Benchmarks
// File operation is the most consuming. One file less means half the time.
// Buffer has a minor advantage over string.
func BenchmarkGetPageStringToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageStringToFile("stringtofile.html"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetPageBufferToFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferToFile(updatedf); err != nil {
			b.Fatal(err)
		}
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageBufferCompare() error {
	i, _, _, _ := runtime.Caller(0)
	fn := ""
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		fn = funcname[1]
	}
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if _, err := os.Stat(fn); err != nil {
		BufferToFile(fn, wantbuf)
	}
	return BufferCompare(wantbuf, fn)
}

func BenchmarkGetPageBufferCompare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageBufferCompare(); err != nil {
			b.Fatal(err)
		}
	}
}

// No got file. Comparing buffer to want file. Got file created only if different
func GetPageReadCloserCompare() error {
	i, _, _, _ := runtime.Caller(0)
	fn := ""
	if funcname := strings.SplitAfter(filepath.Base(runtime.FuncForPC(i).Name()), "."); len(funcname) == 1 {
		return fmt.Errorf("Func name not found")
	} else {
		fn = funcname[0]
	}
	wantbuf := new(bytes.Buffer)
	_, _ = wantbuf.Write(bytes.Replace(wantb, []byte(techName), []byte(myTech), -1))
	if _, err := os.Stat(fn); err != nil {
		BufferToFile(fn, wantbuf)
	}
	return ReadCloserCompare(ioutil.NopCloser(wantbuf), fn)
}

func BenchmarkGetPageReadCloserCompare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if err := GetPageReadCloserCompare(); err != nil {
			b.Fatal(err)
		}
	}
}
