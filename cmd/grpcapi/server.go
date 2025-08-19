package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"schoolmanagementGRPC/internals/api/handlers"
	"schoolmanagementGRPC/internals/respositories/mongodb"
	pb "schoolmanagementGRPC/proto/gen"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	mongodb.CreateMongoClient()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	s := grpc.NewServer()
	pb.RegisterExecsServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsServiceServer(s, &handlers.Server{})
	pb.RegisterTeachersServiceServer(s, &handlers.Server{})

	reflection.Register(s)

	port := os.Getenv("SERVER_PORT")
	fmt.Println("Server runnig on port:", port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Eror listning on  the port", err)
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatal("Failed to serve")
	}
}
