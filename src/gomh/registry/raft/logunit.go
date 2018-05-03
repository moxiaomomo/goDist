package raft

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
)

type LogUnitMeta struct {
	DataLength uint32
}

type LogUnit struct {
	Leader       string
	StartIndex   uint64
	Term         uint64
	LastLogIndex uint64
	LastLogTerm  uint64
}

// save data into file
func (l *LogUnit) Dump(file *os.File) (int64, error) {
	n, _ := file.Seek(0, os.SEEK_END)
	fmt.Println(n)

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
	return n, nil
}

// load data from file
func (l *LogUnit) Load(file *os.File, startIndex int64) (int64, error) {
	n, _ := file.Seek(startIndex, 0)
	r := bufio.NewReader(file)

	meta := &LogUnitMeta{}
	binary.Read(r, binary.BigEndian, meta)

	data := make([]byte, meta.DataLength)
	r.Read(data)

	err := json.Unmarshal(data, l)
	if err != nil {
		return -1, err
	}
	return n + int64(binary.Size(meta)) + int64(meta.DataLength), nil
}
