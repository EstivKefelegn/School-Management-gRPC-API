package handlers

import (
	"context"
	"schoolmanagementGRPC/internals/models"
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

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: addedStudents}, nil
}

func (s *Server) GetStudents(ctx context.Context, req *pb.GetStudentsRequest) (*pb.Students, error) {
	filter, err := buildFilter(req.Student, &models.Student{})
	if err != nil {
		return nil, err
	}

	sortOptions := buildSortOptions(req.GetSortField())

	pageNumber := req.GetPageNumber()
	pageSize := req.GetPageSize()

	if pageNumber < 1 {
		pageNumber = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	students, err := mongodb.GetStudentsFromDb(ctx, sortOptions, filter, uint32(pageNumber), uint32(pageSize))
	if err != nil {
		return nil, err
	}

	return &pb.Students{Students: students}, nil
}

func (s *Server) UpdateStudents(ctx context.Context, req *pb.Students) (*pb.Students, error) {
	updatedStudenst, err := mongodb.ModifyStudentInDb(ctx, req.Students)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Students{Students: updatedStudenst}, nil
}

func (s *Server) DeleteStudents(ctx context.Context, req *pb.StudentIds) (*pb.DeleteStudentsConfirmation, error) {
	ids := req.GetIds()
	var studentIdsToDelete []string
	for _, v := range ids {
		studentIdsToDelete = append(studentIdsToDelete, v.Id)
	}

	deletedIds, err := mongodb.DeleteStudentsFromDb(ctx, studentIdsToDelete)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteStudentsConfirmation{
		Status:     "Students successfully deleted",
		DeletedIds: deletedIds,
	}, nil

}
