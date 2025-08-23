package handlers

import (
	"context"
	"fmt"
	"schoolmanagementGRPC/internals/respositories/mongodb"
	pb "schoolmanagementGRPC/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {

	// i want the id to be filled with by mongo it self so i want the requeted value of the id should be empty
	for _, Student := range req.GetStudents() {
		if Student.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "Request is in incorrect format: non empty id fields are not allowed")
		}
	}

	addedStudents, err := mongodb.AddStudentsToDb(ctx, req.GetStudents())
	fmt.Println("Students ========== ",addedStudents)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: addedStudents}, nil
}
