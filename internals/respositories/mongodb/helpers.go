package mongodb

import (
	"context"
	"reflect"
	"schoolmanagementGRPC/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

// The functions passed as a parameter's are a functions which returns a model like teacher model nothing more than that
func decodeEntities[T any, M any](ctx context.Context, cursor *mongo.Cursor, newEntity func() *T, newModel func() *M) ([]*T, error) {
	var entities []*T
	for cursor.Next(ctx) {
		model := newModel()
		// var teacher models.Teacher
		err := cursor.Decode(&model)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Error")
		}
		entity := newEntity()
		modelVal := reflect.ValueOf(model).Elem()
		pbVal := reflect.ValueOf(entity).Elem()

		for i := 0; i < modelVal.NumField(); i++ {
			modelField := modelVal.Field(i)
			modelFieldName := modelVal.Type().Field(i).Name

			pbField := pbVal.FieldByName(modelFieldName)
			if pbField.IsValid() && pbField.CanSet() {
				pbField.Set(modelField)
			}
		}
		entities = append(entities, entity)
		// teachers = append(teachers, &pb.Teacher{
		// 	Id:        teacher.Id,
		// 	FirstName: teacher.FirstName,
		// 	LastName:  teacher.LastName,
		// 	Email:     teacher.Email,
		// 	Class:     teacher.Class,
		// 	Subject:   teacher.Subject,
		// })
	}
	err := cursor.Err()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Internar server error")
	}
	return entities, nil
}
