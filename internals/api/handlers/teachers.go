package handlers

import (
	"context"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/internals/respositories/mongodb"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {

	// i want the id to be filled with by mongo it self so i want the requeted value of the id should be empty
	for _, teacher := range req.GetTeachers() {
		if teacher.Id != "" {
			return nil, status.Error(codes.InvalidArgument, "Request is in incorrect format: non empty id fields are not allowed")
		}
	}

	addedTeachers, err := mongodb.AddTeachersToDb(ctx, req.GetTeachers())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Teachers{Teachers: addedTeachers}, nil
}

func (s *Server) GetTeachers(ctx context.Context, req *pb.GetTeachersRequest) (*pb.Teachers, error) {

	filter, err := buildFilter(req.Teacher, &models.Teacher{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	// Sorting getting the sort options from the requets
	sortOptions := buildSortOptions(req.GetSortField())

	// Access the database to fetch data,  another function
	teachers, err := mongodb.GetTeachersFromDb(ctx, sortOptions, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Teachers{Teachers: teachers}, nil

}

func (s *Server) UpdateTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	updatedTeachers, err := mongodb.ModifyTeachersfromDb(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.Teachers{Teachers: updatedTeachers}, nil
}

func (s *Server) DeleteTeachers(ctx context.Context, req *pb.TeacherIds) (*pb.DeleteTeachersConfirmation, error) {
	ids := req.GetIds()
	var teacherIdsToDelete []string
	for _, v := range ids {
		teacherIdsToDelete = append(teacherIdsToDelete, v.Id)
	}

	deletedIds, err := mongodb.DeleteTeachersFromDb(ctx, teacherIdsToDelete)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteTeachersConfirmation{
		Status:     "Teachers successfully deleted",
		DeletedIds: deletedIds,
	}, nil

}

func (s *Server) GetStudentCountByClassTeacher(ctx context.Context, req *pb.TeacherId) (*pb.StudentCount, error) {
	teacherId := req.GetId()

	count, err := mongodb.GetStudentCountByTeacherIdFromDb(ctx, teacherId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StudentCount{Status: true, StudentCount: int32(count)}, nil
}
