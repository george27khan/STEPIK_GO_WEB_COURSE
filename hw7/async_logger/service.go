package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoimpl"
	"strings"
	"time"
	//"golang.org/x/text/message/pipeline"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// тут вы пишете код

type Service struct {
	AdminService
	BizService
}

// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type AdminService struct {
	muStatByMethod   sync.Mutex
	muStatByConsumer sync.Mutex
	muLogger         sync.Mutex
	muStat           sync.Mutex
	sessions         string
	pipeEvent        chan Event // канал совместный с сервисом biz для обмена сообщениями
	pipeStat         chan Event // канал совместный с сервисом biz для обмена сообщениями
	host             string
	aclMap           map[string][]string
	subsEvent        map[grpc.ServerStreamingServer[Event]]struct{} //стримы подписанные для отправки событий
	subsStat         map[grpc.ServerStreamingServer[Stat]]struct{}  //стримы подписанные для отправки статистики
	stat             map[grpc.ServerStreamingServer[Stat]]Stat      // хранилище статистики
	timer            time.Time
	onceCalcStat     sync.Once
	onceSendStat     sync.Once
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
func NewAdminService(pipeEvent chan Event, pipeStat chan Event, host string, ACLMap map[string][]string) *AdminService {
	subsEvent := make(map[grpc.ServerStreamingServer[Event]]struct{})
	subsStat := make(map[grpc.ServerStreamingServer[Stat]]struct{})
	stat := make(map[grpc.ServerStreamingServer[Stat]]Stat)
	return &AdminService{
		sync.Mutex{},
		sync.Mutex{},
		sync.Mutex{},
		sync.Mutex{},
		"",
		pipeEvent,
		pipeStat,
		host,
		ACLMap,
		subsEvent,
		subsStat,
		stat,
		time.Now(),
		sync.Once{},
		sync.Once{}}
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
	a.pipeEvent <- Event{protoimpl.MessageState{},
		0, protoimpl.UnknownFields{},
		time.Now().Unix(),
		consumer,
		method,
		getRemoteAddr(str.Context())}

	for val := range a.pipeEvent {
		// отправляем событие во все стримы
		a.muLogger.Lock()
		for subs, _ := range a.subsEvent {
			// остановка отправки сообщения о логировании своего подписчика себе
			if getCtxVal(subs.Context(), "consumer") == val.Consumer && val.Method == "/main.Admin/Logging" {
				continue
			}
			//go a.updateStat(val.Consumer, val.Method) // запуск обновления статистики
			fmt.Println(subs, subs, val.Consumer, consumer, val.Method, method)
			subs.Send(&val)
		}
		a.muLogger.Unlock()
	}
	return nil
}

// updateStat обновляет статистику по совершенному вызову
func (a *AdminService) updateStat(mes Event) {
	wg := &sync.WaitGroup{}
	for str, stat := range a.stat {
		if mes.Consumer == getCtxVal(str.Context(), "consumer") && strings.HasSuffix(mes.Method, "/Statistics") {
			continue
		}
		fmt.Println("updateStat start", mes.Method, str, a.stat)
		wg.Add(2)
		go func() {
			defer wg.Done()
			a.muStatByMethod.Lock()
			if val, ok := stat.ByMethod[mes.Method]; ok {
				a.stat[str].ByMethod[mes.Method] = val + 1
			} else {
				a.stat[str].ByMethod[mes.Method] = 1
			}
			a.muStatByMethod.Unlock()
		}()
		go func() {
			defer wg.Done()
			a.muStatByConsumer.Lock()
			if val, ok := stat.ByConsumer[mes.Consumer]; ok {
				a.stat[str].ByConsumer[mes.Consumer] = val + 1
			} else {
				a.stat[str].ByConsumer[mes.Consumer] = 1
			}
			a.muStatByConsumer.Unlock()
		}()
		fmt.Println("updateStat end", mes.Method, str, a.stat)
	}
	wg.Wait()
}

func (a *AdminService) sendStat(si *StatInterval, str grpc.ServerStreamingServer[Stat]) {
	tik := time.NewTicker(time.Duration(si.IntervalSeconds) * time.Second)
	defer tik.Stop()
	for {
		select {
		case <-tik.C:
			a.muStatByConsumer.Lock()
			a.muStatByMethod.Lock()
			fmt.Println("Отправка статистики", str, a.stat[str])
			fmt.Println(getCtxVal(str.Context(), "consumer"), time.Since(a.timer))
			statRes := a.stat[str]
			fmt.Println(str, statRes)
			str.Send(&statRes)
			//сбрасываем статистику
			fmt.Println("Сброс статистики", str, a.stat[str])
			stat := Stat{}
			stat.ByConsumer = make(map[string]uint64)
			stat.ByMethod = make(map[string]uint64)
			a.stat[str] = stat
			a.muStatByConsumer.Unlock()
			a.muStatByMethod.Unlock()
			//for sub, _ := range a.subsStat {
			//	// отправляем статистику во все стримы\
			//	//fmt.Println(subs, subs, val.Consumer, consumer, val.Method, method)
			//	statRes := a.stat[sub]
			//	sub.Send(&statRes)
			//
			//	//сбрасываем статистику
			//	fmt.Println("Сброс статистики ", a.stat)
			//	stat := Stat{}
			//	stat.ByConsumer = make(map[string]uint64)
			//	stat.ByMethod = make(map[string]uint64)
			//	a.stat[sub] = stat
			//}
			//default:
			//	fmt.Printf("WAIT %v", str)
		}
	}
}

// Statistics подписывает на получение статистики и стримит ее
func (a *AdminService) Statistics(si *StatInterval, str grpc.ServerStreamingServer[Stat]) error {
	if !checkConsumer(str.Context(), a.aclMap) {
		return status.Error(codes.Unauthenticated, "access denied")
	}
	a.muStatByConsumer.Lock()
	// добавляем стрим в мапу, если он новый
	if _, ok := a.stat[str]; !ok {
		stat := Stat{}
		stat.ByMethod = make(map[string]uint64)
		stat.ByConsumer = make(map[string]uint64)
		a.stat[str] = stat
	}
	a.muStatByConsumer.Unlock()
	consumer := getCtxVal(str.Context(), "consumer")
	method := getMethod(str.Context())
	// фиксируем текущий вызов
	a.pipeStat <- Event{protoimpl.MessageState{},
		0, protoimpl.UnknownFields{},
		time.Now().Unix(),
		consumer,
		method,
		getRemoteAddr(str.Context())}
	// отправки статистики
	go a.sendStat(si, str)

	a.onceCalcStat.Do(func() {
		for val := range a.pipeStat {
			//// отправляем событие во все стримы
			//if val.Consumer == consumer && val.Method == "/main.Admin/Statistics" {
			//	continue
			//}
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
	sessions  string
	pipeEvent chan Event // канал отправкии сообщений о вызовах
	pipeStat  chan Event // канал отправкии сообщений о статистике
	host      string
	aclMap    map[string][]string
}

func NewBizService(pipeEvent chan Event, pipeStat chan Event, host string, ACLMap map[string][]string) *BizService {
	return &BizService{sync.RWMutex{}, "", pipeEvent, pipeStat, host, ACLMap}
}

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
	b.pipeEvent <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	if !checkConsumer(ctx, b.aclMap) {
		return nil, status.Error(codes.Unauthenticated, "access denied")
	}
	b.pipeEvent <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	if !checkConsumer(ctx, b.aclMap) {
		return nil, status.Error(codes.Unauthenticated, "access denied")
	}
	b.pipeEvent <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	b.pipeStat <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), getMethod(ctx), getRemoteAddr(ctx)}
	return nil, nil
}
func (b *BizService) mustEmbedUnimplementedBizServer() {
}

func runGRPCServer(ctx context.Context, listenAddr string, ACLMap map[string][]string) {
	pipeEvent := make(chan Event, 10)
	pipeStat := make(chan Event, 10)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listen port addres", err)
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
	//fmt.Println(err, ACLMap)
	go runGRPCServer(ctx, listenAddr, ACLMap)
	return nil
}

//// чтобы не было сюрпризов когда где-то не успела преключиться горутина и не успело что-то стортовать
//func wait(amout int) {
//	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
//}

//// утилитарная функция для коннекта к серверу
//func getGrpcConn() *grpc.ClientConn {
//	listenAddr := "127.0.0.1:8082"
//	grcpConn, err := grpc.NewClient(
//		listenAddr,
//		grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//	if err != nil {
//		fmt.Printf("cant connect to grpc: %v\n", err)
//	}
//	return grcpConn
//}

//// получаем контекст с нужнымы метаданными для ACL
//func getConsumerCtx(consumerName string) context.Context {
//	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
//	ctx := context.Background()
//	md := metadata.Pairs(
//		"consumer", consumerName,
//	)
//	return metadata.NewOutgoingContext(ctx, md)
//}

func main() {
	//	ACLData := `{
	//	"logger1":          ["/main.Admin/Logging"],
	//	"logger2":          ["/main.Admin/Logging"],
	//	"stat1":            ["/main.Admin/Statistics"],
	//	"stat2":            ["/main.Admin/Statistics"],
	//	"biz_user":         ["/main.Biz/Check", "/main.Biz/Add"],
	//	"biz_admin":        ["/main.Biz/*"],
	//	"after_disconnect": ["/main.Biz/Add"]
	//}`
	//	listenAddr := "127.0.0.1:8082"
	//
	//	ctx, finish := context.WithCancel(context.Background())
	//	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	//	if err != nil {
	//		fmt.Printf("cant start server initial: %v\n", err)
	//	}
	//	wait(1)
	//	defer func() {
	//		finish()
	//		wait(1)
	//	}()
	//
	//	conn := getGrpcConn()
	//	defer conn.Close()
	//
	//	biz := NewBizClient(conn)
	//	adm := NewAdminClient(conn)
	//
	//	statStream1, err := adm.Statistics(getConsumerCtx("stat1"), &StatInterval{IntervalSeconds: 2})
	//	wait(1)
	//	statStream2, err := adm.Statistics(getConsumerCtx("stat2"), &StatInterval{IntervalSeconds: 3})
	//	statStream2 = statStream2
	//	mu := &sync.Mutex{}
	//	stat1 := &Stat{}
	//	stat2 := &Stat{}
	//	stat2 = stat2
	//	wg := &sync.WaitGroup{}
	//	fmt.Println("----------------------------------------------------")
	//	wg.Add(2)
	//	go func() {
	//		for {
	//			stat, err := statStream1.Recv()
	//			if err != nil && err != io.EOF {
	//				// fmt.Printf("unexpected error %v\n", err)
	//				return
	//			} else if err == io.EOF {
	//				break
	//			}
	//			// log.Println("stat1", stat, err)
	//			mu.Lock()
	//			// это грязный хак
	//			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
	//			// поэтому берем не оригинал сообщения, а только нужные значения
	//			stat1 = &Stat{
	//				ByMethod:   stat.ByMethod,
	//				ByConsumer: stat.ByConsumer,
	//			}
	//			mu.Unlock()
	//		}
	//	}()
	//	go func() {
	//		for {
	//			stat, err := statStream2.Recv()
	//			if err != nil && err != io.EOF {
	//				// fmt.Printf("unexpected error %v\n", err)
	//				return
	//			} else if err == io.EOF {
	//				break
	//			}
	//			// log.Println("stat2", stat, err)
	//			mu.Lock()
	//			// это грязный хак
	//			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
	//			// поэтому берем не оригинал сообщения, а только нужные значения
	//			stat2 = &Stat{
	//				ByMethod:   stat.ByMethod,
	//				ByConsumer: stat.ByConsumer,
	//			}
	//			mu.Unlock()
	//		}
	//	}()
	//
	//	wait(1)
	//
	//	biz.Check(getConsumerCtx("biz_user"), &Nothing{})
	//	biz.Add(getConsumerCtx("biz_user"), &Nothing{})
	//	biz.Test(getConsumerCtx("biz_admin"), &Nothing{})
	//
	//	wait(200) // 2 sec
	//
	//	expectedStat1 := &Stat{
	//		ByMethod: map[string]uint64{
	//			"/main.Biz/Check":        1,
	//			"/main.Biz/Add":          1,
	//			"/main.Biz/Test":         1,
	//			"/main.Admin/Statistics": 1,
	//		},
	//		ByConsumer: map[string]uint64{
	//			"biz_user":  2,
	//			"biz_admin": 1,
	//			"stat2":     1,
	//		},
	//	}
	//
	//	mu.Lock()
	//	if !reflect.DeepEqual(stat1, expectedStat1) {
	//		fmt.Printf("stat1-1 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	//	} else {
	//		fmt.Printf("DONE 1 !!!!!!!!!!!!!!!!!!!!!!!!\n")
	//	}
	//	mu.Unlock()
	//
	//	biz.Add(getConsumerCtx("biz_admin"), &Nothing{})
	//
	//	wait(200) // 2+ sec
	//
	//	expectedStat1 = &Stat{
	//		Timestamp: 0,
	//		ByMethod: map[string]uint64{
	//			"/main.Biz/Add": 1,
	//		},
	//		ByConsumer: map[string]uint64{
	//			"biz_admin": 1,
	//		},
	//	}
	//	expectedStat2 := &Stat{
	//		Timestamp: 0,
	//		ByMethod: map[string]uint64{
	//			"/main.Biz/Check": 1,
	//			"/main.Biz/Add":   2,
	//			"/main.Biz/Test":  1,
	//		},
	//		ByConsumer: map[string]uint64{
	//			"biz_user":  2,
	//			"biz_admin": 2,
	//		},
	//	}
	//
	//	mu.Lock()
	//	if !reflect.DeepEqual(stat1, expectedStat1) {
	//		fmt.Printf("stat1-2 dont match\nhave %+v\nwant %+v", stat1, expectedStat1)
	//	} else {
	//		fmt.Printf("DONE 1-2 !!!!!!!!!!!!!!!!!!!!!!!!\n")
	//	}
	//	if !reflect.DeepEqual(stat2, expectedStat2) {
	//		fmt.Printf("stat2 dont match\nhave %+v\nwant %+v", stat2, expectedStat2)
	//	} else {
	//		fmt.Printf("DONE 2 !!!!!!!!!!!!!!!!!!!!!!!!\n")
	//	}
	//	mu.Unlock()
	//
	//	finish()

}
