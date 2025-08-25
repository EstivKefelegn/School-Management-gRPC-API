package mongodb

import (
	"context"
	"errors"
	"fmt"
	"schoolmanagementGRPC/internals/models"
	"schoolmanagementGRPC/pkg/utils"
	pb "schoolmanagementGRPC/proto/gen"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		newExecs[i] = mapPbExecToModelExec(pbExec)
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

func GetExecsFromDb(ctx context.Context, sortOptions primitive.D, filter primitive.M) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, err
	}

	defer client.Disconnect(ctx)

	col1 := client.Database("school").Collection("execs")

	var cursor *mongo.Cursor
	if len(sortOptions) < 1 {
		cursor, err = col1.Find(ctx, filter)
	} else {
		cursor, err = col1.Find(ctx, filter, options.Find().SetSort(sortOptions))
	}
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	execs, err := decodeEntities(ctx, cursor, func() *pb.Exec { return &pb.Exec{} }, func() *models.Exec { return &models.Exec{} })
	if err != nil {
		return nil, err
	}
	return execs, nil
}

func ModifyExecsInDb(ctx context.Context, pbExec []*pb.Exec) ([]*pb.Exec, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}
	defer client.Disconnect(ctx)

	var updatedExecs []*pb.Exec
	for _, exec := range pbExec {
		if exec.Id == "" {
			return nil, utils.ErrorHandler(errors.New("id cannot be blank"), "Id cannot be blank")
		}

		modelExec := mapPbExecToModelExec(exec)
		objectId, err := primitive.ObjectIDFromHex(exec.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid Id")
		}

		modelDoc, err := bson.Marshal(modelExec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}

		var updatedDoc bson.M
		err = bson.Unmarshal(modelDoc, &updatedDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}

		delete(updatedDoc, "_id")
		_, err = client.Database("school").Collection("execs").UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{"$set": updatedDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating exec id:", exec.Id))
		}
		updatedExec := mapModelExecToPb(*modelExec)

		updatedExecs = append(updatedExecs, updatedExec)
	}

	return updatedExecs, nil
}
