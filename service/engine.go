package service

import (
	"fmt"
	"github.com/MuhmdHsn313/origin/orm"
	"reflect"
	"strings"
)

type Engine[T any] interface {
	GenerateCreateParameters() (interface{}, error)
	GenerateUpdateParameters() (interface{}, error)
	GenerateFilterParameters() (interface{}, error)
	FillModelFromCreateParameters(createParams interface{}) (*T, error)
	UpdateModelFromUpdateParameters(model *T, updateParams interface{}) (*T, error)
}

type engine[T any] struct {
}

func CreateEngine[M any]() Engine[M] {
	return &engine[M]{}
}

// GenerateCreateParameters generates a new struct type for creating a model, excluding fields from base models.
func (e engine[T]) GenerateCreateParameters() (interface{}, error) {
	var model T
	// Get the reflection type of the model
	modelType := reflect.TypeOf(model)

	// If the model is a pointer, dereference it to get the actual struct type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Create a new struct that will hold the parameters
	var fields []reflect.StructField
	// A set to track added field names and avoid duplicates
	addedFields := make(map[string]bool)

	// Iterate over all the fields of the model struct
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Skip fields from base structs (like orm.Model) by name or type
		if field.Anonymous || e.isBaseField(field) {
			// Check if the field is an embedded struct (like orm.Model or orm.ContentModel)
			if field.Anonymous && e.isBaseField(field) {
				// Handle embedded structs (e.g., ContentModel)
				embeddedFields, err := e.extractBaseEmbeddedFields(field.Type, addedFields)
				if err != nil {
					return nil, err
				}
				fields = append(fields, embeddedFields...)
			}
			continue
		}

		// For slice fields (e.g., []BlocContent), we need to handle them specifically
		if field.Type.Kind() == reflect.Slice {
			// If the slice is of structs, we need to extract the relevant fields from the struct
			if field.Type.Elem().Kind() == reflect.Struct {
				// Extract fields from the slice's struct (e.g., BlocContent)
				innerFields, err := e.generateInnerStruct(field.Type.Elem(), addedFields, true)
				if err != nil {
					return nil, err
				}

				// Add a new field of the struct type
				jsonTag, ok := field.Tag.Lookup("json")
				if !ok {
					jsonTag = field.Name
				}
				validationTag, isValidationExist := field.Tag.Lookup("validate")

				var tag string
				if isValidationExist {
					tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
				} else {
					tag = fmt.Sprintf(`json:"%s"`, jsonTag)
				}

				fields = append(fields, reflect.StructField{
					Name:      field.Name,
					Type:      reflect.SliceOf(innerFields),
					Tag:       reflect.StructTag(tag),
					Anonymous: false,
				})
			}
		} else {
			// For other fields, just add them to the parameters struct
			jsonTag, ok := field.Tag.Lookup("json")
			if !ok {
				jsonTag = field.Name
			}
			validationTag, isValidationExist := field.Tag.Lookup("validate")

			var tag string
			if isValidationExist {
				tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
			} else {
				tag = fmt.Sprintf(`json:"%s"`, jsonTag)
			}

			fields = append(fields, reflect.StructField{
				Name:      field.Name,
				Type:      field.Type,
				Tag:       reflect.StructTag(tag),
				Anonymous: false,
			})
		}
	}

	// Create a new struct type with the extracted fields
	paramStruct := reflect.StructOf(fields)

	// Return a new instance of the generated struct
	return reflect.New(paramStruct).Interface(), nil
}

// GenerateUpdateParameters generates a new struct type for updating a model, excluding fields from base models.
func (e engine[T]) GenerateUpdateParameters() (interface{}, error) {
	var model T
	// Get the reflection type of the model
	modelType := reflect.TypeOf(model)

	// If the model is a pointer, dereference it to get the actual struct type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Create a new struct that will hold the parameters
	var fields []reflect.StructField
	// A set to track added field names and avoid duplicates
	addedFields := make(map[string]bool)

	// Iterate over all the fields of the model struct
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Skip fields from base structs (like orm.Model) by name or type
		if field.Anonymous || e.isBaseField(field) {
			// Check if the field is an embedded struct (like orm.Model or orm.ContentModel)
			if field.Anonymous && e.isBaseField(field) {
				// Handle embedded structs (e.g., ContentModel)
				embeddedFields, err := e.extractBaseEmbeddedFields(field.Type, addedFields)
				if err != nil {
					return nil, err
				}
				fields = append(fields, embeddedFields...)
			}
			continue
		}

		// For slice fields (e.g., []BlocContent), we need to handle them specifically
		if field.Type.Kind() == reflect.Slice {
			// If the slice is of structs, we need to extract the relevant fields from the struct
			if field.Type.Elem().Kind() == reflect.Struct {
				// Extract fields from the slice's struct (e.g., BlocContent)
				innerFields, err := e.generateInnerStruct(field.Type.Elem(), addedFields, true)
				if err != nil {
					return nil, err
				}

				// Add a new field of the struct type
				jsonTag, ok := field.Tag.Lookup("json")
				if !ok {
					jsonTag = field.Name
				}
				validationTag, isValidationExist := field.Tag.Lookup("validate")

				var tag string
				if isValidationExist {
					tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
				} else {
					tag = fmt.Sprintf(`json:"%s"`, jsonTag)
				}

				fields = append(fields, reflect.StructField{
					Name:      field.Name,
					Type:      reflect.PointerTo(reflect.SliceOf(innerFields)),
					Tag:       reflect.StructTag(tag),
					Anonymous: false,
				})
			}
		} else {
			// For other fields, just add them to the parameters struct
			jsonTag, ok := field.Tag.Lookup("json")
			if !ok {
				jsonTag = field.Name
			}
			validationTag, isValidationExist := field.Tag.Lookup("validate")

			var tag string
			if isValidationExist {
				tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
			} else {
				tag = fmt.Sprintf(`json:"%s"`, jsonTag)
			}

			fields = append(fields, reflect.StructField{
				Name:      field.Name,
				Type:      reflect.PointerTo(field.Type),
				Tag:       reflect.StructTag(tag),
				Anonymous: false,
			})
		}
	}

	// Create a new struct type with the extracted fields
	paramStruct := reflect.StructOf(fields)

	// Return a new instance of the generated struct
	return reflect.New(paramStruct).Interface(), nil
}

// GenerateFilterParameters generates a new struct type for filtering a model.
// It flattens the main model's fields and, for content model slices, extracts
// the inner struct fields (e.g. "Content", "LanguageID") as top-level filter parameters.
// All fields are pointers and use `url:"..."` tags.
func (e engine[T]) GenerateFilterParameters() (interface{}, error) {
	var model T
	// Get the reflection type of the model.
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	var fields []reflect.StructField
	// A set to track added field names and avoid duplicates.
	addedFields := make(map[string]bool)

	// Iterate over all the fields of the model struct.
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Skip fields from base structs (like orm.Model) by name or type.
		if field.Anonymous || e.isBaseField(field) {
			continue
		}

		// If the field is a slice and its element is a struct, check if it is a content model.
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			// Check if the slice element implements the IContentModel interface.
			if orm.IsContentModel(reflect.New(field.Type.Elem()).Elem().Interface()) {
				innerType := field.Type.Elem()
				// Iterate over the inner fields and flatten them.
				for j := 0; j < innerType.NumField(); j++ {
					innerField := innerType.Field(j)
					if strings.HasSuffix(innerField.Name, "ID") {
						continue
					}
					// For embedded base fields (like orm.ContentModel), extract only LanguageID.
					if innerField.Anonymous || e.isBaseField(innerField) {
						if innerField.Name == "LanguageID" && !addedFields[innerField.Name] {
							tag := fmt.Sprintf(`url:"%s"`, toSnakeCase(innerField.Name))
							fields = append(fields, reflect.StructField{
								Name:      innerField.Name,
								Type:      reflect.PtrTo(innerField.Type),
								Tag:       reflect.StructTag(tag),
								Anonymous: false,
							})
							addedFields[innerField.Name] = true
						}
						continue
					}
					// For other inner fields, add them if not already added.
					if !addedFields[innerField.Name] {
						tag := fmt.Sprintf(`url:"%s"`, toSnakeCase(innerField.Name))
						fields = append(fields, reflect.StructField{
							Name:      innerField.Name,
							Type:      reflect.PtrTo(innerField.Type),
							Tag:       reflect.StructTag(tag),
							Anonymous: false,
						})
						addedFields[innerField.Name] = true
					}
				}
				// Skip adding the whole slice field since we've flattened its inner fields.
				continue
			}
		}

		// For non-slice fields, add them as pointer types with a URL tag.
		if !addedFields[field.Name] {
			tag := fmt.Sprintf(`url:"%s"`, toSnakeCase(field.Name))
			fields = append(fields, reflect.StructField{
				Name:      field.Name,
				Type:      reflect.PtrTo(field.Type),
				Tag:       reflect.StructTag(tag),
				Anonymous: false,
			})
			addedFields[field.Name] = true
		}
	}

	// Create a new struct type with the collected fields.
	paramStruct := reflect.StructOf(fields)
	// Return a new instance of the generated struct.
	return reflect.New(paramStruct).Interface(), nil
}

// Helper function to generate inner structs (like BlocContent)
func (e engine[T]) generateInnerStruct(innerType reflect.Type, addedFields map[string]bool, includeForeignKeys bool) (reflect.Type, error) {
	var innerFields []reflect.StructField

	// Iterate over the fields of the inner struct
	for i := 0; i < innerType.NumField(); i++ {
		field := innerType.Field(i)

		// Skip fields from base structs (like orm.ContentModel)
		if field.Anonymous && e.isBaseField(field) {
			// Handle embedded structs (e.g., orm.ContentModel)
			if field.Anonymous && e.isBaseField(field) {
				embeddedFields, err := e.extractBaseEmbeddedFields(field.Type, addedFields)
				if err != nil {
					return nil, err
				}
				innerFields = append(innerFields, embeddedFields...)
			}
			continue
		}

		// If the field is already added (i.e., LanguageID), skip it
		if addedFields[field.Name] {
			continue
		}

		// Skip foreign keys if includeForeignKeys is true
		if includeForeignKeys && strings.HasSuffix(field.Name, "ID") {
			continue
		}

		// Add the field to the inner struct
		jsonTag, ok := field.Tag.Lookup("json")
		if !ok {
			jsonTag = field.Name
		}
		validationTag, isValidationExist := field.Tag.Lookup("validate")

		var tag string
		if isValidationExist {
			tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
		} else {
			tag = fmt.Sprintf(`json:"%s"`, jsonTag)
		}

		innerFields = append(innerFields, reflect.StructField{
			Name:      field.Name,
			Type:      field.Type,
			Tag:       reflect.StructTag(tag),
			Anonymous: false,
		})

		// Mark the field as added
		addedFields[field.Name] = true
	}

	// Create and return the inner struct type
	return reflect.StructOf(innerFields), nil
}

// Check if the field belongs to a base model (like orm.Model or orm.ContentModel)
func (e engine[T]) isBaseField(field reflect.StructField) bool {
	// For simplicity, check by field name or type
	// This can be extended to check by type name or a specific struct tag, etc.
	baseTypes := []string{"Model", "ContentModel"}

	for _, base := range baseTypes {
		if strings.Contains(field.Name, base) || field.Type.Name() == base {
			return true
		}
	}
	return false
}

// Extract fields from embedded structs (like ContentModel) to ensure LanguageID is included
func (e engine[T]) extractBaseEmbeddedFields(embeddedType reflect.Type, addedFields map[string]bool) ([]reflect.StructField, error) {
	var fields []reflect.StructField

	// Iterate over the fields of the embedded struct
	for i := 0; i < embeddedType.NumField(); i++ {
		field := embeddedType.Field(i)

		// If LanguageID is present, ensure it's added only once
		if field.Name == "LanguageID" && !addedFields[field.Name] {
			jsonTag, ok := field.Tag.Lookup("json")
			if !ok {
				jsonTag = field.Name
			}
			validationTag, isValidationExist := field.Tag.Lookup("validate")

			var tag string
			if isValidationExist {
				tag = fmt.Sprintf(`json:"%s" validate:"%s"`, jsonTag, validationTag)
			} else {
				tag = fmt.Sprintf(`json:"%s"`, jsonTag)
			}

			fields = append(fields, reflect.StructField{
				Name:      field.Name,
				Type:      field.Type,
				Tag:       reflect.StructTag(tag),
				Anonymous: false,
			})
			// Mark LanguageID as added
			addedFields[field.Name] = true
		}
	}

	return fields, nil
}

// FillModelFromCreateParameters creates and populates a model instance from create parameters
func (e engine[T]) FillModelFromCreateParameters(createParams interface{}) (*T, error) {
	modelType := reflect.TypeOf((*T)(nil)).Elem()
	modelVal := reflect.New(modelType) // *T
	modelElem := modelVal.Elem()       // T

	cpVal := reflect.ValueOf(createParams)
	if cpVal.Kind() == reflect.Ptr {
		cpVal = cpVal.Elem()
	}

	for i := 0; i < cpVal.NumField(); i++ {
		cpField := cpVal.Type().Field(i)
		cpFieldVal := cpVal.Field(i)

		modelField := modelElem.FieldByName(cpField.Name)
		if !modelField.IsValid() {
			continue // Skip missing fields
		}

		if !modelField.CanSet() {
			return nil, fmt.Errorf("model field %s cannot be set", cpField.Name)
		}

		// Handle slice fields
		if cpFieldVal.Kind() == reflect.Slice {
			newSlice := reflect.MakeSlice(modelField.Type(), cpFieldVal.Len(), cpFieldVal.Len())
			for j := 0; j < cpFieldVal.Len(); j++ {
				srcElem := cpFieldVal.Index(j)
				dstElem := newSlice.Index(j)

				// Dereference pointers to structs
				if srcElem.Kind() == reflect.Ptr {
					if srcElem.IsNil() {
						continue // Skip nil pointers
					}
					srcElem = srcElem.Elem()
				}

				if srcElem.Kind() == reflect.Struct && dstElem.Kind() == reflect.Struct {
					if err := copyStruct(dstElem, srcElem); err != nil {
						return nil, fmt.Errorf("%s[%d]: %w", cpField.Name, j, err)
					}
				} else if err := copyField(dstElem, srcElem); err != nil {
					return nil, fmt.Errorf("%s[%d]: %w", cpField.Name, j, err)
				}
			}
			modelField.Set(newSlice)
		} else {
			if err := copyField(modelField, cpFieldVal); err != nil {
				return nil, fmt.Errorf("%s: %w", cpField.Name, err)
			}
		}
	}

	return modelVal.Interface().(*T), nil
}

//// FillModelFromCreateParameters1 fills the model instance with values from the createParams instance.
//// model should be a pointer to the target struct, and createParams is a pointer to the create parameters struct.
//func (e engine[T]) FillModelFromCreateParameters1(createParams interface{}) (*T, error) {
//	var model *T
//	// Get reflect.Value of model (dereferenced) and createParams (dereferenced)
//	modelVal := reflect.ValueOf(model).Elem()
//	cpVal := reflect.ValueOf(createParams).Elem()
//
//	// Iterate over each field in the create parameters struct.
//	cpType := cpVal.Type()
//	for i := 0; i < cpType.NumField(); i++ {
//		cpField := cpType.Field(i)
//		cpFieldVal := cpVal.Field(i)
//
//		// Look for the same field name in the model.
//		modelField := modelVal.FieldByName(cpField.Name)
//		if !modelField.IsValid() || !modelField.CanSet() {
//			// Field not present or not settable; skip.
//			continue
//		}
//
//		// Handle slice fields (e.g., contents)
//		if cpFieldVal.Kind() == reflect.Slice {
//			// Create a new slice for the model field with the same length.
//			newSlice := reflect.MakeSlice(modelField.Type(), cpFieldVal.Len(), cpFieldVal.Len())
//			for j := 0; j < cpFieldVal.Len(); j++ {
//				srcElem := cpFieldVal.Index(j)
//				dstElem := newSlice.Index(j)
//				// If the element is a struct, do a recursive fill.
//				if srcElem.Kind() == reflect.Struct && dstElem.Kind() == reflect.Struct {
//					if err := fillStruct(dstElem, srcElem); err != nil {
//						return nil, fmt.Errorf("failed to fill nested struct in field %s: %w", cpField.Name, err)
//					}
//				} else if srcElem.Type().AssignableTo(dstElem.Type()) {
//					dstElem.Set(srcElem)
//				} else {
//					return nil, fmt.Errorf("type mismatch in slice field %s", cpField.Name)
//				}
//			}
//			modelField.Set(newSlice)
//		} else {
//			// For non-slice fields, assign directly if types are assignable.
//			if cpFieldVal.Type().AssignableTo(modelField.Type()) {
//				modelField.Set(cpFieldVal)
//			} else if cpFieldVal.Type().ConvertibleTo(modelField.Type()) {
//				modelField.Set(cpFieldVal.Convert(modelField.Type()))
//			} else {
//				return nil, fmt.Errorf("cannot assign value of field %s: type mismatch (%s vs %s)", cpField.Name, cpFieldVal.Type(), modelField.Type())
//			}
//		}
//	}
//
//	return model, nil
//}

// UpdateModelFromUpdateParameters updates the model and returns the modified instance
func (e engine[T]) UpdateModelFromUpdateParameters(model *T, updateParams interface{}) (*T, error) {
	modelVal := reflect.ValueOf(model).Elem()
	paramsVal := reflect.ValueOf(updateParams)

	if paramsVal.Kind() == reflect.Ptr {
		paramsVal = paramsVal.Elem()
	}

	for i := 0; i < paramsVal.NumField(); i++ {
		paramField := paramsVal.Type().Field(i)
		paramValue := paramsVal.Field(i)

		// Skip nil pointers
		if paramValue.Kind() == reflect.Ptr && paramValue.IsNil() {
			continue
		}

		modelField := modelVal.FieldByName(paramField.Name)
		if !modelField.IsValid() || !modelField.CanSet() {
			continue // Skip non-existing fields
		}

		switch {
		case modelField.Kind() == reflect.Slice:
			var sliceValue reflect.Value
			if paramValue.Kind() == reflect.Ptr {
				sliceValue = paramValue.Elem()
			} else {
				sliceValue = paramValue
			}

			if paramField.Name == "Contents" {
				if err := handleContentUpdate(modelField, sliceValue); err != nil {
					return model, fmt.Errorf("field %s: %w", paramField.Name, err)
				}
			}

		case paramValue.Kind() == reflect.Ptr:
			// Handle pointer parameters
			if err := copyField(modelField, paramValue.Elem()); err != nil {
				return model, fmt.Errorf("field %s: %w", paramField.Name, err)
			}

		default:
			// Handle direct value parameters
			if err := copyField(modelField, paramValue); err != nil {
				return model, fmt.Errorf("field %s: %w", paramField.Name, err)
			}
		}
	}

	return model, nil
}

// UpdateModelFromUpdateParameters updates the given model instance with the non-nil values provided in updateParams.
// - model: A pointer to the target model (for example, *Blog).
// - updateParams: A pointer to the update parameters struct (with fields as pointers).
//
// For simple fields, if the corresponding update parameter is non-nil, its value is set on the model.
// For slice fields which represent content models, the function:
//
//	a) Extracts the new contents using ExtractContent,
//	b) Merges the new values with the existing slice via GetAllContentsWithUpdated,
//	c) Sets the merged slice on the model.
func (e engine[T]) UpdateModelFromUpdateParameters1(updateParams interface{}) (*T, error) {
	var model *T
	// Obtain the reflect.Value of the model (dereferenced) and the update parameters (also dereferenced)
	modelVal := reflect.ValueOf(model).Elem()
	paramsVal := reflect.ValueOf(updateParams).Elem()
	paramsType := paramsVal.Type()

	// Iterate over each field in the update parameters struct.
	for i := 0; i < paramsType.NumField(); i++ {
		fieldDef := paramsType.Field(i)
		updateField := paramsVal.Field(i)

		// Skip if update parameter field is nil (i.e. no update for this field)
		if updateField.IsNil() {
			continue
		}

		// Get the corresponding field in the model by name.
		modelField := modelVal.FieldByName(fieldDef.Name)
		if !modelField.IsValid() || !modelField.CanSet() {
			// No matching field, or it cannot be set; skip.
			continue
		}

		// Check if the model field is a slice. We handle content slices specially.
		//if  modelField.Kind() == reflect.Slice {
		//	elemType := modelField.Type().Elem()
		//	elemDummy := reflect.New(elemType).Elem().Interface()
		//	if orm.IsContentModel(elemDummy) {
		//		newContentsIface := updateField.Elem().Interface()
		//		// Assume elemType is the concrete struct (e.g., ContentModel)
		//		newContents := orm.ExtractContent(elemType)(newContentsIface)
		//		existingContents := modelField.Interface().([]orm.ContentModel) // Adjust with concrete type
		//		mergedContents := orm.GetAllContentsWithUpdated(existingContents, newContents)
		//		modelField.Set(reflect.ValueOf(mergedContents))
		//		continue
		//	}
		//}

		// For non-slice fields, assign the update value.
		// updateField is a pointer, so we take its element.
		newVal := updateField.Elem()

		if newVal.Type().AssignableTo(modelField.Type()) {
			modelField.Set(newVal)
		} else if newVal.Type().ConvertibleTo(modelField.Type()) {
			modelField.Set(newVal.Convert(modelField.Type()))
		} else {
			return nil, fmt.Errorf("cannot assign update parameter for field %s: type mismatch (%s vs %s)",
				fieldDef.Name, newVal.Type(), modelField.Type())
		}
	}

	return model, nil
}
