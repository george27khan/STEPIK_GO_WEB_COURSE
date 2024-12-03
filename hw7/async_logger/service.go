package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/runtime/protoimpl"
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
	mu       sync.RWMutex
	sessions string
	pipe     chan Event
	host     string
}

// получаем значение по ключу из контекста grpc
func getCtxVal(ctx context.Context, key string) string {
	//t1, t2 := metadata.FromIncomingContext(ctx)
	//fmt.Println("ctx", t1, t2)
	if vals := metadata.ValueFromIncomingContext(ctx, key); len(vals) > 0 {
		//fmt.Println(2)
		return vals[0]
	} else {
		//fmt.Println(3)
		return ""
	}
}

func NewAdminService(pipe chan Event, host string) *AdminService {
	return &AdminService{sync.RWMutex{}, "", pipe, host}
}

// возвращает клиенту стрим, и в него будем отправлять вызовы пришедшие в канал
func (a *AdminService) Logging(n *Nothing, str grpc.ServerStreamingServer[Event]) error {
	//str.
	for val := range a.pipe {
		str.Send(&val)
	}
	return nil
}
func (a *AdminService) Statistics(si *StatInterval, str grpc.ServerStreamingServer[Stat]) error {
	return nil
}
func (a *AdminService) mustEmbedUnimplementedAdminServer() {
}

type BizService struct {
	mu       sync.RWMutex
	sessions string
	pipe     chan Event
	host     string
}

func NewBizService(pipe chan Event, host string) *BizService {
	return &BizService{sync.RWMutex{}, "", pipe, host}
}

func (b *BizService) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	b.pipe <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), "/main.Biz/Check", b.host + "/main.Biz/Check"}
	return nil, nil
}
func (b *BizService) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	b.pipe <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), "/main.Biz/Add", b.host + "/main.Biz/Add"}
	return nil, nil
}
func (b *BizService) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	b.pipe <- Event{protoimpl.MessageState{}, 0, protoimpl.UnknownFields{}, time.Now().Unix(), getCtxVal(ctx, "consumer"), "/main.Biz/Test", b.host + "/main.Biz/Test"}
	return nil, nil
}
func (b *BizService) mustEmbedUnimplementedBizServer() {
}

func runGRPCServer(ctx context.Context, listenAddr string) {
	pipe := make(chan Event)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listen port addres", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	defer server.Stop()
	RegisterAdminServer(server, NewAdminService(pipe, listenAddr)) // регистрируем микросервис Admin на grpc сервере
	RegisterBizServer(server, NewBizService(pipe, listenAddr))     // регистрируем микросервис Biz на grpc сервере
	fmt.Printf("starting server at %s\n", listenAddr)
	go server.Serve(lis)
	<-ctx.Done() // ждем сигнала завершения
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	ACLMap := make(map[string][]string)
	err = json.Unmarshal([]byte(ACLData), &ACLMap)
	if err != nil {
		return
	}
	//fmt.Println(err, ACLMap)
	go runGRPCServer(ctx, listenAddr)
	return nil
}

//
//// чтобы не было сюрпризов когда где-то не успела преключиться горутина и не успело что-то стортовать
//func wait(amout int) {
//	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
//}
//
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
//
//// получаем контекст с нужнымы метаданными для ACL
//func getConsumerCtx(consumerName string) context.Context {
//	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
//	ctx := context.Background()
//	md := metadata.Pairs(
//		"consumer", consumerName,
//	)
//	return metadata.NewOutgoingContext(ctx, md)
//}

//func main() {
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
//	logStream1, err := adm.Logging(getConsumerCtx("logger1"), &Nothing{})
//	time.Sleep(1 * time.Millisecond)
//
//	logStream2, err := adm.Logging(getConsumerCtx("logger2"), &Nothing{})
//
//	logData1 := []*Event{}
//	logData2 := []*Event{}
//
//	wait(1)
//
//	go func() {
//		select {
//		case <-ctx.Done():
//			return
//		case <-time.After(3 * time.Second):
//			fmt.Println("looks like you dont send anything to log stream in 3 sec")
//		}
//	}()
//
//	wg := &sync.WaitGroup{}
//	wg.Add(2)
//	go func() {
//		defer wg.Done()
//		for i := 0; i < 4; i++ {
//			evt, err := logStream1.Recv()
//			// log.Println("logger 1", evt, err)
//			if err != nil {
//				fmt.Printf("unexpected error: %v, awaiting event", err)
//				return
//			}
//			// evt.Host читайте как evt.RemoteAddr
//			if !strings.HasPrefix(evt.GetHost(), "127.0.0.1:") || evt.GetHost() == listenAddr {
//				fmt.Printf("bad host: %v", evt.GetHost())
//				return
//			}
//			// это грязный хак
//			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
//			// поэтому берем не оригинал сообщения, а только нужные значения
//			logData1 = append(logData1, &Event{Consumer: evt.Consumer, Method: evt.Method})
//		}
//	}()
//	go func() {
//		defer wg.Done()
//		for i := 0; i < 3; i++ {
//			evt, err := logStream2.Recv()
//			// log.Println("logger 2", evt, err)
//			if err != nil {
//				fmt.Printf("unexpected error: %v, awaiting event", err)
//				return
//			}
//			if !strings.HasPrefix(evt.GetHost(), "127.0.0.1:") || evt.GetHost() == listenAddr {
//				fmt.Printf("bad host: %v", evt.GetHost())
//				return
//			}
//			// это грязный хак
//			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
//			// поэтому берем не оригинал сообщения, а только нужные значения
//			logData2 = append(logData2, &Event{Consumer: evt.Consumer, Method: evt.Method})
//		}
//	}()
//
//	biz.Check(getConsumerCtx("biz_user"), &Nothing{})
//	time.Sleep(2 * time.Millisecond)
//
//	biz.Check(getConsumerCtx("biz_admin"), &Nothing{})
//	time.Sleep(2 * time.Millisecond)
//
//	biz.Test(getConsumerCtx("biz_admin"), &Nothing{})
//	time.Sleep(2 * time.Millisecond)
//
//	wg.Wait()
//	expectedLogData1 := []*Event{
//		{Consumer: "logger2", Method: "/main.Admin/Logging"},
//		{Consumer: "biz_user", Method: "/main.Biz/Check"},
//		{Consumer: "biz_admin", Method: "/main.Biz/Check"},
//		{Consumer: "biz_admin", Method: "/main.Biz/Test"},
//	}
//	expectedLogData2 := []*Event{
//		{Consumer: "biz_user", Method: "/main.Biz/Check"},
//		{Consumer: "biz_admin", Method: "/main.Biz/Check"},
//		{Consumer: "biz_admin", Method: "/main.Biz/Test"},
//	}
//
//	if !reflect.DeepEqual(logData1, expectedLogData1) {
//		fmt.Printf("logs1 dont match\nhave %+v\nwant %+v", logData1, expectedLogData1)
//	}
//	if !reflect.DeepEqual(logData2, expectedLogData2) {
//		fmt.Printf("logs2 dont match\nhave %+v\nwant %+v", logData2, expectedLogData2)
//	}
//
//}
