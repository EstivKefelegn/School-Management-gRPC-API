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
		pbTeacher := mapModelTeacherToPb(*teacher)
		addedTeachers = append(addedTeachers, pbTeacher)
	}
	return addedTeachers, nil
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

	return decodeEntities(ctx, cursor, func() *pb.Teacher { return &pb.Teacher{} }, func() *models.Teacher { return &models.Teacher{} })
}

func ModifyTeachersfromDb(ctx context.Context, req *pb.Teachers) ([]*pb.Teacher, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	defer client.Disconnect(ctx)

	var updatedTeachers []*pb.Teacher
	for _, teacher := range req.Teachers {
		if teacher.Id == "" {
			return nil, utils.ErrorHandler(errors.New("Id can't be blank"), "Id can't be blank")
		}

		modelTeacher := mapPbTeacherToModelTeacher(teacher)

		objId, err := primitive.ObjectIDFromHex(teacher.Id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Invalid Id")
		}

		// Convert modelTeacher to BSON document
		modelDoc, err := bson.Marshal(modelTeacher)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		var updateDoc bson.M
		err = bson.Unmarshal(modelDoc, &updateDoc)
		if err != nil {
			return nil, utils.ErrorHandler(err, "internal error")
		}

		// Remove the _id field from the update document
		delete(updateDoc, "_id")

		_, err = client.Database("school").Collection("teachers").UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": updateDoc})
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("error updating teacher id: %s", teacher.Id))
		}

		updatedTeacher := mapModelTeacherToPb(*modelTeacher)
		updatedTeachers = append(updatedTeachers, updatedTeacher)
	}
	return updatedTeachers, nil
}

func DeleteTeachersFromDb(ctx context.Context, teacherIdsToDelete []string) ([]string, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	defer client.Disconnect(ctx)
	objectIds := make([]primitive.ObjectID, len(teacherIdsToDelete))
	for i, id := range teacherIdsToDelete {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, utils.ErrorHandler(err, fmt.Sprintf("incorrect id: %v", id))
		}
		objectIds[i] = objectId
	}

	filter := bson.M{"_id": bson.M{"$in": objectIds}}
	result, err := client.Database("school").Collection("teachers").DeleteMany(ctx, filter)
	if err != nil {
		return nil, utils.ErrorHandler(err, "internal error")
	}

	if result.DeletedCount == 0 {
		return nil, utils.ErrorHandler(err, "no teachers were deleted")
	}

	deletedIds := make([]string, result.DeletedCount)
	for i, id := range objectIds {
		deletedIds[i] = id.Hex()
	}
	return deletedIds, nil
}

func GetStudentsByTeacherId(ctx context.Context, teacherId string) ([]*pb.Student, error) {

	client, err := CreateMongoClient()

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	defer client.Disconnect(ctx)
	objectId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Invalid Teacher Id")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objectId}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, utils.ErrorHandler(err, "teacher not found")
		}
		return nil, utils.ErrorHandler(err, "Internal Error")
	}

	cursor, err := client.Database("school").Collection("students").Find(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	defer cursor.Close(ctx)
	students, err := decodeEntities(ctx, cursor, func() *pb.Student { return &pb.Student{} }, func() *models.Student { return &models.Student{} })

	if err != nil {
		return nil, utils.ErrorHandler(err, "Internal error")
	}

	return students, nil
}


func GetStudentCountByTeacherIdFromDb(ctx context.Context, teacherId string) (int64, error) {
	client, err := CreateMongoClient()
	if err != nil {
		return 0, utils.ErrorHandler(err, "internal error")
	}
	defer client.Disconnect(ctx)

	objId, err := primitive.ObjectIDFromHex(teacherId)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Invalid Teacher Id")
	}

	var teacher models.Teacher
	err = client.Database("school").Collection("teachers").FindOne(ctx, bson.M{"_id": objId}).Decode(&teacher)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, utils.ErrorHandler(err, "teacher not found")
		}
		return 0, utils.ErrorHandler(err, "internal error")
	}

	count, err := client.Database("school").Collection("students").CountDocuments(ctx, bson.M{"class": teacher.Class})
	if err != nil {
		return 0, utils.ErrorHandler(err, "internal error")
	}

	return count, nil
}