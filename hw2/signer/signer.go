package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var md5mux = sync.Mutex{}
var counter atomic.Int32

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}

//// сюда писать код
//func SingleHash(in, out chan interface{}) {
//	defer un(trace("SingleHash"))
//	buff := make([]string, 2)
//	wg := &sync.WaitGroup{}
//	val := <-in
//
//	md5mux.Lock() // блокируем расчет
//	md5 := DataSignerMd5(strconv.Itoa(val.(int)))
//	md5mux.Unlock()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		buff[0] = DataSignerCrc32(strconv.Itoa(val.(int)))
//	}()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		buff[1] = DataSignerCrc32(md5)
//	}()
//	wg.Wait()
//	res := buff[0] + "~" + buff[1]
//	fmt.Println("SingleHash ", res)
//	out <- res
//}
//
//func MultiHash(in, out chan interface{}) {
//	defer un(trace("MultiHash"))
//	wg := &sync.WaitGroup{}
//	val := <-in
//	buff := make([]string, 6)
//	for i := 0; i <= 5; i++ {
//		wg.Add(1)
//		go func(iter int) {
//			defer wg.Done()
//			buff[iter] = DataSignerCrc32(strconv.Itoa(iter) + val.(string))
//		}(i)
//	}
//	wg.Wait() //ждем расчета хэшей для конкатенации
//	fmt.Println("MultiHash ", strings.Join(buff, ""))
//	out <- strings.Join(buff, "")
//}

// сюда писать код
func SingleHash(in, out chan interface{}) {
	defer un(trace("SingleHash"))
	buff := make([]string, 2)
	wg := &sync.WaitGroup{}
	for val := range in {
		go func() {
			md5mux.Lock() // блокируем расчет
			md5 := DataSignerMd5(strconv.Itoa(val.(int)))
			md5mux.Unlock()

			wg.Add(1)
			go func() {
				defer wg.Done()
				buff[0] = DataSignerCrc32(strconv.Itoa(val.(int)))
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				buff[1] = DataSignerCrc32(md5)
			}()
			wg.Wait()
			res := buff[0] + "~" + buff[1]
			fmt.Println("SingleHash ", res)
			out <- res
		}()
	}
}

func MultiHash(in, out chan interface{}) {
	defer un(trace("MultiHash"))
	wg := &sync.WaitGroup{}
	for val := range in {
		go func() {
			buff := make([]string, 6)
			for i := 0; i <= 5; i++ {
				wg.Add(1)
				go func(iter int) {
					defer wg.Done()
					buff[iter] = DataSignerCrc32(strconv.Itoa(iter) + val.(string))
				}(i)
			}
			wg.Wait() //ждем расчета хэшей для конкатенации
			fmt.Println("MultiHash ", strings.Join(buff, ""))
			out <- strings.Join(buff, "")
		}()

	}
}

func CombineResults(in, out chan interface{}) {
	valSlice := make([]string, 0)
	for val := range in {
		counter.Add(-1)
		valSlice = append(valSlice, val.(string))
	}
	sort.Strings(valSlice)
	fmt.Println(strings.Join(valSlice, "_"))
	out <- strings.Join(valSlice, "_")
}

func ExecutePipeline(hashSignJobs ...job) {
	chanSlice := make([]chan interface{}, 0)
	in := make(chan interface{})
	for _, j := range hashSignJobs {
		out := make(chan interface{}, 100)
		go j(in, out)
		in = out
		chanSlice = append(chanSlice, in)
	}

	//out3 := make(chan interface{}, 100)
	//out4 := make(chan interface{})
	//in := make(chan interface{})
	//go hashSignJobs[0](nil, out1)
	//go hashSignJobs[1](out1, out2)
	//go hashSignJobs[2](out2, out3)
	//go hashSignJobs[3](out3, out4)
	//val := <-out4
	//in <- val
	//go hashSignJobs[4](in, nil)
	fmt.Println("len(chanSlice[0])", len(chanSlice[0]))
	close(chanSlice[0])
	time.Sleep(3 * time.Second)
	for _, ch := range chanSlice {
		close(ch)
	}
	fmt.Println(<-chanSlice[len(chanSlice)-1])

}

func main() {
	//CGO_ENABLED = 1
	inputData := []int{0, 1}
	//inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				fmt.Println("start ", fibNum)
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}
	ExecutePipeline(hashSignJobs...)

}
