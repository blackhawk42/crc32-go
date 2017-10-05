package main

import(
	"fmt"
	"flag"
	"hash/crc32"
	"io"
	"os"
)

type Crc32Report struct {
	Filename string
	Checksum uint32
	Err error
}

func (r *Crc32Report) ChecksumToHex() string {
	return fmt.Sprintf("%0.8X", r.Checksum)
}

func (r *Crc32Report) Report() string {
	if r.Err == nil {
		return fmt.Sprintf("%s: %s", r.Filename, r.ChecksumToHex())
	} else {
		return fmt.Sprintf("Error while reading %s: %v", r.Filename, r.Err)
	}
}

func main() {
	flag.Parse()
	ch := make(chan *Crc32Report)
	
	for _, filename := range flag.Args() {
		go crc32File(filename, ch)
	}
	
	for range flag.Args() {
		fmt.Println((<-ch).Report())
	}
}

func crc32File(filename string, ch chan *Crc32Report) {
	report := &Crc32Report{Filename: filename}
	
	f, err := os.Open(filename)
	if err != nil {
		report.Err = err
		ch <- report
		return
	}
	defer f.Close()
	
	hash := crc32.NewIEEE()
	
	_, err = io.Copy(hash, f)
	if err != nil {
		report.Err = err
		ch <- report
		return
	}
	
	report.Checksum = hash.Sum32()
	report.Err = nil
	ch <- report
	
	return
}
