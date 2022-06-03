# go-weight-shuffling
[![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/k8gb-io/go-weight-shuffling.svg)](https://pkg.go.dev/github.com/k8gb-io/go-weight-shuffling?branch=main)
![Build Status](https://github.com/k8gb-io/go-weight-shuffling/actions/workflows/test.yaml/badge.svg?branch=main)
![Linter](https://github.com/k8gb-io/go-weight-shuffling/actions/workflows/lint.yaml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/k8gb-io/go-weight-shuffling)](https://goreportcard.com/report/github.com/k8gb-io/go-weight-shuffling?branch=main)


This library provides a Weight Shuffling support function which achieves both performance and simplicity. The functionality
is suitable for weight round-robin in a distributed environments.

For detailed information about the concept, you should take a look at the following resources:
- [CDF x PDF](https://www.statology.org/cdf-vs-pdf/)
- [What is Weight Round Robin?](https://www.educative.io/edpresso/what-is-the-weighted-round-robin-load-balancing-technique)

## Table of Content
- [Install](#install)
- [Pick() Usage](#pick-usage)
- [PickVector() Usage](#pickvector-usage)
- [Examples](#examples)

## Install
With a correctly configured Go environment:
```
go get github.com/k8gb-io/go-weight-shuffling
```

## Pick() Usage

```go
// pdf requires to be 100% in total.  
pdf := []int{30, 40, 20, 10}
// handle error in real code
ws, _ := NewWS(pdf)
// the index is selected from the probability determined by the pdf 
index := ws.Pick()
```
If you don't remember [the probability](https://www.statology.org/cdf-vs-pdf/) right now, that's okay.
PDF is a simple slice that contains percentages. Depending on how the percentages are divided, the function will
return an index. For example, for PDF={5,90,5}, the function will return 1 in about 90 out of 100 cases,
while it will return 0 or 2 in about 10 out of 100 cases.

**The only condition is that the sum of all values in the PDF is always equal to 100!**

## PickVector() Usage

```go
// pdf requires to be 100% in total.  
pdf := []int{30, 40, 20, 10}
// handle error in real code
ws, _ := NewWS(pdf)
// the result will be slices of the index, which will be "probably" sorted by probability
indexes := wrr.PickVector()
```

A bit more complex case is when you need to shuffle the indexes in the array to match the PDF instead of one element.
The PDF again contains the same percentage distribution, but we want the slice to contain not one index, but the whole
vector. For example, for `PDF={30,40,20,10}` the result will be like this:

```
[2,1,3,0]
[0,1,3,2]
[0,1,2,3]
...
```
the function returns an index slice such that index0 will be represented in the zero position in about 30% of cases,
index1 will be in the first position in about 40% of cases, etc.

## Examples
This library is ideal for Weight RoundRobin. Imagine you need to balance these addresses (can be applied to whole groups
of addresses):
```shell
# dig wrr.cloud.example.com +short
10.1.0.1
10.0.0.1
10.2.0.1
10.3.0.1
```

We want to shuffle the addresses in PDF order `[30 40 20 10]`: The item with the highest probability (index 01 = 40%) will
occur more often at the 01 position that has the highest probability in the PDF.

```txt
 IP:  [10.0.0.1, 10.1.0.1, 10.2.0.1, 10.3.0.1]
 PDF: [30 40 20 10]
    -----------------
 0. [289 401 200 110] 
 1. [298 315 258 129] 
 2. [291 216 307 186] 
 3. [122 68 235 575] 
```

The example matrix was created by 1000x hitting the list of IP addresses with help of WRR.
If we map the indexes to a slice with IP addresses (or groups of IP addresses) the IP at
zero index (`10.0.0.1`) is used 289x on the first position returned by DNS server (e.g: `[10.0.0.1, 10.1.0.1, 10.2.0.1, 10.3.0.1]`).
However, 298x used on the second position (e.g: `[10.1.0.1, 10.0.0.1, 10.3.0.1, 10.2.0.1]`).

The address (`10.3.0.1`) has only 10% probability of to be chosen. It occurs only 110x (cca 10%) on the zero position
while 575x on the last position.

The index was calculated 1000 times. When you sum individual columns or rows, the result is always 1000x so everything
is  mathematically OK. Let me add a few more examples.

#### 100%
Let's say we set `pdf={0,0,100,0}` (the sum of PDF must always be 100!).
The `PickVecor` function will always generate this indexes: `[2 1 0 3]`, so for our IP addresses they will always be
sorted like this: `[10.2.0.1,10.1.0.1,10.0.0.1,10.3.0.1]`.  This is the result matrix
```
    [10.0.0.1],[10.1.0.1],[10.2.0.1],[10.3.0.1]
    [0 0 100 0]
    -----------------
 0. [0 0 100 0] 
 1. [0 100 0 0] 
 2. [100 0 0 0] 
 3. [0 0 0 100] 
```

#### 50% / 50%
the last case is a bit redundant, although very explanatory. Let's say we have `pdf={50,50}`.
The generated wil look like following:
```
[0 1]
[1 0]
[0 1]
[1 0]
[0 1]
[0 1]
[1 0]
[1 0]
...

    [10.0.0.1],[10.1.0.1]
    [50 50]
    -----------------
 0. [511 489] 
 1. [489 511] 
```
The address `10.0.0.1` occurred 511 times in `1000` hits at index zero , while 489 times at index 1.