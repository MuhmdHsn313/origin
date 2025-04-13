package repository

import (
	"github.com/MuhmdHsn313/origin/orm"
	"reflect"
)

// HasContents returns true if the model has a "Contents" field that is a slice and whose element type implements orm.IContentModel.
func hasContents(model any) bool {
	// Get the reflect.Type of the model.
	typ := reflect.TypeOf(model)
	// If the model is a pointer, get its element type.
	if typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		typ = typ.Elem()
	}

	// Iterate over the fields of the struct.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// Check if the field's name is "Contents".
		if field.Name == "Contents" && field.Type.Kind() == reflect.Slice {
			// Get the interface type for IContentModel.
			iContentModelType := reflect.TypeOf((*orm.IContentModel)(nil)).Elem()
			// Check if the element type implements IContentModel.
			if field.Type.Elem().Implements(iContentModelType) {
				return true
			}
		}
	}
	return false
}
