package mongodb

import (
	"context"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddStudentsToDb(ctx context.Context, studentsFromReq []*pb.Student) ([]*pb.Student, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Couldn't connect to db")
	}
	
	defer client.Disconnect(ctx)

	newStudents := make([]*models.Student, len(studentsFromReq))
	for i, student := range studentsFromReq {
		newStudents[i] = mapPbStudentToModelTeacher(student)
	}


	var addedStudents []*pb.Student
	for  _, student := range newStudents {
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
