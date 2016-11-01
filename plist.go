// Copyright 2016 Vinzenz Feenstra. All rights reserved.
// Use of this source code is governed by a BSD-2-clause
// license that can be found in the LICENSE file.
package plist

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

var whitespaceReplacer *strings.Replacer

func init() {
	whitespaceReplacer = strings.NewReplacer("\t", "", " ", "")
}

var InvalidTypeError = fmt.Errorf("Invalid Value Type")

type invalidPListError struct {
	inputOffset   int64
	internalError error
}

func (self invalidPListError) Error() string {
	return fmt.Sprintf("PList error line: %d: %s", self.inputOffset, self.internalError.Error())
}

func plistErrorFromString(offset int64, msg string) *invalidPListError {
	return &invalidPListError{
		offset,
		fmt.Errorf("%s", msg),
	}
}

func plistErrorFromError(offset int64, err error) *invalidPListError {
	return &invalidPListError{
		offset,
		err,
	}
}

type ValueType int

const (
	// InvalidType refers to an invalid value.
	InvalidType ValueType = iota
	// StringType refers to string.
	StringType
	// DateType refers to time.Time.
	DateType
	// IntegerType refers to int64.
	IntegerType
	// RealType refers to float64.
	RealType
	// BooleanType refers to bool.
	BooleanType
	// DataType refers to []byte.
	DataType
	// DictType refers to map[string]Value.
	DictType
	// ArrayType refers to []Value
	ArrayType

	typeCount
)

var valueTypeNames = [typeCount]string{
	InvalidType: "invalid",
	StringType:  "string",
	DateType:    "date",
	IntegerType: "integer",
	RealType:    "real",
	BooleanType: "boolean",
	DataType:    "data",
	DictType:    "dict",
	ArrayType:   "array",
}

// Name returns a human readable string as name of the ValueType
func (self ValueType) Name() string {
	return valueTypeNames[self]
}

// Value holds the data and type information
type Value struct {
	Value interface{}
	Type  ValueType
}

// InvalidValue is a conenience pre-initialized constant to return on errors.
var InvalidValue = Value{nil, InvalidType}

const preamble = xml.Header + `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
`

// Write writes the plist representation of this Value instance to writer.
func (self Value) Write(writer io.Writer) error {
	if _, err := io.WriteString(writer, preamble); err != nil {
		return err
	}
	encoder := xml.NewEncoder(writer)
	elem := xml.StartElement{Name: xml.Name{Local: "plist"}, Attr: []xml.Attr{{Name: xml.Name{Space: "", Local: "version"}, Value: "1.0"}}}
	encoder.Indent("", "  ")
	if err := encoder.EncodeToken(elem); err != nil {
		return err
	}
	if err := self.writeXml(encoder); err != nil {
		return err
	}
	if err := encoder.EncodeToken(elem.End()); err != nil {
		return err
	}
	return encoder.Flush()
}

func encodeElem(encoder *xml.Encoder, value interface{}, name string) error {
	return encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: name}})
}

func (self Value) writeXml(encoder *xml.Encoder) error {
	switch self.Type {
	case ArrayType:
		elem := xml.StartElement{Name: xml.Name{Local: "array"}}
		if err := encoder.EncodeToken(elem); err != nil {
			return err
		}
		for _, v := range self.Value.([]Value) {
			if err := v.writeXml(encoder); err != nil {
				return err
			}
		}
		return encoder.EncodeToken(elem.End())
	case DictType:
		elem := xml.StartElement{Name: xml.Name{Local: "dict"}}
		if err := encoder.EncodeToken(elem); err != nil {
			return err
		}
		m := self.Value.(map[string]Value)
		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if err := encodeElem(encoder, k, "key"); err != nil {
				return err
			}
			if err := m[k].writeXml(encoder); err != nil {
				return err
			}
		}
		return encoder.EncodeToken(elem.End())
	case StringType:
		return encodeElem(encoder, self.Value, "string")
	case IntegerType:
		return encodeElem(encoder, self.Value, "integer")
	case RealType:
		return encodeElem(encoder, self.Value, "real")
	case DataType:
		return encodeElem(encoder, base64.StdEncoding.EncodeToString(self.Value.([]byte)), "data")
	case DateType:
		return encodeElem(encoder, self.Value, "date")
	case BooleanType:
		if !self.Value.(bool) {
			return encodeElem(encoder, "", "false")
		} else {
			return encodeElem(encoder, "", "true")
		}
	}
	return InvalidTypeError
}

// Raw returns a pure golang structure of the value data instead of Value wrapped objects.
// Dicts become map[string]interface{} and arrays []interface{}
// Otherwise the value types stay as defined.
func (self Value) Raw() interface{} {
	switch self.Type {
	case ArrayType:
		result := make([]interface{}, len(self.Value.([]Value)))
		for i, e := range self.Value.([]Value) {
			result[i] = e.Raw()
		}
		return result
	case DictType:
		result := map[string]interface{}{}
		for k, v := range self.Value.(map[string]Value) {
			result[k] = v.Raw()
		}
		return result
	default:
		return self.Value
	}
}

// Read parses a plist xml representation from reader.
func Read(reader io.Reader) (Value, error) {
	decoder := xml.NewDecoder(reader)
	for {
		if token, err := decoder.Token(); err != nil {
			return InvalidValue, err
		} else {
			if element, ok := token.(xml.StartElement); ok {
				if element.Name.Local != "plist" {
					return InvalidValue, plistErrorFromError(decoder.InputOffset(), fmt.Errorf("Unexpected element %s", element.Name.Local))
				}
				break
			}
		}
	}
	return readValue(decoder)
}

type decodeFilter func(string) (Value, error)

func elementDecoder(decoder *xml.Decoder, element xml.StartElement) func(decodeFilter) (Value, error) {
	return func(filter decodeFilter) (Value, error) {
		var data xml.CharData
		if err := decoder.DecodeElement(&data, &element); err != nil {
			return InvalidValue, err
		} else {
			return filter(string(data))
		}
	}
}

func nullFilter(s string) (Value, error) {
	return Value{s, StringType}, nil
}

func valueWrap(valueType ValueType) func(value interface{}, err error) (Value, error) {
	return func(value interface{}, err error) (Value, error) {
		if err != nil {
			return InvalidValue, err
		}
		return Value{value, valueType}, nil
	}
}

func parseElement(decoder *xml.Decoder, element xml.StartElement) (Value, error) {
	decodeData := elementDecoder(decoder, element)
	switch element.Name.Local {
	case "string":
		return decodeData(nullFilter)
	case "date":
		return decodeData(func(s string) (Value, error) {
			return valueWrap(DateType)(time.ParseInLocation(time.RFC3339, s, time.UTC))
		})
	case "integer":
		return decodeData(func(s string) (Value, error) {
			if len(s) > 2 && strings.ToLower(s[:2]) == "0x" {
				return valueWrap(IntegerType)(strconv.ParseInt(s[2:], 16, 64))
			}
			return valueWrap(IntegerType)(strconv.ParseInt(s, 10, 64))
		})
	case "real":
		return decodeData(func(s string) (Value, error) {
			return valueWrap(RealType)(strconv.ParseFloat(s, 64))
		})
	case "true", "false":
		decoder.Skip()
		return valueWrap(BooleanType)(strings.ToLower(element.Name.Local) == "true", nil)
	case "data":
		return decodeData(func(s string) (Value, error) {
			return valueWrap(DataType)(base64.StdEncoding.DecodeString(whitespaceReplacer.Replace(s)))
		})
	case "dict":
		result := map[string]Value{}
		for {
			if token, err := decoder.Token(); err == nil {
				if element, ok := token.(xml.EndElement); ok {
					if element.Name.Local == "dict" {
						return Value{result, DictType}, nil
					}
				} else if element, ok := token.(xml.StartElement); ok {
					if element.Name.Local == "key" {
						if key, err := elementDecoder(decoder, element)(nullFilter); err != nil {
							return InvalidValue, err
						} else {
							if value, err := readValue(decoder); err != nil {
								return InvalidValue, err
							} else {
								result[key.Value.(string)] = value
							}
						}
					} else {
						return InvalidValue, fmt.Errorf("Unexpected element '%s' at %d", element.Name.Local, decoder.InputOffset())
					}
				}
			} else {
				return InvalidValue, err
			}
		}
	case "array":
		result := []Value{}
		for {
			if token, err := decoder.Token(); err == nil {
				if element, ok := token.(xml.EndElement); ok {
					if element.Name.Local == "array" {
						return Value{result, ArrayType}, nil
					}
				} else if element, ok := token.(xml.StartElement); ok {
					if value, err := parseElement(decoder, element); err != nil {
						return InvalidValue, err
					} else {
						result = append(result, value)
					}
				}
			} else {
				return InvalidValue, err
			}
		}
	}
	return InvalidValue, fmt.Errorf("Unsupported element %s at %d", element.Name.Local, decoder.InputOffset())
}

func readValue(decoder *xml.Decoder) (Value, error) {
	for {
		if token, err := decoder.Token(); err == nil {
			if element, ok := token.(xml.StartElement); ok {
				return parseElement(decoder, element)
			}
		} else {
			return InvalidValue, plistErrorFromError(decoder.InputOffset(), err)
		}
	}
	return InvalidValue, fmt.Errorf("Unknown error")
}
