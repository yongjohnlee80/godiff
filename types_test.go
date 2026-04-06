package godiff

import (
	"reflect"
	"testing"
)

func TestDataTypes_Merge(t *testing.T) {
	tests := []struct {
		name     string
		initial  DataTypes
		other    DataTypes
		expected DataTypes
	}{
		{
			name:     "merge with non-overlapping keys",
			initial:  DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
			other:    DataTypes{"key2": DataType{Kind: "type2", Tags: []string{"tag2"}}},
			expected: DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}, "key2": DataType{Kind: "type2", Tags: []string{"tag2"}}},
		},
		{
			name:     "merge with overlapping keys (overwrite existing key)",
			initial:  DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
			other:    DataTypes{"key1": DataType{Kind: "newType1", Tags: []string{"newTag1", "newTag2"}}},
			expected: DataTypes{"key1": DataType{Kind: "newType1", Tags: []string{"newTag1", "newTag2"}}},
		},
		{
			name:     "merge with empty initial DataTypes",
			initial:  DataTypes{},
			other:    DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
			expected: DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
		},
		{
			name:     "merge with empty other DataTypes",
			initial:  DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
			other:    DataTypes{},
			expected: DataTypes{"key1": DataType{Kind: "type1", Tags: []string{"tag1"}}},
		},
		{
			name:     "merge empty initial and other DataTypes",
			initial:  DataTypes{},
			other:    DataTypes{},
			expected: DataTypes{},
		},
		{
			name:     "merge with special characters in keys and tags",
			initial:  DataTypes{"key#@!": DataType{Kind: "type1", Tags: []string{"special1"}}},
			other:    DataTypes{"key#@!": DataType{Kind: "type2", Tags: []string{"special2"}}},
			expected: DataTypes{"key#@!": DataType{Kind: "type2", Tags: []string{"special2"}}},
		},
		{
			name:     "merge with nil tags",
			initial:  DataTypes{"key1": DataType{Kind: "type1", Tags: nil}},
			other:    DataTypes{"key2": DataType{Kind: "type2", Tags: nil}},
			expected: DataTypes{"key1": DataType{Kind: "type1", Tags: nil}, "key2": DataType{Kind: "type2", Tags: nil}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Merge(&tt.other)
			if !reflect.DeepEqual(tt.initial, tt.expected) {
				t.Errorf("Merge() = %v, want %v", tt.initial, tt.expected)
			}
		})
	}
}

func TestDataTypes_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  DataTypes
		field    string
		typ      string
		tags     []string
		expected DataType
	}{
		{
			name:     "add new field with type and tags",
			initial:  DataTypes{},
			field:    "key1",
			typ:      "type1",
			tags:     []string{"tag1", "tag2"},
			expected: DataType{Kind: "type1", Tags: []string{"tag1", "tag2"}},
		},
		{
			name:     "add new field with empty type and no tags",
			initial:  DataTypes{},
			field:    "key2",
			typ:      "",
			tags:     []string{},
			expected: DataType{Kind: "", Tags: []string{}},
		},
		{
			name:     "overwrite existing field with new type and tags",
			initial:  DataTypes{"key3": DataType{Kind: "oldType", Tags: []string{"oldTag"}}},
			field:    "key3",
			typ:      "newType",
			tags:     []string{"newTag1", "newTag2"},
			expected: DataType{Kind: "newType", Tags: []string{"newTag1", "newTag2"}},
		},
		{
			name:     "add field with special characters in key and tags",
			initial:  DataTypes{},
			field:    "key#@!",
			typ:      "specialType",
			tags:     []string{"special1", "special2"},
			expected: DataType{Kind: "specialType", Tags: []string{"special1", "special2"}},
		},
		{
			name:     "add field with nil tags",
			initial:  DataTypes{},
			field:    "keyNil",
			typ:      "typeNil",
			tags:     nil,
			expected: DataType{Kind: "typeNil", Tags: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Add(tt.field, tt.typ, tt.tags...)
			if got := tt.initial[tt.field]; !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Add(%q, %q, %v) = %v, want %v", tt.field, tt.typ, tt.tags, got, tt.expected)
			}
		})
	}
}

func TestDataTypes_Get(t *testing.T) {
	tests := []struct {
		name     string
		data     DataTypes
		field    string
		expected string
	}{
		{
			name:     "get existing key",
			data:     DataTypes{"key1": DataType{Kind: "value1"}},
			field:    "key1",
			expected: "value1",
		},
		{
			name:     "get non-existing key",
			data:     DataTypes{"key1": DataType{Kind: "value1"}},
			field:    "key2",
			expected: "",
		},
		{
			name:     "get existing key with empty value",
			data:     DataTypes{"key1": DataType{Kind: ""}},
			field:    "key1",
			expected: "",
		},
		{
			name:     "get key with special characters",
			data:     DataTypes{"key#@!": DataType{Kind: "special_value"}},
			field:    "key#@!",
			expected: "special_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.Get(tt.field)
			if got != tt.expected {
				t.Errorf("Get(%q) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}
}
func TestDataTypes_GetTags(t *testing.T) {
	tests := []struct {
		name     string
		data     DataTypes
		field    string
		expected []string
	}{
		{
			name:     "get tags for existing key",
			data:     DataTypes{"key1": DataType{Tags: []string{"tag1", "tag2"}}},
			field:    "key1",
			expected: []string{"tag1", "tag2"},
		},
		{
			name:     "get tags for non-existing key",
			data:     DataTypes{"key1": DataType{Tags: []string{"tag1", "tag2"}}},
			field:    "key2",
			expected: nil,
		},
		{
			name:     "get tags for key with no tags",
			data:     DataTypes{"key1": DataType{Tags: []string{}}},
			field:    "key1",
			expected: []string{},
		},
		{
			name:     "get tags for key with special characters",
			data:     DataTypes{"key#@!": DataType{Tags: []string{"special1", "special2"}}},
			field:    "key#@!",
			expected: []string{"special1", "special2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.GetTags(tt.field)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetTags(%q) = %v, want %v", tt.field, got, tt.expected)
			}
		})
	}
}
