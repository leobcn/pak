// Package pak reads and writes chromium .pak resource files
package pak

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
)

type PakFile struct {
	Version   uint32
	Encoding  uint8
	Resourses map[uint16][]byte // maps resource id -> resource data
}

const (
	EncodingBinary = iota
	EncodingUTF8
	EncodingUTF16
)

type resourceInfo struct {
	id     uint16
	offset uint32
}

// Reads pak struct from io.Reader
func Read(r io.Reader) (*PakFile, error) {
	var err error

	// Read header:
	// 4 byte version number
	// 4 byte number of resources
	// 1 byte encoding

	const headerLength = 4 + 4 + 1
	var version uint32
	var numberOfResources uint32
	var encoding uint8

	err = binary.Read(r, binary.LittleEndian, &version)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &numberOfResources)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &encoding)
	if err != nil {
		return nil, err
	}

	pak := &PakFile{
		Version:   version,
		Encoding:  encoding,
		Resourses: make(map[uint16][]byte),
	}

	// For each resource read info:
	// 2 byte resource id
	// 4 byte resource offset in file
	// Extra resource entry at the end with ID 0 giving the end of the last resource

	resInfos := make([]resourceInfo, numberOfResources+1, numberOfResources+1)

	var i uint32

	for i = 0; i < numberOfResources+1; i++ {
		ri := resourceInfo{}

		err = binary.Read(r, binary.LittleEndian, &ri.id)
		if err != nil {
			return nil, err
		}

		err = binary.Read(r, binary.LittleEndian, &ri.offset)
		if err != nil {
			return nil, err
		}

		resInfos[i] = ri
	}

	if resInfos[numberOfResources].id != 0 {
		return nil, fmt.Errorf("error reading resources: last id != 0")
	}

	// Read resources
	for i = 0; i < numberOfResources; i++ {
		resId := resInfos[i].id
		resLength := resInfos[i+1].offset - resInfos[i].offset
		resData := make([]byte, resLength, resLength)

		n, err := r.Read(resData)
		if err != nil {
			return nil, err
		}
		if uint32(n) != resLength {
			return nil, fmt.Errorf("error reading resource id=%d", resId)
		}

		pak.Resourses[resId] = resData
	}

	return pak, nil
}

// Reads pak struct from file
func ReadFile(name string) (*PakFile, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

// Writes pak struct to io.Writer
func Write(w io.Writer, p *PakFile) error {
	var err error

	if p == nil {
		return fmt.Errorf("error writing pak: p == nil")
	}

	// Write header:
	// 4 byte version number
	// 4 byte number of resources
	// 1 byte encoding

	const headerLength = 4 + 4 + 1
	numberOfResources := uint32(len(p.Resourses))

	err = binary.Write(w, binary.LittleEndian, p.Version)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, numberOfResources)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, p.Encoding)
	if err != nil {
		return err
	}

	// For each resource write info:
	// 2 byte resource id
	// 4 byte resource offset in file

	indexLength := (2 + 4) * (numberOfResources + 1) // count one extra entry for last offset
	curOffset := headerLength + indexLength          // start offset for resource data

	// Sort resource ids
	ids := make([]int, 0, numberOfResources)
	for resId, _ := range p.Resourses {
		ids = append(ids, int(resId)) // use int type for easy sorting
	}
	sort.Ints(ids)

	for _, id := range ids {
		resId := uint16(id)
		err = binary.Write(w, binary.LittleEndian, resId)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.LittleEndian, curOffset)
		if err != nil {
			return err
		}

		curOffset += uint32(len(p.Resourses[resId]))
	}

	// Extra resource entry at the end with ID 0 giving the end of the last resource
	err = binary.Write(w, binary.LittleEndian, uint16(0))
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, curOffset)
	if err != nil {
		return err
	}

	// Write resources
	for _, id := range ids {
		resId := uint16(id)
		resData := p.Resourses[resId]
		resLength := len(resData)

		n, err := w.Write(resData)
		if err != nil {
			return err
		}
		if n != resLength {
			return fmt.Errorf("error writing resource id=%d", resId)
		}
	}

	return nil
}

// Writes pak struct to file
func WriteFile(name string, p *PakFile) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return Write(f, p)
}
