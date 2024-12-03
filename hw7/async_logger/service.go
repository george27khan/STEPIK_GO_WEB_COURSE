package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	mu       sync.RWMutex
	sessions string
	pipe     chan Event
}

func NewAdminService(pipe chan Event) *AdminService {
	return &AdminService{sync.RWMutex{}, "", pipe}
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
}

func NewBizService(pipe chan Event) *BizService {
	return &BizService{sync.RWMutex{}, "", pipe}
}

func (b *BizService) Check(ctx context.Context, _ *Nothing) (*Nothing, error) {
	b.pipe <- Event{protoimpl.MessageState{}, protoimpl.SizeCache{}, protoimpl.UnknownFields{}, time.Now().Unix(), "", "", ""}
	return nil, nil
}
func (b *BizService) Add(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return nil, nil
}
func (b *BizService) Test(ctx context.Context, _ *Nothing) (*Nothing, error) {
	return nil, nil
}
func (b *BizService) mustEmbedUnimplementedBizServer() {
}

func runGRPCServer(ctx context.Context, listenAddr string) {
	var pipe chan string
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listen port addres", err)
	}
	defer lis.Close()

	server := grpc.NewServer()
	defer server.Stop()
	RegisterAdminServer(server, NewAdminService(pipe)) // регистрируем микросервис Admin на grpc сервере
	RegisterBizServer(server, NewBizService(pipe))     // регистрируем микросервис Biz на grpc сервере
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
	fmt.Println(err, ACLMap)
	go runGRPCServer(ctx, listenAddr)
	return nil
}

func main() {
	ACLData := `{
	"logger1":          ["/main.Admin/Logging"],
	"logger2":          ["/main.Admin/Logging"],
	"stat1":            ["/main.Admin/Statistics"],
	"stat2":            ["/main.Admin/Statistics"],
	"biz_user":         ["/main.Biz/Check", "/main.Biz/Add"],
	"biz_admin":        ["/main.Biz/*"],
	"after_disconnect": ["/main.Biz/Add"]
}`
	ctx, finish := context.WithCancel(context.Background())
	err := StartMyMicroservice(ctx, listenAddr, ACLData)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}
	wait(1)
	defer func() {
		finish()
		wait(1)
	}()

	conn := getGrpcConn(t)
	defer conn.Close()

	biz := NewBizClient(conn)
	adm := NewAdminClient(conn)

	logStream1, err := adm.Logging(getConsumerCtx("logger1"), &Nothing{})
	time.Sleep(1 * time.Millisecond)

	logStream2, err := adm.Logging(getConsumerCtx("logger2"), &Nothing{})

	logData1 := []*Event{}
	logData2 := []*Event{}

	wait(1)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			fmt.Println("looks like you dont send anything to log stream in 3 sec")
			t.Errorf("looks like you dont send anything to log stream in 3 sec")
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 4; i++ {
			evt, err := logStream1.Recv()
			// log.Println("logger 1", evt, err)
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}
			// evt.Host читайте как evt.RemoteAddr
			if !strings.HasPrefix(evt.GetHost(), "127.0.0.1:") || evt.GetHost() == listenAddr {
				t.Errorf("bad host: %v", evt.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData1 = append(logData1, &Event{Consumer: evt.Consumer, Method: evt.Method})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			evt, err := logStream2.Recv()
			// log.Println("logger 2", evt, err)
			if err != nil {
				t.Errorf("unexpected error: %v, awaiting event", err)
				return
			}
			if !strings.HasPrefix(evt.GetHost(), "127.0.0.1:") || evt.GetHost() == listenAddr {
				t.Errorf("bad host: %v", evt.GetHost())
				return
			}
			// это грязный хак
			// protobuf добавляет к структуре свои поля, которвые не видны при приведении к строке и при reflect.DeepEqual
			// поэтому берем не оригинал сообщения, а только нужные значения
			logData2 = append(logData2, &Event{Consumer: evt.Consumer, Method: evt.Method})
		}
	}()

	biz.Check(getConsumerCtx("biz_user"), &Nothing{})
	time.Sleep(2 * time.Millisecond)

	biz.Check(getConsumerCtx("biz_admin"), &Nothing{})
	time.Sleep(2 * time.Millisecond)

	biz.Test(getConsumerCtx("biz_admin"), &Nothing{})
	time.Sleep(2 * time.Millisecond)

	wg.Wait()

}
