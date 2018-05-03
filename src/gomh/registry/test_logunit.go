package main

import (
	"fmt"
	"gomh/registry/raft"
	"os"
)

func main() {
	file, err := os.OpenFile("/tmp/test0503.txt", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//	data := &raft.LogUnit{
	//		Leader:     "cdf898787",
	//		StartIndex: 9,
	//		Term:       4,
	//	}

	//	err = data.Dump(file)
	//	fmt.Println(err)
	//
	//	data = &raft.LogUnit{
	//		Leader:     "cdf898787kkk",
	//		StartIndex: 7,
	//		Term:       7,
	//	}
	//
	//	err = data.Dump(file)
	//	fmt.Println(err)

	file.Close()

	file, err = os.OpenFile("/tmp/test0503.txt", os.O_RDWR, os.ModePerm)
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
