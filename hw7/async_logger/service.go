package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type AdminService struct {
	mu       sync.RWMutex
	sessions string
}

func NewAdminService() *AdminService {
	return &AdminService{sync.RWMutex{}, ""}
}
func (a *AdminService) Logging(n *Nothing, str grpc.ServerStreamingServer[Event]) error {
	return nil
}
func (a *AdminService) Statistics(si *StatInterval, str grpc.ServerStreamingServer[Stat]) error {
	return nil
}
func (a *AdminService) mustEmbedUnimplementedAdminServer() {
}

//Logging(*Nothing, grpc.ServerStreamingServer[Event]) error
//Statistics(*StatInterval, grpc.ServerStreamingServer[Stat]) error
//mustEmbedUnimplementedAdminServer()

type BizService struct {
	mu       sync.RWMutex
	sessions string
}

func NewBizService() *BizService {
	return &BizService{sync.RWMutex{}, ""}
}

func (a *BizService) Check(ctx context.Context) error {
	return nil
}
func (a *BizService) Add(ctx context.Context) error {
	return nil
}
func (a *BizService) Test(ctx context.Context) error {
	return nil
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalln("cant listen port addres", err)
	}

	server := grpc.NewServer()
	RegisterAdminServer(server, NewAdminService())

	fmt.Printf("starting server at %s\n", listenAddr)
	server.Serve(lis)

	return nil
}

func main() {
	StartMyMicroservice(context.Background(), "", "")
}
