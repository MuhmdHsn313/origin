package service

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"reflect"
	"unicode"
)

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// fillStruct recursively copies values from src to dst. Both must be structs.
func fillStruct(dst, src reflect.Value) error {
	if dst.Kind() != reflect.Struct || src.Kind() != reflect.Struct {
		return fmt.Errorf("fillStruct requires both dst and src to be structs")
	}

	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		srcFieldVal := src.Field(i)

		dstField := dst.FieldByName(srcField.Name)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		if srcFieldVal.Kind() == reflect.Struct {
			// Recursively fill nested structs.
			if err := fillStruct(dstField, srcFieldVal); err != nil {
				return err
			}
		} else if srcFieldVal.Kind() == reflect.Slice {
			// Handle nested slice of structs.
			newSlice := reflect.MakeSlice(dstField.Type(), srcFieldVal.Len(), srcFieldVal.Len())
			for j := 0; j < srcFieldVal.Len(); j++ {
				sElem := srcFieldVal.Index(j)
				dElem := newSlice.Index(j)
				if sElem.Kind() == reflect.Struct {
					if err := fillStruct(dElem, sElem); err != nil {
						return err
					}
				} else if sElem.Type().AssignableTo(dElem.Type()) {
					dElem.Set(sElem)
				} else if sElem.Type().ConvertibleTo(dElem.Type()) {
					dElem.Set(sElem.Convert(dElem.Type()))
				} else {
					return fmt.Errorf("type mismatch in nested slice field %s", srcField.Name)
				}
			}
			dstField.Set(newSlice)
		} else {
			// For simple types, assign or convert.
			if srcFieldVal.Type().AssignableTo(dstField.Type()) {
				dstField.Set(srcFieldVal)
			} else if srcFieldVal.Type().ConvertibleTo(dstField.Type()) {
				dstField.Set(srcFieldVal.Convert(dstField.Type()))
			} else {
				return fmt.Errorf("cannot assign field %s: type mismatch (%s vs %s)", srcField.Name, srcFieldVal.Type(), dstField.Type())
			}
		}
	}
	return nil
}

//// Helper function to copy values between compatible fields
//func copyField(dst, src reflect.Value) error {
//	if src.Kind() == reflect.Ptr {
//		if src.IsNil() {
//			return nil // No action for nil pointers
//		}
//		src = src.Elem()
//	}
//
//	if dst.Kind() == reflect.Ptr {
//		if dst.IsNil() {
//			dst.Set(reflect.New(dst.Type().Elem()))
//		}
//		dst = dst.Elem()
//	}
//
//	if src.Type().AssignableTo(dst.Type()) {
//		dst.Set(src)
//	} else if src.Type().ConvertibleTo(dst.Type()) {
//		dst.Set(src.Convert(dst.Type()))
//	} else {
//		return fmt.Errorf("type mismatch: %s vs %s", src.Type(), dst.Type())
//	}
//	return nil
//}

// copyField handles type conversions and pointer dereferencing
func copyField(dst, src reflect.Value) error {
	// Dereference pointers
	for src.Kind() == reflect.Ptr && !src.IsNil() {
		src = src.Elem()
	}
	for dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	// Assign or convert values
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
	} else if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
	} else {
		return fmt.Errorf("type mismatch: %s → %s", src.Type(), dst.Type())
	}
	return nil
}

// copyStruct copies fields between two structs by name
func copyStruct(dst, src reflect.Value) error {
	if dst.Kind() != reflect.Struct || src.Kind() != reflect.Struct {
		return fmt.Errorf("both values must be structs")
	}

	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		srcVal := src.Field(i)
		dstField := dst.FieldByName(srcField.Name)

		if dstField.IsValid() && dstField.CanSet() {
			if err := copyField(dstField, srcVal); err != nil {
				return fmt.Errorf("field %s: %w", srcField.Name, err)
			}
		}
	}
	return nil
}

func handleSliceUpdate(modelField, paramValue reflect.Value) error {
	// Get concrete element type from model field
	elemType := modelField.Type().Elem()

	// Extract new contents from parameters
	newContents := reflect.MakeSlice(reflect.SliceOf(elemType), 0, paramValue.Len())
	for i := 0; i < paramValue.Len(); i++ {
		srcElem := paramValue.Index(i)
		if srcElem.Kind() == reflect.Ptr {
			srcElem = srcElem.Elem()
		}

		dstElem := reflect.New(elemType).Elem()
		if err := copyStruct(dstElem, srcElem); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
		newContents = reflect.Append(newContents, dstElem)
	}

	// Merge with existing contents using reflection
	merged := mergeContents(modelField, newContents)
	modelField.Set(merged)
	return nil
}

func mergeContents(existing, newContents reflect.Value) reflect.Value {
	contentMap := make(map[string]reflect.Value)

	// Add existing contents
	for i := 0; i < existing.Len(); i++ {
		elem := existing.Index(i)
		langID := getLanguageID(elem)
		contentMap[langID] = elem
	}

	// Add/override with new contents
	for i := 0; i < newContents.Len(); i++ {
		elem := newContents.Index(i)
		langID := getLanguageID(elem)
		contentMap[langID] = elem
	}

	// Create merged slice
	merged := reflect.MakeSlice(existing.Type(), 0, len(contentMap))
	for _, v := range contentMap {
		merged = reflect.Append(merged, v)
	}

	return merged
}

//	func handleContentUpdate(modelField, paramValue reflect.Value) error {
//		if paramValue.Kind() == reflect.Ptr {
//			paramValue = paramValue.Elem()
//		}
//
//		// Create map of existing items by LanguageID
//		existingMap := make(map[string]reflect.Value)
//		existing := modelField.Interface()
//		existingVal := reflect.ValueOf(existing)
//
//		for i := 0; i < existingVal.Len(); i++ {
//			item := existingVal.Index(i)
//			langID := getLanguageID(item)
//			if langID != "" {
//				existingMap[langID] = item
//			}
//		}
//
//		// Process update items
//		for i := 0; i < paramValue.Len(); i++ {
//			updateItem := paramValue.Index(i)
//			if updateItem.Kind() == reflect.Ptr {
//				updateItem = updateItem.Elem()
//			}
//
//			langID := getLanguageID(updateItem)
//			if langID == "" {
//				continue
//			}
//
//			// Update existing item or create new
//			if existingItem, exists := existingMap[langID]; exists {
//				if err := copyStruct(existingItem, updateItem); err != nil {
//					return fmt.Errorf("failed to update %s: %w", langID, err)
//				}
//			} else {
//				newItem := reflect.New(modelField.Type().Elem()).Elem()
//				if err := copyStruct(newItem, updateItem); err != nil {
//					return fmt.Errorf("failed to create %s: %w", langID, err)
//				}
//				existingMap[langID] = newItem
//			}
//		}
//
//		// Convert map back to slice
//		newSlice := reflect.MakeSlice(modelField.Type(), 0, len(existingMap))
//		for _, v := range existingMap {
//			newSlice = reflect.Append(newSlice, v)
//		}
//
//		modelField.Set(newSlice)
//		return nil
//	}
func handleContentUpdate(modelField, paramValue reflect.Value) error {
	// Always create new map for merging
	contentMap := make(map[string]reflect.Value)

	// 1. Populate with existing content
	for i := 0; i < modelField.Len(); i++ {
		item := modelField.Index(i)
		langID := getLanguageID(item)
		contentMap[langID] = item
	}

	// 2. Process updates
	for i := 0; i < paramValue.Len(); i++ {
		updateItem := paramValue.Index(i)
		langID := getLanguageID(updateItem)

		// Create new instance of the content type
		newItem := reflect.New(modelField.Type().Elem()).Elem()
		if err := copyStruct(newItem, updateItem); err == nil {
			contentMap[langID] = newItem
		}
	}

	// 3. Rebuild slice
	newSlice := reflect.MakeSlice(modelField.Type(), 0, len(contentMap))
	for _, v := range contentMap {
		newSlice = reflect.Append(newSlice, v)
	}
	modelField.Set(newSlice)

	return nil
}

// Enhanced getLanguageID to handle both value and pointer receivers
func getLanguageID(v reflect.Value) string {
	method := v.MethodByName("GetLanguageID")
	if !method.IsValid() && v.CanAddr() {
		method = v.Addr().MethodByName("GetLanguageID")
	}
	if !method.IsValid() {
		field := v.FieldByName("LanguageID")
		if field.IsValid() && field.CanAddr() {
			return field.String()
		}
		return ""
	}
	results := method.Call(nil)
	if len(results) == 0 {
		return ""
	}
	return results[0].String()
}

// structNameToSnake takes any struct instance and returns the snake_case version of its type name.
// It uses reflection to handle both value and pointer types.
func structNameToSnake(i interface{}) string {
	// Get the type information using reflection.
	t := reflect.TypeOf(i)
	// If the type is a pointer, retrieve the element type.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check if the type is a struct.
	if t.Kind() == reflect.Struct {
		// Convert the type name to snake_case.
		return strcase.ToSnake(t.Name())
	}
	// Return an empty string if the input is not a struct.
	return ""
}
