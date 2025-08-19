package handlers

import grpcapipb "schoolmanagementGRPC/proto/gen"

type Server struct {
	grpcapipb.UnimplementedStudentsServiceServer
	grpcapipb.UnimplementedTeachersServiceServer
	grpcapipb.UnimplementedExecsServiceServer
}