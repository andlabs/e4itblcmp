// 2 january 2015
package main

import (
	"fmt"
	"os"
	"encoding/binary"
	"reflect"
	"flag"
	"strings"
)

var (
	notime = flag.Bool("notime", false, "omit times")
)

type Inode struct {
	Mode			uint16
	UID				uint16
	SizeLow32		uint32
	ATime			uint32
	CTime			uint32
	MTime			uint32
	DTime			uint32
	GID				uint16
	LinkCount			uint16
	BlockCountLow32	uint32
	Flags			uint32
	OSD1			uint32
	BlockMap			[60]byte
	VersionLow32		uint32
	ACLBlockLow32	uint32
	SizeHigh32		uint32
	FragmentAddress	uint32
	OSD2			[12]byte
	ExtraInodeSize		uint16
	ChecksumHigh16	uint16
	CTimeExtra		uint32
	MTimeExtra		uint32
	ATimeExtra		uint32
	CreationTime		uint32
	CreationTimeExtra	uint32
	VersionHigh32		uint32
	Else				[0x100 - 0x9C]byte
}

var dups [4096 / 0x100]struct {
	Mode			[]uint16
	UID				[]uint16
	SizeLow32		[]uint32
	ATime			[]uint32
	CTime			[]uint32
	MTime			[]uint32
	DTime			[]uint32
	GID				[]uint16
	LinkCount			[]uint16
	BlockCountLow32	[]uint32
	Flags			[]uint32
	OSD1			[]uint32
	BlockMap			[][60]byte
	VersionLow32		[]uint32
	ACLBlockLow32	[]uint32
	SizeHigh32		[]uint32
	FragmentAddress	[]uint32
	OSD2			[][12]byte
	ExtraInodeSize		[]uint16
	ChecksumHigh16	[]uint16
	CTimeExtra		[]uint32
	MTimeExtra		[]uint32
	ATimeExtra		[]uint32
	CreationTime		[]uint32
	CreationTimeExtra	[]uint32
	VersionHigh32		[]uint32
	Else				[][0x100 - 0x9C]byte
}

func tallyone(vi reflect.Value, vn reflect.Value, i int) {
	needle := vi.Field(i)
	haystack := vn.Elem().Field(i)
	for j := 0; j < haystack.Len(); j++ {
		if reflect.DeepEqual(haystack.Index(j).Interface(), needle.Interface()) {
			return
		}
	}
	haystack = reflect.Append(haystack, needle)
	vn.Elem().Field(i).Set(haystack)
}

func tally(n int, inode Inode) {
	ty := reflect.TypeOf(inode)
	vi := reflect.ValueOf(inode)
	vn := reflect.ValueOf(&dups[n])
	for i := 0; i < ty.NumField(); i++ {
		tallyone(vi, vn, i)
	}
}

func scan(filename string) {
	var inode Inode

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for n := 0; n < len(dups); n++ {
		err = binary.Read(f, binary.LittleEndian, &inode)
		if err != nil {
			panic(err)
		}
		tally(n, inode)
	}
}

func report() {
	for i := 0; i < len(dups); i++ {
		fmt.Printf("entry %d:\n", i)
		s := reflect.ValueOf(dups[i])
		ty := reflect.TypeOf(dups[i])
		for j := 0; j < ty.NumField(); j++ {
			e := s.Field(j)
			if *notime && strings.Contains(ty.Field(j).Name, "Time") {
				continue
			}
			if e.Len() > 1 {
				fmt.Printf("%s %v\n", ty.Field(j).Name, e.Interface())
			}
		}
		fmt.Printf("\n")
	}
}

func main() {
	// TODO set flag.Usage
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		scan(flag.Arg(i))
	}
	report()
}
