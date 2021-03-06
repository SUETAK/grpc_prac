package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	myapp "prac-grpc/pkg/grpc"
	"time"
)

func main() {
	// 1. 8080番portのLisnterを作成
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// 2. gRPCサーバーを作成
	s := grpc.NewServer()

	// gRPCサーバにGreetingService を登録
	myapp.RegisterGreetingServiceServer(s, NewMyServer())

	reflection.Register(s)

	// 3. 作成したgRPCサーバーを、8080番ポートで稼働させる
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// 4.Ctrl+Cが入力されたらGraceful shutdownされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}

func NewMyServer() *myServer {
	return &myServer{}
}

type myServer struct {
	myapp.UnimplementedGreetingServiceServer
}

func (s myServer) Hello(ctx context.Context, req *myapp.HelloRequest) (*myapp.HelloResponse, error) {
	// リクエストからnameフィールドを取り出して
	// "Hello, [名前]!"というレスポンスを返す
	return &myapp.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

// HelloServerStream myapp_grpc.pb.go に定義されたシグネチャ
// HelloServerStream (*HelloRequest, GreetingService_HelloServerStreamServer) error
func (s *myServer) HelloServerStream(req *myapp.HelloRequest, stream myapp.GreetingService_HelloServerStreamServer) error {
	resCount := 5
	for i := 0; i < resCount; i++ {
		// stream のSendメソッドを使っている
		if err := stream.Send(&myapp.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	// nil を返却することでストリームを終了させる
	return nil
}
