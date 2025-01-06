package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoimpl"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Тут вы пишете код

// Обращаю ваше внимание - в этом задании запрещены глобальные переменные

type AdminService struct {
	muLogger     sync.Mutex
	muStat       sync.Mutex
	pipeEvent    chan *Event // Канал совместный с сервисом biz для передачи событий
	pipeStat     chan *Event // Канал совместный с сервисом biz для передачи статистики
	host         string
	aclMap       map[string][]string
	subsEvent    map[grpc.ServerStreamingServer[Event]]struct{} // Стримы подписанные для отправки событий
	stat         sync.Map                                       // Хранилище статистики
	onceCalcStat sync.Once                                      // Разовый запуск горутины расчета статистики
}

// getCtxVal получаем значение по ключу из контекста grpc
func getCtxVal(ctx context.Context, key string) string {
	if vals := metadata.ValueFromIncomingContext(ctx, key); len(vals) > 0 {
		return vals[0]
	} else {
		return ""
	}
}

// getMethod получение названия вызванного gRPC метода
func getMethod(ctx context.Context) string {
	method, ok := grpc.Method(ctx)
	if !ok {
		return ""
	}
	return method
}

// getRemoteAddr получение из контекста адрес вызывающего метод
func getRemoteAddr(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

// NewAdminService конструктор для структуры AdminService
func NewAdminService(pipeEvent chan *Event, pipeStat chan *Event, host string, ACLMap map[string][]string) *AdminService {
	subsEvent := make(map[grpc.ServerStreamingServer[Event]]struct{})

	return &AdminService{
		sync.Mutex{},
		sync.Mutex{},
		pipeEvent,
		pipeStat,
		host,
		ACLMap,
		subsEvent,
		sync.Map{},
		sync.Once{},
	}
}

// Logging подписывает пользователя на получение логов, стримит вызовы
func (a *AdminService) Logging(n *Nothing, str grpc.ServerStreamingServer[Event]) error {
	if !checkConsumer(str.Context(), a.aclMap) {
		return status.Error(codes.Unauthenticated, "access denied")
	}
	consumer := getCtxVal(str.Context(), "consumer")
	method := getMethod(str.Context())
	// добавляем стрим в мапу, если он новый
	a.muLogger.Lock()
	if _, ok := a.subsEvent[str]; !ok {
		a.subsEvent[str] = struct{}{}
	}
	a.muLogger.Unlock()
	a.pipeEvent <- &Event{protoimpl.MessageState{},
		0, protoimpl.UnknownFields{},
		time.Now().Unix(),
		consumer,
		method,
		getRemoteAddr(str.Context())}

	for val := range a.pipeEvent {
		// отправляем событие во все стримы
		a.muLogger.Lock()
		for subs := range a.subsEvent {
			// остановка отправки сообщения о логировании своего подписчика себе
			if getCtxVal(subs.Context(), "consumer") == val.Consumer && val.Method == "/main.Admin/Logging" {
				continue
			}
			//go a.updateStat(val.Consumer, val.Method) // запуск обновления статистики
			//fmt.Println(subs, subs, val.Consumer, consumer, val.Method, method)
			subs.Send(val)
		}
		a.muLogger.Unlock()
	}
	return nil
}

// updateStat обновляет статистику по совершенному вызову
func (a *AdminService) updateStat(mes *Event) {
	wg := &sync.WaitGroup{}
	a.stat.Range(func(key, value interface{}) bool {
		str := key.(grpc.ServerStreamingServer[Stat])
		stat := value.(*Stat)
		if mes.Consumer == getCtxVal(str.Context(), "consumer") && strings.HasSuffix(mes.Method, "/Statistics") {
			return true // Это аналог continue
		}
		wg.Add(1)
		// Запускаем параллельно обновление статистики для стрима
		go func() {
			defer wg.Done()
			a.muStat.Lock()
			if val, ok := stat.ByMethod[mes.Method]; ok {
				stat.ByMethod[mes.Method] = val + 1
			} else {
				stat.ByMethod[mes.Method] = 1
			}
			if val, ok := stat.ByConsumer[mes.Consumer]; ok {
				stat.ByConsumer[mes.Consumer] = val + 1
			} else {
				stat.ByConsumer[mes.Consumer] = 1
			}
			a.stat.Store(key, stat) // перезаписываем обновленное значение
			a.muStat.Unlock()
		}()
		//fmt.Println("updateStat end", mes.Method, str, a.stat)
		return true
	})
	wg.Wait() // Ждем все расчеты статистики
}

func (a *AdminService) sendStat(si *StatInterval, str grpc.ServerStreamingServer[Stat]) {
	tik := time.NewTicker(time.Duration(si.IntervalSeconds) * time.Second)
	defer tik.Stop()
	for {
		select {
		case <-tik.C:
			//fmt.Println("Отправка статистики", str, a.stat[str])
			stat, _ := a.stat.Load(str)
			statRes := stat.(*Stat)
			fmt.Println("Отправка статистики", str, statRes)
			str.Send(statRes)
			//сбрасываем статистику
			fmt.Println("Сброс статистики", str, statRes)
			statNew := Stat{}
			statNew.ByConsumer = make(map[string]uint64)
			statNew.ByMethod = make(map[string]uint64)
			a.stat.Store(str, &statNew)
		}
	}
}

// Statistics подписывает на получение статистики и стримит ее
func (a *AdminService) Statistics(si *StatInterval, str grpc.ServerStreamingServer[Stat]) error {
	if !checkConsumer(str.Context(), a.aclMap) {
		return status.Error(codes.Unauthenticated, "access denied")
	}
	// добавляем стрим в мапу
	stat := Stat{}
	stat.ByMethod = make(map[string]uint64)
	stat.ByConsumer = make(map[string]uint64)
	a.stat.Store(str, &stat)

	consumer := getCtxVal(str.Context(), "consumer")
	method := getMethod(str.Context())
	// фиксируем текущий вызов
	a.pipeStat <- &Event{protoimpl.MessageState{},
		0, protoimpl.UnknownFields{},
		time.Now().Unix(),
		consumer,
		method,
		getRemoteAddr(str.Context())}

	// запускаем отправку статистики
	go a.sendStat(si, str)

	// запускаем процесс обновления статистики при первом запуске
	a.onceCalcStat.Do(func() {
		for val := range a.pipeStat {
			go a.updateStat(val) // запуск обновления статистики
		}
	})
	return nil
}

// mustEmbedUnimplementedAdminServer нереализуемый метод требуемый grpc
func (a *AdminService) mustEmbedUnimplementedAdminServer() {
}

type BizService struct {
	mu        sync.RWMutex
	pipeEvent chan *Event // Канал отправки сообщений о вызовах
	pipeStat  chan *Event // Канал отправки сообщений о статистике
	host      string
	aclMap    map[string][]string
}

func NewBizService(pipeEvent chan *Event, pipeStat chan *Event, host string, ACLMap map[string][]string) *BizService {
	return &BizService{sync.RWMutex{}, pipeEvent, pipeStat, host, ACLMap}
}

// checkConsumer проверка подписчика по карте доступов
func checkConsumer(ctx context.Context, ACLMap map[string][]string) bool {
	var (
		methodCall    string
		methodListACL []string
		ok            bool
	)
	if methodListACL, ok = ACLMap[getCtxVal(ctx, "consumer")]; !ok || len(methodListACL) == 0 {
		return false
	}
	methodCall = getMethod(ctx)
	for _, methodACL := range methodListACL {
		if strings.HasSuffix(methodACL, "/*") {
			path, _ := strings.CutSuffix(methodACL, "/*")
			if strings.HasPrefix(methodCall, path) {
				return true
			}
		} else {
			if methodCall == methodACL {
				return true
			}
		}
	}
	return false
}

func (b *BizService) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	if !checkConsumer(ctx, b.aclMap) {
		return nil, status.Error(codes.Unauthenticated, "access denied")
	}
	b.pipeEvent <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	if !checkConsumer(ctx, b.aclMap) {
		return nil, status.Error(codes.Unauthenticated, "access denied")
	}
	b.pipeEvent <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	if !checkConsumer(ctx, b.aclMap) {
		return nil, status.Error(codes.Unauthenticated, "access denied")
	}
	b.pipeEvent <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- &Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) mustEmbedUnimplementedBizServer() {
}

// runGRPCServer запуск сервера логирования
func runGRPCServer(ctx context.Context, listenAddr string, ACLMap map[string][]string) {
	pipeEvent := make(chan *Event, 10)
	pipeStat := make(chan *Event, 10)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listen port address", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	defer server.Stop()
	RegisterAdminServer(server, NewAdminService(pipeEvent, pipeStat, listenAddr, ACLMap)) // регистрируем микросервис Admin на grpc сервере
	RegisterBizServer(server, NewBizService(pipeEvent, pipeStat, listenAddr, ACLMap))     // регистрируем микросервис Biz на grpc сервере
	fmt.Printf("starting server at %s\n", listenAddr)
	go server.Serve(lis)
	for {
		select {
		case <-ctx.Done(): // ждем сигнала завершения
			close(pipeEvent)
			close(pipeStat)
			return
		}
	}
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	ACLMap := make(map[string][]string)
	err = json.Unmarshal([]byte(ACLData), &ACLMap)
	if err != nil {
		return
	}
	go runGRPCServer(ctx, listenAddr, ACLMap)
	return nil
}
