package pgcql

import (
	"github.com/indexdata/cql-go/cql"
)

type FieldTsVector struct {
	FieldString
}

func NewFieldTsVector() *FieldTsVector {
	return &FieldTsVector{FieldString{language: "simple", disableTsConvert: true}}
}

func (f *FieldTsVector) WithColumn(column string) *FieldTsVector {
	f.column = column
	return f
}

func (f *FieldTsVector) WithLanguage(language string) *FieldTsVector {
	if language == "" {
		f.language = "simple"
	} else {
		f.language = language
	}
	return f
}

func (f *FieldTsVector) WithServerChoiceRel(relation cql.Relation) *FieldTsVector {
	f.serverChoiceRel = relation
	return f
}
