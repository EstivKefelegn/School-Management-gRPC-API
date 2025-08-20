package handlers

import (
	"context"
	"fmt"
	"reflect"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/internals/respositories/mongodb"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	// Filtering getting the filtes from the request, another function
	filter, err := buildFilterForTeacher(req.Teacher)
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

func buildFilterForTeacher(teacherObj *pb.Teacher) (bson.M, error) {
	filter := bson.M{}

	if teacherObj == nil {
		// This means there is no data to filter
		return filter, nil
	}

	var modelTeacher models.Teacher
	modelVal := reflect.ValueOf(&modelTeacher).Elem()
	modelType := modelVal.Type()

	reqVal := reflect.ValueOf(teacherObj).Elem()
	reqType := reqVal.Type()

	for i := 0; i < reqVal.NumField(); i++ {
		fieldVal := reqVal.Field(i)
		fieldName := reqType.Field(i).Name

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			modelField := modelVal.FieldByName(fieldName)
			if modelField.IsValid() && modelField.CanSet() {
				modelField.Set(fieldVal)
			}
		}
	}

	for i := 0; i < modelVal.NumField(); i++ {
		fieldVal := modelVal.Field(i)

		if fieldVal.IsValid() && !fieldVal.IsZero() {
			bsonTag := modelType.Field(i).Tag.Get("bson")
			bsonTag = strings.TrimSuffix(bsonTag, ",omitempty")

			// Filter with the specific id
			if bsonTag == "_id" {
				objId, err := primitive.ObjectIDFromHex(teacherObj.Id)
				if err != nil {
					return nil, utils.ErrorHandler(err, "Invalid Id")
				}
				filter[bsonTag] = objId
			} else {
				filter[bsonTag] = fieldVal.Interface().(string) // Converting to string we make it interface to make the field value more suitable to be converted to string
			}

		}
	}

	fmt.Println("Filter:", filter)
	return filter, nil
}

func buildSortOptions(sortFields []*pb.SortField) bson.D {
	var sortOptions bson.D
	for _, sortField := range sortFields {
		order := 1
		if sortField.GetOrder() == pb.Order_DESC {
			order = 0
		}
		sortOptions = append(sortOptions, bson.E{Key: sortField.Field, Value: order})
	}

	fmt.Println("Sort Option:", sortOptions)
	return sortOptions
}
