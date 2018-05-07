package raft

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	//"fmt"
	"os"
)

type LogUnitMeta struct {
	DataLength uint32
}

type LogUnit struct {
	CurLeader     string `json:"curleader"`
	CurTerm       uint64 `json:"curterm"`
	CurIndexStart uint64 `json:"curindexstart"`
	//	CurIndexEnd    uint64 `json:"curindexend"`
	LastTerm       uint64 `json:"lastterm"`
	LastIndexStart uint64 `json:"lastindexstart"`
	//	LastIndexEnd   uint64 `json:"lastindexend"`
}

func NewLogUnit(leader string, curterm, lastterm uint64,
	lastindexstart uint64, lastindexend uint64) *LogUnit {
	lu := &LogUnit{
		CurLeader:     leader,
		CurTerm:       curterm,
		CurIndexStart: lastindexend + 1,
		//		CurIndexEnd:    999,
		LastTerm:       lastterm,
		LastIndexStart: lastindexstart,
		//		LastIndexEnd:   lastindexend,
	}
	return lu
}

func (l *LogUnit) ComputeLogLength() (int64, error) {
	d, err := json.Marshal(l)
	if err != nil {
		return -1, err
	}
	data := []byte(d)
	meta := LogUnitMeta{
		DataLength: uint32(len(data)),
	}

	return int64(binary.Size(meta)) + int64(meta.DataLength), nil
}

// save data into file
func (l *LogUnit) dump(file *os.File) (indexend int64, err error) {
	n, _ := file.Seek(0, os.SEEK_END)

	w := bufio.NewWriter(file)

	d, err := json.Marshal(l)
	if err != nil {
		return -1, err
	}
	data := []byte(d)
	meta := LogUnitMeta{
		DataLength: uint32(len(data)),
	}
	err = binary.Write(w, binary.BigEndian, &meta)
	if err != nil {
		return -1, err
	}

	w.Write(data)
	w.Flush()
	return n + int64(binary.Size(meta)) + int64(meta.DataLength), nil
}

// load data from file
func (l *LogUnit) load(file *os.File, startIndex int64) (indexend int64, err error) {
	n, _ := file.Seek(startIndex, 0)
	r := bufio.NewReader(file)

	meta := &LogUnitMeta{}
	err = binary.Read(r, binary.BigEndian, meta)
	if err != nil {
		return -1, err
	}

	data := make([]byte, meta.DataLength)
	_, err = r.Read(data)
	if err != nil {
		return -1, nil
	}

	err = json.Unmarshal(data, l)
	if err != nil {
		return -1, err
	}
	return n + int64(binary.Size(meta)) + int64(meta.DataLength), nil
}
