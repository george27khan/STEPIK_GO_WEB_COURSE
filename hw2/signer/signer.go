package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var md5mux = sync.Mutex{}

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}

// сюда писать код
func SingleHash(in, out chan interface{}) {
	defer un(trace("SingleHash"))
	buff := make([]string, 2)
	wg := &sync.WaitGroup{}
	val := <-in

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
	fmt.Println(res)
	out <- res
}

func MultiHash(in, out chan interface{}) {
	defer un(trace("MultiHash"))
	wg := &sync.WaitGroup{}
	val := <-in
	buff := make([]string, 6)
	for i := 0; i <= 5; i++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			buff[iter] = DataSignerCrc32(strconv.Itoa(iter) + val.(string))
		}(i)
	}
	wg.Wait() //ждем расчета хэшей для конкатенации
	fmt.Println(strings.Join(buff, ""))
	out <- strings.Join(buff, "")
}

func CombineResults(in, out chan interface{}) {
	valSlice := make([]string, 0)
	for val := range in {
		valSlice = append(valSlice, val.(string))
	}
	sort.Strings(valSlice)
	fmt.Println(strings.Join(valSlice, "_"))
	out <- strings.Join(valSlice, "_")
}

func ExecutePipeline(hashSignJobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	//out2 := make(chan interface{})
	wg := &sync.WaitGroup{}
	out2 := make(chan interface{}, 100)
	out3 := make(chan interface{}, 100)
	//out3 := make(chan interface{})
	//out4 := make(chan interface{})
	go hashSignJobs[0](in, out)
	go hashSignJobs[3](out2, out3)
	//wg.Add(1)
	go func() {
		fmt.Println("dep-1")
		for val := range out {
			wg.Add(1)
			fmt.Println("dep-2")
			go func(val interface{}) {
				defer wg.Done()
				in := make(chan interface{}, 100)
				out1 := make(chan interface{}, 100)
				fmt.Println("dep-3")
				in <- val
				go hashSignJobs[1](in, out1)
				go hashSignJobs[2](out1, out2)
			}(val)
		}
	}()
	//time.Sleep(time.Second * 50)
	wg.Wait()

	close(out)
	//go hashSignJobs[1](out1, out2)
	//go hashSignJobs[2](out2, out3)
	////go hashSignJobs[3](out3, out4)
	//fmt.Println(<-out3)
	//fmt.Println(<-out3)
	//close(out1)

	//time.Sleep(time.Second)

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

	//in := make(chan int, 15)

	//out := make(chan int)
	//wg := &sync.WaitGroup{}
	////заполняем канал значениями
	//go func() {
	//	for i := 0; i < 5; i++ {
	//		fmt.Println("in chan <- ", i)
	//		in <- i
	//	}
	//}()
	//go func() {
	//	for i := 0; i < 5; i++ {
	//		fmt.Println("out chan <- ", <-in)
	//	}
	//}()
	//time.Sleep(time.Second * 4)
	//go func(wg *sync.WaitGroup) {
	//	fmt.Println("1 ", <-in)
	//}(wg)
	//go func(wg *sync.WaitGroup) {
	//	fmt.Println("2 ", <-in)
	//}(wg)
	//go func(wg *sync.WaitGroup) {
	//	fmt.Println("2 ", <-in)
	//}(wg)
	//time.Sleep(time.Second * 50)
	//fmt.Println("END")
	//fmt.Println(wg)
	//wg.Wait() //жду когда счетчик обнулится
	//close(in)
}
