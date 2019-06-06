package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

func CombineResults(in, out chan interface{}) {
	buf := make([]string, 0)
	for str := range in {
		buf = append(buf, str.(string))
	}
	sort.Strings(buf)
	out <- strings.Join(buf, "_")
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)
		go func(i interface{}) {
			defer wg.Done()
			outline := ""
			buf := make(chan string, 6)
			data := fmt.Sprint(i)
			for i := 0; i < 6; i++ {
				go func(i int) {
					buf <- DataSignerCrc32(fmt.Sprint(i) + data)
				}(i)
				time.Sleep(time.Millisecond)
			}
			for i := 0; i < 6; i++ {
				tmp := <-buf
				outline = outline + tmp
			}
			close(buf)
			out <- outline
		}(i)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for i := range in {
		wg.Add(1)
		go func(v interface{}) {
			defer wg.Done()
			data1 := fmt.Sprint(v)
			buf := make(chan interface{}, 2)
			go func() {
				buf <- DataSignerCrc32(data1)
			}()
			mu.Lock()
			data2 := DataSignerMd5(data1)
			mu.Unlock()
			go func() {
				buf <- DataSignerCrc32(data2)
			}()
			outLine := fmt.Sprint(<-buf) + "~" + fmt.Sprint(<-buf)
			out <- outLine
		}(i)
	}
	wg.Wait()
}

func ExecutePipeline(freeFlowJobs ...job) {
	channels := make([]chan interface{}, len(freeFlowJobs)+1)
	wg := &sync.WaitGroup{}
	for key, j := range freeFlowJobs {
		channels[key+1] = make(chan interface{}, 100)
		wg.Add(1)
		go func(i int, f func(in, out chan interface{})) {
			f(channels[i], channels[i+1])
			close(channels[i+1])
			wg.Done()
		}(key, j)
	}
	wg.Wait()
}
