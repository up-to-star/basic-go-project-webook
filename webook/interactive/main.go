package main

import (
	intrv1 "github.com/basic-go-project-webook/webook/api/proto/gen/intr/v1"
	"github.com/basic-go-project-webook/webook/interactive/grpc"
	grpc2 "google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	server := grpc2.NewServer()
	intrSvc := &grpc.InteractiveServiceServer{}
	intrv1.RegisterInteractiveServiceServer(server, intrSvc)
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	err = server.Serve(l)
	log.Println(err)
}
