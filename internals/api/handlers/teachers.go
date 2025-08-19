package handlers

import (
	"context"
	"fmt"
	"reflect"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/internals/respositories/mongodb"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) AddTeachers(ctx context.Context, req *pb.Teachers) (*pb.Teachers, error) {
	client, err := mongodb.CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal Error")
	}

	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(req.GetTeachers()))
	for i, pbTeacher := range req.GetTeachers() {
		modelTeacher := models.Teacher{FirstName: "john"}
		pbVal := reflect.ValueOf(pbTeacher).Elem()
		modelVal := reflect.ValueOf(&modelTeacher).Elem()

		for i := 0; i < pbVal.NumField(); i++ {
			pbField := pbVal.Field(i)
			fieldName := pbVal.Type().Field(i).Name

			modelField := modelVal.FieldByName(fieldName)
			if modelField.IsValid() && modelField.CanSet() {
				modelField.Set(pbField)
			} else {
				// fmt.Printf("Field %s is not valid or cannot be set\n", fieldName)
				continue
			}
		}
		fmt.Println(newTeachers)
		newTeachers[i] = &modelTeacher
	}
	var addedTeachers []*pb.Teacher
	for _, teacher := range newTeachers {
		result, err := client.Database("school").Collection("teachers").InsertOne(ctx, teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}

		// insertedID is interface type so we will convert it to string HEX will convert that to string
		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			teacher.Id = objectID.Hex()
		}

		// The return the values to the user
		pbTeacher := &pb.Teacher{}
		modelVal := reflect.ValueOf(*teacher)
		pbVal := reflect.ValueOf(pbTeacher).Elem()

		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldType := modelVal.Type().Field(i)
			pbField := pbVal.FieldByName(modelFieldType.Name)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}
		addedTeachers = append(addedTeachers, pbTeacher)
	}

	return &pb.Teachers{Teachers: addedTeachers}, nil
}
