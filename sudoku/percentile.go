package sudoku

import (
	"bytes"
	"math"
	"sort"
	"strconv"
)

type Percentile struct {
	stats *bytes.Buffer
}

func NewPercentile(latencies []int, inlfy int) *Percentile {
	p := &Percentile{stats: new(bytes.Buffer)}
	l := len(latencies)

	p.stats.WriteString("recv " + strconv.Itoa(l) + " in-fly " + strconv.Itoa(inlfy))

	if l > 0 {
		sort.Ints(latencies)
		min := latencies[0]
		max := latencies[l-1]
		sum := p.sum(latencies)
		mean := sum / l
		median := p.getPercentile(latencies, 50)
		p90 := p.getPercentile(latencies, 90)
		p99 := p.getPercentile(latencies, 99)

		p.stats.WriteString(" min " + strconv.Itoa(min))
		p.stats.WriteString(" mam " + strconv.Itoa(max))
		p.stats.WriteString(" avg " + strconv.Itoa(mean))
		p.stats.WriteString(" median " + strconv.Itoa(median))
		p.stats.WriteString(" p90 " + strconv.Itoa(p90))
		p.stats.WriteString(" p99 " + strconv.Itoa(p99))
	}

	return p
}

func (p *Percentile) Report() string {
	return p.stats.String()
}

func (p *Percentile) sum(numbers []int) (s int) {
	for _, num := range numbers {
		s += num
	}
	return
}

func (p *Percentile) getPercentile(latencies []int, percent int) int {
	var idx float64
	if percent > 0 {
		idx = math.Ceil(float64(len(latencies))*float64(percent)/100) - 1
	}

	return latencies[int(idx)]
}
