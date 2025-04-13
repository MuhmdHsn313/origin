// Package orm provides a lightweight ORM framework tailored for multilingual content management.
// It defines common base models, interfaces, and utility functions for merging and extracting
// localized content models. The package leverages Go generics and reflection to allow flexible
// operations on types that implement the IContentModel interface.
package orm

import (
	"reflect"
	"time"
)

// Model is a base struct embedding common fields for all database entities.
// It includes a primary key and automatic timestamp tracking for creation and updates.
//
// Fields:
//   - ID: Unique identifier for the record. Annotated as the primary key for GORM.
//   - CreatedAt: Timestamp when the record is first created.
//   - UpdatedAt: Timestamp that updates automatically whenever the record is modified.
type Model struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;autoUpdateTime:milli"`
}

// IContentModel is an interface that must be implemented by all content models that
// support multilingual content. The sole responsibility of this interface is to return
// a language identifier, which is used to uniquely identify content by language.
type IContentModel interface {
	// GetLanguageID returns the identifier of the language associated with the content.
	// This identifier is used to merge or override localized content.
	GetLanguageID() string
}

// Language represents a language supported by the application.
// It is used to map content to human-readable language data and supports localization.
//
// Fields:
//   - ID: A short string representing the language code (e.g., "en", "ar").
//     This field is the primary key and is indexed for fast lookups.
//   - Title: The full name or description of the language.
//   - CreatedAt: Timestamp when the language record was created.
//   - UpdatedAt: Timestamp when the language record was last updated.
type Language struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(2);index"`
	Title     string    `json:"title" gorm:"not null;type:varchar(50)"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;autoUpdateTime:milli"`
}

// GetAllContentsWithUpdated merges two slices of content models (of generic type CM)
// that implement IContentModel. It produces a single slice where content with the same
// language identifier is overridden by the latter slice (inputContents).
//
// Parameters:
//   - currentContents: The current set of content models.
//   - inputContents: A new set of content models which can add new languages or override existing ones.
//
// Returns:
//   - A merged slice of content models, with each unique language identified exactly once.
//
// This function is useful for performing idempotent updates on multilingual content,
// ensuring that the latest content is present based on the language identifier.
func GetAllContentsWithUpdated[CM IContentModel](currentContents, inputContents []CM) []CM {
	// Initialize a map keyed by LanguageID to hold merged contents.
	// The map capacity is preallocated based on the total length of both slices.
	contentMap := make(map[string]CM, len(currentContents)+len(inputContents))

	// Add all existing content models into the map using their language identifier as the key.
	for _, content := range currentContents {
		contentMap[content.GetLanguageID()] = content
	}

	// Iterate over the new input content models, overwriting any that share the same language ID.
	// This ensures that any update in inputContents takes precedence over currentContents.
	for _, content := range inputContents {
		contentMap[content.GetLanguageID()] = content
	}

	// Prepare the final merged result slice by collecting all values from the map.
	contents := make([]CM, 0, len(contentMap))
	for _, content := range contentMap {
		contents = append(contents, content)
	}

	return contents
}

// ExtractContent dynamically maps a slice of arbitrary structs (inputContents)
// to a slice of a target type CM that implements IContentModel.
// It leverages reflection to iterate over the input slice and copy matching fields by name.
//
// Parameters:
//   - inputContents: An interface{} that is expected to be a slice of structs.
//     It must be a slice; otherwise, the function panics.
//
// Returns:
//   - A slice of type CM populated with data copied from inputContents.
//
// Implementation details:
//   - It first validates that inputContents is a slice.
//   - It then creates a new instance of the target type for each element and copies over
//     matching fields from the input element to the new content model.
//   - Only fields that exist in both the input struct and the target struct (by name)
//     and that can be set, are copied over.
//
// Use case:
//   - This function is useful when you have dynamic or loosely-typed data (such as JSON unmarshaled
//     into a slice of maps or structs) and you need to convert it into a strongly-typed slice of content models.
func ExtractContent[CM IContentModel](inputContents interface{}) []CM {
	// Determine the concrete type of the target content model using reflection.
	contentType := reflect.TypeOf(new(CM)).Elem()

	// Validate that inputContents is indeed a slice.
	inputValue := reflect.ValueOf(inputContents)
	if inputValue.Kind() != reflect.Slice {
		panic("inputContents must be a slice")
	}

	// Preallocate a result slice to hold the converted content models.
	result := make([]CM, 0, inputValue.Len())

	// Iterate over each element of the input slice.
	for i := 0; i < inputValue.Len(); i++ {
		inputItem := inputValue.Index(i)

		// Create a new instance of the target type.
		newContent := reflect.New(contentType).Elem()

		// Iterate through all fields in the input struct.
		// For each field, if a field with the same name exists in the target,
		// and if it is settable, copy the value.
		for j := 0; j < inputItem.NumField(); j++ {
			fieldName := inputItem.Type().Field(j).Name
			inputFieldValue := inputItem.Field(j)

			// Check if the target content model has the field and that it is writable.
			contentField := newContent.FieldByName(fieldName)
			if contentField.IsValid() && contentField.CanSet() {
				contentField.Set(inputFieldValue)
			}
		}

		// Append the newly populated instance to the result slice.
		// A type assertion converts the reflection-based instance into the target type.
		result = append(result, newContent.Interface().(CM))
	}

	return result
}

// IsContentModel is a generic utility function that checks if the provided model
// implements the IContentModel interface.
//
// Parameters:
//   - m: A model instance of any type that satisfies the comparable constraint.
//
// Returns:
//   - A boolean indicating whether the type of m implements IContentModel.
//
// This function is useful for enforcing type constraints at runtime,
// especially in contexts where generic types need to be validated.
func IsContentModel[model comparable](m model) bool {
	// Obtain the reflection type for the provided model instance.
	modelType := reflect.TypeOf(m)

	// Check whether the model type implements the IContentModel interface.
	return modelType.Implements(reflect.TypeOf((*IContentModel)(nil)).Elem())
}

// ContentModel is a concrete implementation of IContentModel.
// It represents a generic content record associated with a specific language,
// and includes both a reference to the Language struct and timestamp fields.
//
// Fields:
//   - LanguageID: Acts as a primary key for the content model and references the Language ID.
//   - Language: A nested struct containing language details.
//   - CreatedAt: Automatically set timestamp when the record is created.
//   - UpdatedAt: Automatically updated timestamp when the record is modified.
type ContentModel struct {
	LanguageID string    `json:"language_id" gorm:"primaryKey;type:varchar(2);index"`
	Language   Language  `json:"-"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"not null;autoUpdateTime:milli"`
}

// GetLanguageID returns the language identifier associated with this ContentModel.
// It satisfies the IContentModel interface.
func (content ContentModel) GetLanguageID() string {
	return content.LanguageID
}

// IsContained is a convenience method on ContentModel that checks whether the
// current content model is contained within a given slice of ContentModel instances.
// It utilizes the IsContentContained function for the comparison.
func (content ContentModel) IsContained(contents []ContentModel) bool {
	return IsContentContained(content, contents)
}

// IsContentContained is a helper function that checks if a given ContentModel exists
// in a slice of ContentModel based on the LanguageID.
//
// Parameters:
//   - content: The ContentModel instance to search for.
//   - contents: The slice of ContentModel instances to search within.
//
// Returns:
//   - true if a ContentModel with the same LanguageID is found; otherwise, false.
func IsContentContained(content ContentModel, contents []ContentModel) bool {
	for _, c := range contents {
		if c.LanguageID == content.LanguageID {
			return true
		}
	}
	return false
}
