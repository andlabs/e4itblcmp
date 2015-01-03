// 2 january 2015
// based on diff.go
package main

import (
	"fmt"
	"os"
	"encoding/binary"
	"flag"
	"time"
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
	ATime			time.Time
	CTime			time.Time
	MTime			time.Time
	DTime			time.Time
	CreationTime		time.Time
	CTimeLatest		string
	MTimeLatest		string
	ATimeLatest		string
	DTimeLatest		string
	CreationTimeLatest	string
	CTimeNewer		bool
	MTimeNewer		bool
	ATimeNewer		bool
	DTimeNewer		bool
	CreationTimeNewer	bool
}

func totime(base uint32, extra uint32) time.Time {
	b := int64(int32(base))
	epochbits := extra & 3
	nano := extra >> 2
	offset := int64(epochbits) << 32
	b += offset
	return time.Unix(offset, int64(nano))
}

func tally(filename string, n int, inode Inode) {
	ct := totime(inode.CTime, inode.CTimeExtra)
	mt := totime(inode.MTime, inode.MTimeExtra)
	at := totime(inode.ATime, inode.ATimeExtra)
	dt := totime(inode.DTime, 0)
	crt := totime(inode.CreationTime, inode.CreationTimeExtra)
	if ct.After(dups[n].CTime) {
		dups[n].CTime = ct
		dups[n].CTimeLatest = filename
		dups[n].CTimeNewer = true
	}
	if mt.After(dups[n].MTime) {
		dups[n].MTime = mt
		dups[n].MTimeLatest = filename
		dups[n].MTimeNewer = true
	}
	if at.After(dups[n].MTime) {
		dups[n].ATime = at
		dups[n].ATimeLatest = filename
		dups[n].ATimeNewer = true
	}
	if dt.After(dups[n].DTime) {
		dups[n].DTime = dt
		dups[n].DTimeLatest = filename
		dups[n].DTimeNewer = true
	}
	if crt.After(dups[n].CreationTime) {
		dups[n].CreationTime = crt
		dups[n].CreationTimeLatest = filename
		dups[n].CreationTimeNewer = true
	}
}

func scan(filename string, first bool) {
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
		if first {
			dups[n].CTime = totime(inode.CTime, inode.CTimeExtra)
			dups[n].MTime = totime(inode.MTime, inode.MTimeExtra)
			dups[n].ATime = totime(inode.ATime, inode.ATimeExtra)
			dups[n].DTime = totime(inode.DTime, 0)
			dups[n].CreationTime = totime(inode.CreationTime, inode.CreationTimeExtra)
			dups[n].CTimeLatest = filename
			dups[n].MTimeLatest = filename
			dups[n].ATimeLatest = filename
			dups[n].DTimeLatest = filename
			dups[n].CreationTimeLatest = filename
		} else {
			tally(filename, n, inode)
		}
	}
}

func report() {
	for i := 0; i < len(dups); i++ {
		fmt.Printf("entry %d:\n", i)
		if dups[i].CTimeNewer {
			fmt.Printf("ctime: %s\n", dups[i].CTimeLatest)
		}
		if dups[i].MTimeNewer {
			fmt.Printf("mtime: %s\n", dups[i].MTimeLatest)
		}
		if dups[i].ATimeNewer {
			fmt.Printf("atime: %s\n", dups[i].ATimeLatest)
		}
		if dups[i].DTimeNewer {
			fmt.Printf("dtime: %s\n", dups[i].DTimeLatest)
		}
		if dups[i].CreationTimeNewer {
			fmt.Printf("crtime: %s\n", dups[i].CreationTimeLatest)
		}
		fmt.Printf("\n")
	}
}

func main() {
	// TODO set flag.Usage
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		scan(flag.Arg(i), i == 0)
	}
	report()
}
