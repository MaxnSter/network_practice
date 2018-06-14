package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
	"github.com/MaxnSter/network_practice/topk"
)

const (
	// 限制每次最多读入100w个单词
	maxSize = 1000 * 10000 * 4
)

type sharder struct {
	buckets []io.ReadWriteCloser
}

func NewSharder(bucket int) *sharder {
	s := &sharder{}
	s.buckets = make([]io.ReadWriteCloser, bucket)
	for i := 0; i < bucket; i++ {
		filename := fmt.Sprintf("shard-%05d-of-%05d", i, bucket)
		var err error
		s.buckets[i], err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
	}
	return s
}

func (s *sharder) output(word string, count int) {
	// 根据hash值写入不同的buckets
	line := fmt.Sprintf("%s\t%d\n", word, count)
	idx := crc32.ChecksumIEEE(util.StringToBytes(line)) % uint32(len(s.buckets))
	s.buckets[idx].(*os.File).WriteString(line)
}

type source struct {
	word  string
	count int
	idx   int

	scanner *bufio.Scanner
}

func NewSource(reader io.Reader) *source {
	s := &source{}
	s.scanner = bufio.NewScanner(reader)
	return s
}

func (s *source) next() bool {
	if !s.scanner.Scan() {
		return false
	} else {
		line := s.scanner.Text()
		tabIdx := strings.Index(line, "\t")
		if tabIdx == -1 {
			logger.Fatalf("tab not found, line:%s\n", line)
		}

		// 写入格式:word\t个数\n
		var err error
		s.word = line[:tabIdx]
		s.count, err = strconv.Atoi(line[tabIdx+1:])
		if err != nil {
			logger.Fatalf("atoi error at line:%s, error:%s", line, err)
		}

		return true
	}
}

func (s *source) output(writer io.Writer) {
	writer.Write([]byte(fmt.Sprintf("%d\t%s\n", s.count, s.word)))
}

type sourceHeap []*source

func (sh sourceHeap) Len() int { return len(sh) }
func (sh sourceHeap) Less(i, j int) bool {
	if sh[i].count == sh[j].count {
		return strings.Compare(sh[i].word, sh[j].word) < 0
	}
	return sh[i].count > sh[j].count
}
func (sh sourceHeap) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
	sh[i].idx = i
	sh[j].idx = j
}
func (sh *sourceHeap) Push(x interface{}) {
	n := len(*sh)
	item := x.(*source)
	item.idx = n
	*sh = append(*sh, item)
}

func (sh *sourceHeap) Pop() interface{} {
	old := *sh
	n := len(old)
	item := old[n-1]
	item.idx = -1
	*sh = old[:n-1]
	return item
}

func shard(buckets int, filenames ...string) {
	s := NewSharder(buckets)
	for _, filename := range filenames {
		logger.Infof("processing input file %s\n", filename)
		f, err := os.Open(filename)
		if err != nil {
			logger.Errorf("open file %s error:%s\n", filename, err)
			continue
		}

		m := map[string]int{}
		scan := bufio.NewScanner(f)
		scan.Split(bufio.ScanWords)
		for scan.Scan() {
			m[scan.Text()]++

			// 达到上限,写入文件
			//FIXME 次数可能有问题
			if len(m) > maxSize {
				for word, count := range m {
					s.output(word, count)
				}

				// 让gc清理内存
				m = map[string]int{}
			}
		}

		for word, count := range m {
			s.output(word, count)
		}

		f.Close()
	}

	logger.Infoln("shuffling done")
}

func readShard(idx, buckets int) map[string]int {
	wordCounts := map[string]int{}

	// 打开index对应的文件
	filename := fmt.Sprintf("shard-%05d-of-%05d", idx, buckets)
	logger.Infoln("reading ", filename)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// 读取每一行
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		tabIdx := strings.Index(line, "\t")
		if tabIdx == -1 {
			logger.Fatalf("tab not found, line:%s\n", line)
		}

		// 写入格式:word\t个数\n
		word := line[:tabIdx]
		count, err := strconv.Atoi(line[tabIdx+1:])
		if err != nil {
			logger.Fatalf("atoi error at line:%s, error:%s", line, err)
		}

		// 记录单词对应的个数,单词可能会在同一个文件中出现多次
		// 但只会在同一个文件中
		wordCounts[word] += count
	}

	syscall.Unlink(filename)
	return wordCounts
}

func sortShards(buckets int) {
	for i := 0; i < buckets; i++ {

		// 读取一个buckets文件
		wordCounts := topk.WordCounters{}
		m := readShard(i, buckets)
		for word, count := range m {
			wordCounts = append(wordCounts, &topk.WordCounter{Word: word, Count: count})
		}

		// 得到一个buckets中所有的单词信息,排序
		sort.Sort(wordCounts)

		// 排序后结果写入文件中
		filename := fmt.Sprintf("count-%05d-of-%05d", i, buckets)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		for _, wc := range wordCounts {
			f.WriteString(fmt.Sprintf("%s\t%d\n", wc.Word, wc.Count))
		}
		f.Close()

	}

	logger.Infoln("reducing done")
}

func merge(buckets int) {
	var srcs sourceHeap = make([]*source, 0)
	for i := 0; i < buckets; i++ {
		filename := fmt.Sprintf("count-%05d-of-%05d", i, buckets)
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		//读入buckets对应文件
		src := NewSource(f)
		if src.next() {
			srcs = append(srcs, src)
		}

		//之前生成的文件可以unlink了
		syscall.Unlink(filename)
	}

	// 最终结果汇总
	fout, err := os.OpenFile("output", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	// 开始N路归并
	heap.Init(&srcs)
	for len(srcs) > 0 {
		//弹出最大值并记录
		maxSource := heap.Pop(&srcs).(*source)
		maxSource.output(fout)

		// 重复上述步骤知道所有文件都已经读完
		if maxSource.next() {
			heap.Push(&srcs, maxSource)
		}
	}

	logger.Infoln("merging done")
}

func main() {
	logger.Infoln("pid = ", os.Getpid())
	shard(20, os.Args[1:]...)
	sortShards(20)
	merge(20)
}
