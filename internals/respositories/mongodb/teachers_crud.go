package mongodb

import (
	"context"
	"reflect"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/pkg/utils"

	pb "schoolmanagementGRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddTeachersToDb(ctx context.Context, teachersFromReq []*pb.Teacher) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal Error")
	}

	defer client.Disconnect(ctx)

	newTeachers := make([]*models.Teacher, len(teachersFromReq))
	for i, pbTeacher := range teachersFromReq {
		newTeachers[i] = mapPbTeacherToModelTeacher(pbTeacher)
		// newTeachers[i] = &modelTeacher
	}
	// Inserting a value
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
		pbTeacher := mapModelTeacherToPb(teacher)
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
}

func mapModelTeacherToPb(teacher *models.Teacher) *pb.Teacher {
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
	return pbTeacher
}

func mapPbTeacherToModelTeacher(pbTeacher *pb.Teacher) *models.Teacher {
	// Decoding the protobuf request to model struct
	modelTeacher := models.Teacher{}
	pbVal := reflect.ValueOf(pbTeacher).Elem()
	modelVal := reflect.ValueOf(&modelTeacher).Elem()

	for i := 0; i < pbVal.NumField(); i++ {
		pbField := pbVal.Field(i)               // getting the field names
		fieldName := pbVal.Type().Field(i).Name // The values in the field

		modelField := modelVal.FieldByName(fieldName)

		if modelField.IsValid() && modelField.CanSet() {
			modelField.Set(pbField)
		} else {
			// fmt.Printf("Field %s is not valid or cannot be set\n", fieldName)
			continue
		}
	}
	return &modelTeacher
}

func GetTeachersFromDb(ctx context.Context, sortOptions bson.D, filter bson.M) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	defer client.Disconnect(ctx)

	col1 := client.Database("school").Collection("teachers")

	var cursor *mongo.Cursor
	if len(sortOptions) < 1 {
		cursor, err = col1.Find(ctx, filter)
	} else {
		cursor, err = col1.Find(ctx, filter, options.Find().SetSort(sortOptions))
	}

	defer cursor.Close(ctx)

	var teachers []*pb.Teacher
	for cursor.Next(ctx) {
		var teacher models.Teacher
		err := cursor.Decode(&teacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}

		teachers = append(teachers, &pb.Teacher{
			Id:        teacher.Id,
			FirstName: teacher.FirstName,
			LastName:  teacher.LastName,
			Email:     teacher.Email,
			Class:     teacher.Class,
			Subject:   teacher.Subject,
		})
	}
	return teachers, nil
}
