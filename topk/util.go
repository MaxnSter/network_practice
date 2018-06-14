package topk

import "strings"

type WordCounter struct {
	Word  string
	Count int
}

type WordCounters []*WordCounter

func (wc WordCounters) Len() int { return len(wc) }
func (wc WordCounters) Less(i, j int) bool {
	//升序
	if wc[i].Count == wc[j].Count {
		return strings.Compare(wc[i].Word, wc[j].Word) < 0
	}

	return wc[i].Count > wc[j].Count
}
func (wc WordCounters) Swap(i, j int) { wc[i], wc[j] = wc[j], wc[i] }
