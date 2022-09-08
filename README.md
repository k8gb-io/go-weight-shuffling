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
- [Introduction](#introduction)
- [Pick() Usage](#pick-usage)
- [PickVector() Usage](#pickvector-usage)
- [Examples](#examples)

## Install
With a correctly configured Go environment:
```
go get github.com/k8gb-io/go-weight-shuffling
```

## Introduction
Use this package in case you need to select elements or balance the load with certain probability.Basically the only 
thing you need to understand is PDF distribution. The PDF - or weights - determines with what probability the individual 
elements in the array will be selected.

For example, I have a slice with these IP addresses:
```go
ips := []string{"10.1.0.1","10.2.0.1","10.3.0.1","10.4.0.1"}
```
I would like to pick the indexes of these addresses with a certain probability, therefore
I'm defining `weights := {3,4,2,1}` to determine such probabilities. The chance of selecting the first index 
(address `10.1.0.1`) is 30%, the chance of selecting the second index (`10.2.0.1`) is 40%, etc. 
So the final probability is: `{0:30%, 1:40%, 2:20%, 3:10%}`

For some reason I decide I don't want to select `10.3.0.1` anymore. Therefore, I set the weight in the 
pdf to 0 for the element I no longer want to use `weights := {3,4,0,1}`. The probabilities are automatically 
recalculated and the returned indexes will be returned with probabilities `{0:38%,1:50%, 2:0%, 3:12%}`
value `2` (index of `10.3.0.1`) is dropped.

The usage is simple, the package defines two methods: 
- `Pick()` returning one index
- `PickVector()` returning all indexes in such order that the ones with the highest weight appears at 
the beginning of the returned slice, while the ones with the lowest weight appear at the end.

![](https://user-images.githubusercontent.com/7195836/189152064-6e105001-75c1-4381-9089-5d1be556c324.png)


## Pick() Usage
Pick returns single index with probability given by weights.
```go
weights := []uint32{30, 40, 20, 10}
// handle error in real code
w := gows.NewWS(weights)
// the index is selected from the probability determined by the weight 
index,_ := w.Pick()
```
If the sum of the weights is equal to zero the function generates an error (there is no index to choose from if everything is 0)

## PickVector(Settings) Usage
PickVector returns slice shuffled by weights distribution. returning all indexes in such order that the ones with the 
highest weight appears at the beginning of the returned slice, while the ones with the lowest weight appear at the end.
```go
weights := []int{30, 40, 20, 10}
// handle error in real code
ws := gows.NewWS(weights)
// the result will be slices of the index, which will be "probably" sorted by probability
indexes := wrr.PickVector(gows.IncludeZeroWeights)
```

For example: `weights={30,40,20,10}` will produce such results:
```
[1,2,3,0]
[0,1,3,2]
[0,1,2,3]
[1,0,2,3]
[1,3,0,2]
[0,3,2,1]
[1,0,2,3]
[2,1,0,3]
[3,0,1,2]
...
```
The function returns an index slice such that index 0 will be represented in the zero position in about 30% of cases,
index 1 will be in the first position in about 40% of cases, etc. Similarly, there are heavier weights in the second position. 
The last position belongs mostly to the low weights. 

### Settings argument
The Settings argument defines how the PickVector function will return indexes. Imagine you have 
a weights for three different parts and you set one of them to 0 (just turn it off, because the 
probability of this index will be 0). The solution is not universal, each use-case requires 
different behavior. Currently we define two versions of the behavior.

- `IncludeZeroWeights` keeps indexes for zero weight; e.g: for `weights=[0,50,50,0,0,0]` returns only `[1,2,0,3,4,5]` or `[2,1,0,3,4,5]`
- `DropZeroWeights` filter indexes for zero weight; e.g: for `weights=[0,50,50,0,0,0]` returns only `[1,2]` or `[2,1]`

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

We want to shuffle the addresses for weights `[30 40 20 10]`: The item with the highest probability (index 1 = 40%) will
occur more often at the 0 position.

```txt
 IP:      [10.0.0.1, 10.1.0.1, 10.2.0.1, 10.3.0.1]
 WEIGHTS: [30 40 20 10]
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
Let's say we set `weight={0,0,1,0}`. The `PickVecor` function will always generate this indexes: `[2 1 0 3]`, 
so for our IP addresses they will always be sorted like this: `[10.2.0.1,10.1.0.1,10.0.0.1,10.3.0.1]`.  
This is the result matrix
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
the last case is a bit redundant, although very explanatory. Let's say we have `weight={1,1}`.
The generated sample will look like following:
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
The address `10.0.0.1` occurred 511 times in `1000` hits at index 0 , while 489 times at index 1.
