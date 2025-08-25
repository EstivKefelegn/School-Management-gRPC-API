package mongodb

import (
	"context"
	"errors"
	"fmt"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddStudentsToDb(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to db")
	}

	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, len(studentsFromReq))
	for i, student := range studentsFromReq {
		newStudents[i] = mapPbStudentToModelStudent(student)
	}

	var addedStudents []*pb.Student
	for _, student := range newStudents {
		result, err := client.Database("school").Collection("students").InsertOne(ctx, student)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error on adding a value to database")
		}

		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			student.Id = objectID.Hex()
		}

		pbStudent := mapModelStudentToPb(*student)
		addedStudents = append(addedStudents, pbStudent)
	}

	return addedStudents, nil
}

func GetStudentsFromDb(ctx context.Context, sortOptions primitive.D, filter primitive.M, pageNumber, pageSize uint32) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	defer client.Disconnect(ctx)

	col1 := client.Database("school").Collection("students")
	findOptions := options.Find()
	findOptions.SetSkip(int64(pageNumber-1) * int64(pageSize))
	findOptions.SetLimit(int64(pageSize))

	if len(sortOptions) > 0 {
		findOptions.SetSort(sortOptions)
	}
	cursor, err := col1.Find(ctx, filter, findOptions)

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	defer cursor.Close(ctx)

	students, err := decodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student { return &models.Student{} })
	if err != nil {
		return nil, err
	}

	return students, nil
}

func ModifyStudentInDb(ctx context.Context, pbStudent []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}
	defer client.Disconnect(ctx)

	var updatedStudents []*pb.Student
	for _, student := range pbStudent {
		if student.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id can't be blank"), "id can't be blank")
		}

		modelStudent := mapPbStudentToModelStudent(student)
		objectId, err := primitive.ObjectIDFromHex(student.Id)

		if err != nil {
			return nil, utils.ErrorHandler(err, "Imvalid ID")
		}

		modelDoc, err := bson.Marshal(modelStudent)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}

		var updatedDoc bson.M
		err = bson.Unmarshal(modelDoc, &updatedDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}

		delete(updatedDoc, "_id")
		_, err = client.Database("school").Collection("students").UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{"$set": updatedDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating student id:", student.Id))
		}

		updatdedStudent := mapModelStudentToPb(*modelStudent)
		updatedStudents = append(updatedStudents, updatdedStudent)
	}

	return updatedStudents, nil
}
