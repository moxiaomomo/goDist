package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	wosname = flag.String("wosname", "c301", "download wosname")
	execIdx = flag.String("execidx", "0", "execute index")
	execCnt = flag.String("execcnt", "1", "execute count")
)

type DownloadResult struct {
	oid      string
	suc      bool
	size     int64
	st       int64
	et       int64
	duration int64
}

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

func Download(wosname string, node string, oid string) DownloadResult {
	nt := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s]To download %s\n", nt, oid)

	url := fmt.Sprintf("http://%s/objects/%s", node, oid)
	fpath := fmt.Sprintf("/www/files/download/%s_%s", wosname, oid)
	os.Remove(fpath)
	newFile, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err.Error())
		return DownloadResult{
			oid: oid,
			suc: false,
		}
	}
	defer newFile.Close()

	st := time.Now().UnixNano()

	client := http.Client{Timeout: 900 * time.Second}
	resp, err := client.Get(url)
	defer resp.Body.Close()

	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	et := time.Now().UnixNano()
	return DownloadResult{
		oid:      oid,
		suc:      true,
		size:     resp.ContentLength,
		st:       st,
		et:       et,
		duration: et - st,
	}
}

func main() {
	flag.Parse()

	startIdx, err := strconv.Atoi(*execIdx)
	if err != nil {
		return
	}
	execNum, err := strconv.Atoi(*execCnt)
	if err != nil {
		return
	}

	nodelist := ReadLines(fmt.Sprintf("%s_node.txt", *wosname))
	if len(nodelist) == 0 {
		return
	}

	oidlist := ReadLines(fmt.Sprintf("%s_oid.txt", *wosname))
	if len(oidlist) == 0 {
		return
	}

	ch := make(chan DownloadResult)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for idx := startIdx; idx < len(oidlist) && idx-startIdx < execNum; idx++ {
		node := nodelist[r.Intn(len(nodelist))]
		go func(node, oid string) {
			ch <- Download(*wosname, node, oid)
		}(node, oidlist[idx])
	}

	timeout := time.After(900 * time.Second)
	for idx := startIdx; idx < len(oidlist) && idx-startIdx < execNum; idx++ {
		select {
		case res := <-ch:
			nt := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("[%s]Finish download %s %d %d\n", nt, res.oid, res.size, res.duration)
		case <-timeout:
			fmt.Println("Timeout...")
			break
		}
	}
}
