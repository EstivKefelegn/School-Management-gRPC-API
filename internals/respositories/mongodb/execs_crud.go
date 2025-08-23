package mongodb

import (
	"context"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddExecsToDb(ctx context.Context, execsFromReq []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal Error")
	}

	defer client.Disconnect(ctx)

	// newTeachers := make([]*models.Teacher, len(teachersFromReq))
	newExecs := make([]*models.Exec, len(execsFromReq))
	for i, pbExec := range execsFromReq {
		newExecs[i] = mapPbExecToModelTeacher(pbExec)
		hashPassword, err := utils.HashPassword(newExecs[i].Password)	
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}
		newExecs[i].Password = hashPassword
		currentTime := time.Now().Format(time.RFC3339)
		newExecs[i].UserCreatedAt = currentTime
		newExecs[i].InactiveStatus = false
		// newTeachers[i] = &modelTeacher
	}
	// Inserting a value
	var addedExecs []*pb.Exec
	for _, exec := range newExecs {
		result, err := client.Database("school").Collection("execs").InsertOne(ctx, exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error adding value to database")
		}

		// insertedID is interface type so we will convert it to string HEX will convert that to string
		objectID, ok := result.InsertedID.(primitive.ObjectID)
		if ok {
			exec.Id = objectID.Hex()
		}		

		// The return the values to the user
		pbExec := mapModelExecToPb(*exec)
		addedExecs = append(addedExecs, pbExec)
	}
	return addedExecs, nil
}
