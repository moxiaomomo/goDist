package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	wosname = flag.String("wosname", "c301", "download wosname")
)

func ReadLines(fpath string) []string {
	fd, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	var lines []string
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return lines
}

func Download(wosname string, node string, oid string) string {
	nt := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s]To download %s\n", nt, oid)

	url := fmt.Sprintf("http://%s/objects/%s", node, oid)
	fpath := fmt.Sprintf("/www/files/download/%s_%s", wosname, oid)
	newFile, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err.Error())
		return "process failed for " + oid
	}
	defer newFile.Close()

	client := http.Client{Timeout: 900 * time.Second}
	resp, err := client.Get(url)
	defer resp.Body.Close()

	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	return oid
}

func main() {
	flag.Parse()

	nodelist := ReadLines(fmt.Sprintf("%s_node.txt", *wosname))
	if len(nodelist) == 0 {
		return
	}

	oidlist := ReadLines(fmt.Sprintf("%s_oid.txt", *wosname))
	if len(oidlist) == 0 {
		return
	}

	ch := make(chan string)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, oid := range oidlist {
		node := nodelist[r.Intn(len(nodelist))]
		go func(node, oid string) {
			ch <- Download(*wosname, node, oid)
		}(node, oid)
	}

	timeout := time.After(900 * time.Second)
	for idx := 0; idx < len(oidlist); idx++ {
		select {
		case res := <-ch:
			nt := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("[%s]Finish download %s\n", nt, res)
		case <-timeout:
			fmt.Println("Timeout...")
			break
		}
	}
}
