package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var md5mux = sync.Mutex{}
var counter atomic.Int32

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: //slog.LevelDebug,
		slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler)) // Устанавливаем глобальный логгер
}

func trace(s string) (string, time.Time) {
	slog.Debug("START", "func", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	slog.Debug("END", "func", s, "ElapsedTime in seconds", endTime.Sub(startTime))
}

// сюда писать код
func SingleHash(in, out chan interface{}) {
	defer un(trace("SingleHash"))
	buff := make([]string, 2)
	wg := &sync.WaitGroup{}
	slog.Debug("val calc", "in", in)
	for val := range in {
		wg.Add(1)
		slog.Debug("val calc", "val", val)
		val := strconv.Itoa(val.(int))
		go func() {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}
			md5mux.Lock() // блокируем расчет
			md5 := DataSignerMd5(val)
			md5mux.Unlock()

			wg1.Add(1)
			go func() {
				defer wg1.Done()
				buff[0] = DataSignerCrc32(val)
			}()

			wg1.Add(1)
			go func() {
				defer wg1.Done()
				buff[1] = DataSignerCrc32(md5)
			}()
			wg1.Wait()
			res := buff[0] + "~" + buff[1]
			slog.Debug("SingleHash calc", "in", val, "res", res)
			out <- res
		}()
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	defer un(trace("MultiHash"))
	wg := &sync.WaitGroup{}
	for val := range in {
		wg.Add(1)
		val := val.(string)
		go func() {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}
			buff := make([]string, 6)
			for i := 0; i <= 5; i++ {
				wg1.Add(1)
				go func(iter int) {
					defer wg1.Done()
					buff[iter] = DataSignerCrc32(strconv.Itoa(iter) + val)
				}(i)
			}
			wg1.Wait() //ждем расчета хэшей для конкатенации
			slog.Debug("MultiHash calc", "in", val, "res", strings.Join(buff, ""))
			out <- strings.Join(buff, "")
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	valSlice := make([]string, 0)
	for val := range in {
		counter.Add(-1)
		valSlice = append(valSlice, val.(string))
	}
	sort.Strings(valSlice)
	slog.Debug("CombineResults calc", "res", strings.Join(valSlice, "_"))
	out <- strings.Join(valSlice, "_")
}

func ExecutePipeline(hashSignJobs ...job) {
	chanSlice := make([]chan interface{}, 0)
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	chanSlice = append(chanSlice, in)
	for i, j := range hashSignJobs {
		out := make(chan interface{}, 100)
		chanSlice = append(chanSlice, out)
		wg.Add(1)
		go func(i int, j job) {
			defer wg.Done()
			j(chanSlice[i], chanSlice[i+1])
			close(chanSlice[i+1])
		}(i, j)
	}
	wg.Wait()

}

func main() {
	//CGO_ENABLED = 1
	inputData := []int{0, 1}
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
