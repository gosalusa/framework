package dialects

import "gosalusa.com/extra/sets"

// DataType is a canonical, database-independent column type. Size optionally
// carries a length for types that support it, such as string.
type DataType struct {
	Name string
	Size int
}

var (
	// DataTypeBlob is a binary large object type.
	DataTypeBlob = DataType{Name: "blob"}
	// DataTypeString is a variable length string type. Use Size to set the
	// maximum length where the database supports it.
	DataTypeString = DataType{Name: "string"}
	// DataTypeText is a long string type.
	DataTypeText = DataType{Name: "text"}
	// DataTypeEnum is an enumerated string type.
	DataTypeEnum = DataType{Name: "enum"}

	// DataTypeBoolean is a boolean type.
	DataTypeBoolean = DataType{Name: "bool"}

	// DataTypeDate is a calendar date type.
	DataTypeDate = DataType{Name: "date"}
	// DataTypeDateTime is a date and time type.
	DataTypeDateTime = DataType{Name: "date-time"}

	// DataTypeFloat32 is a 32-bit floating point type.
	DataTypeFloat32 = DataType{Name: "float32"}
	// DataTypeFloat64 is a 64-bit floating point type.
	DataTypeFloat64 = DataType{Name: "float64"}

	// DataTypeInt8 is an 8-bit signed integer type.
	DataTypeInt8 = DataType{Name: "int8"}
	// DataTypeInt16 is a 16-bit signed integer type.
	DataTypeInt16 = DataType{Name: "int16"}
	// DataTypeInt32 is a 32-bit signed integer type.
	DataTypeInt32 = DataType{Name: "int32"}
	// DataTypeInt64 is a 64-bit signed integer type.
	DataTypeInt64 = DataType{Name: "int64"}

	// DataTypeUInt8 is an 8-bit unsigned integer type.
	DataTypeUInt8 = DataType{Name: "uint8"}
	// DataTypeUInt16 is a 16-bit unsigned integer type.
	DataTypeUInt16 = DataType{Name: "uint16"}
	// DataTypeUInt32 is a 32-bit unsigned integer type.
	DataTypeUInt32 = DataType{Name: "uint32"}
	// DataTypeUInt64 is a 64-bit unsigned integer type.
	DataTypeUInt64 = DataType{Name: "uint64"}

	// DataTypeJSON is a JSON document type.
	DataTypeJSON = DataType{Name: "json"}
)

var dataTypes = sets.New(
	DataTypeBlob.Name,
	DataTypeString.Name,
	DataTypeText.Name,
	DataTypeEnum.Name,

	DataTypeBoolean.Name,

	DataTypeDate.Name,
	DataTypeDateTime.Name,

	DataTypeFloat32.Name,
	DataTypeFloat64.Name,

	DataTypeInt8.Name,
	DataTypeInt16.Name,
	DataTypeInt32.Name,
	DataTypeInt64.Name,

	DataTypeUInt8.Name,
	DataTypeUInt16.Name,
	DataTypeUInt32.Name,
	DataTypeUInt64.Name,

	DataTypeJSON.Name,
)

// IsValid reports whether the DataType name is one of the known data type
// names.
func (d DataType) IsValid() bool {
	return dataTypes.Has(d.Name)
}

// DataTyper is implemented by types that report their own DataType. It must
// not be implemented on an interface.
type DataTyper interface {
	DataType() DataType
}
