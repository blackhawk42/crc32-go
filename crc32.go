package main

import(
	"fmt"
	"flag"
	"path/filepath"
	"hash/crc32"
	"io"
	"os"
	"sort"
)

// Structs and types

type Crc32Report struct {
	Number int
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

type Crc32ReportCollection []*Crc32Report

func (c Crc32ReportCollection) Len() int {
	return len(c)
}

func (c Crc32ReportCollection) Less(i, j int) bool {
	if c[i].Number < c[j].Number {
		return true
	} else {
		return false
	}
}

func (c Crc32ReportCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Main function

func main() {
	// Flag config
	var sortReports = flag.Bool("s", false, "Sort results in the order given, instead of the random order intrinsic to concurrency")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s: %[1]s [-s] FILE1 [FILE2...]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	
	// Program
	
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	
	ch := make(chan *Crc32Report)
	
	for i, filename := range flag.Args() {
		go crc32File(filename, i, ch)
	}
	
	if !*sortReports {
		for range flag.Args() {
			fmt.Println((<-ch).Report())
		}
	} else {
		col := Crc32ReportCollection( make([]*Crc32Report, len(flag.Args())) )
		for i := range flag.Args() {
			col[i] = <-ch
		}
		
		sort.Sort(col)
		
		for _, r := range col {
			fmt.Println(r.Report())
		}
	}
}

// Other functions

func crc32File(filename string, number int, ch chan *Crc32Report) {
	report := &Crc32Report{Filename: filename, Number: number}
	
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
