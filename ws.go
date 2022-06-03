package gows

/*
Copyright 2022 The k8gb Contributors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/

import (
	"fmt"
	"math/rand"
	"time"
)

// WS Weight Round Robin Alghoritm
type WS struct {
	pdf      []int
	index100 int
}

// NewWS instantiate weight round robin
func NewWS(pdf []int) (wrr *WS, err error) {
	r := 0
	max100 := -1
	for i, v := range pdf {
		if v == 100 {
			max100 = i
		}
		r += v
		if v < 0 || v > 100 {
			return wrr, fmt.Errorf("value %v out of range [0;100]", v)
		}
	}
	if r != 100 {
		return wrr, fmt.Errorf("sum of pdf elements must be equal to 100 perent")
	}
	rand.Seed(time.Now().UnixNano())
	wrr = new(WS)
	wrr.pdf = pdf
	wrr.index100 = max100
	return wrr, nil
}

// PickVector returns slice shuffled by pdf distribution.
// The item with the highest probability will occur more often
// at the position that has the highest probability in the PDF
// see README.md
func (w *WS) PickVector() (indexes []int) {
	if w.index100 != -1 {
		return w.handle100()
	}

	pdf := make([]int, len(w.pdf))
	copy(pdf, w.pdf)
	balance := 100
	for i := 0; i < len(pdf); i++ {
		cdf := w.getCDF(pdf)
		index := w.pick(cdf, balance)
		indexes = append(indexes, index)

		balance -= pdf[index]
		pdf[index] = 0

		// Summary of new pdf must be 100%. Need to add missing percentage
		for q, v := range pdf {
			if v != 0 {
				pdf[q] = v
			}
		}
	}
	return indexes
}

// Pick returns one index with probability given by pdf
// see README.md
func (w *WS) Pick() int {
	cdf := w.getCDF(w.pdf)
	return w.pick(cdf, 100)
}

// pick one index
func (w *WS) pick(cdf []int, n int) int {
	r := rand.Intn(n)
	index := 0
	for r >= cdf[index] {
		index++
	}
	return index
}

func (w *WS) getCDF(pdf []int) (cdf []int) {
	// prepare cdf
	for i := 0; i < len(pdf); i++ {
		cdf = append(cdf, 0)
	}
	cdf[0] = pdf[0]
	for i := 1; i < len(pdf); i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}
	return cdf
}

// there is no reason to calculate CDF and recompute PDF's if some field has 100%
func (w *WS) handle100() (indexes []int) {
	for i := 0; i < len(w.pdf); i++ {
		indexes = append(indexes, i)
	}
	indexes[0], indexes[w.index100] = indexes[w.index100], indexes[0]
	return indexes
}