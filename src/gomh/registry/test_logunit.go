package main

import (
	"encoding/json"
	"fmt"
	"gomh/registry/raft"
	"os"
)

func main() {
	file, err := os.OpenFile("/data/apps/goDist/src/gomh/registry/raft/internlog/raft-log",
		os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	data := raft.NewLogUnit("cdf898787", 1, 2, 3, 4)
	fmt.Println(data)
	b, err := json.Marshal(data)
	fmt.Println(b)
	_, err = data.Dump(file)
	fmt.Println(err)

	data = raft.NewLogUnit("kokouiyi", 5, 6, 7, 8)
	_, err = data.Dump(file)
	fmt.Println(err)

	file.Close()

	file, err = os.OpenFile("/data/apps/goDist/src/gomh/registry/raft/internlog/raft-log",
		os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	data2 := &raft.LogUnit{}
	n, err := data2.Load(file, 0)
	fmt.Printf("%v %v\n", data2, err)

	data2 = &raft.LogUnit{}
	n, err = data2.Load(file, n)
	fmt.Printf("%v %v\n", data2, err)

	data2 = &raft.LogUnit{}
	n, err = data2.Load(file, n)
	fmt.Printf("%v %v\n", data2, err)

	file.Close()
}
