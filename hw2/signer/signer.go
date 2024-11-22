package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

func SingleHash(in, out chan interface{}) {
	var md5mux = sync.Mutex{}
	defer un(trace("SingleHash"))
	buff := make([]string, 2) // хранить части вычислений
	wg := &sync.WaitGroup{}
	slog.Debug("val calc", "in", in)
	for val := range in {
		slog.Debug("val calc", "val", val)
		val := strconv.Itoa(val.(int)) // фиксируем значение для замыкания

		wg.Add(1)
		// параллелим расчеты для входящих значений
		go func() {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}

			md5mux.Lock() // блокируем расчет вычисления MD5, так по условию задачи
			md5 := DataSignerMd5(val)
			md5mux.Unlock()

			wg1.Add(1)
			// параллелим вычисление первой части хеша
			go func() {
				defer wg1.Done()
				buff[0] = DataSignerCrc32(val)
			}()

			wg1.Add(1)
			// параллелим вычисление первой части хеша
			go func() {
				defer wg1.Done()
				buff[1] = DataSignerCrc32(md5)
			}()
			// ждем расчетов
			wg1.Wait()
			res := buff[0] + "~" + buff[1]
			slog.Debug("SingleHash calc", "in", val, "res", res)
			out <- res
		}()
	}
	//ждем пока для всех значений рассчитаем хэш
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	defer un(trace("MultiHash"))
	wg := &sync.WaitGroup{}
	for val := range in {
		val := val.(string) // фиксируем значение для замыкания
		wg.Add(1)
		// параллелим расчеты для входящих значений
		go func() {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}
			buff := make([]string, 6)
			for i := 0; i <= 5; i++ {
				wg1.Add(1)
				// параллелим расчеты частей хеша
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
	wg.Wait() //ждем расчета хэшей для всех пришедших значений
}

func CombineResults(in, out chan interface{}) {
	valSlice := make([]string, 0)
	// не параллелим расчеты, т.к. тут идет ожидание всех значений, которые могут придти из канала до его закрытия
	for val := range in {
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
	chanSlice = append(chanSlice, in) //слайс для хранения цепочки каналов
	for i, j := range hashSignJobs {
		out := make(chan interface{}, 100)
		chanSlice = append(chanSlice, out)
		wg.Add(1)
		// запускаем параллельно цепочку джобов для обработки
		// запуск джобов идет в обертке, чтобы принимать решение о закрытии канала
		go func(i int, j job) {
			defer wg.Done()
			j(chanSlice[i], chanSlice[i+1])
			close(chanSlice[i+1])
		}(i, j)
	}
	wg.Wait()
}

// тестовый запуск
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
