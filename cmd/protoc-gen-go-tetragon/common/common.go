// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cilium/tetragon/pkg/logger"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TetragonPackageName is the import path for the Tetragon package
var TetragonPackageName = "github.com/cilium/tetragon"

// TetragonApiPackageName is the import path for the code generated package
var TetragonApiPackageName = "api/v1/tetragon"

// TetragonCopyrightHeader is the license header to prepend to all generated files
var TetragonCopyrightHeader = `// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon`

// NewGeneratedFile creates a new codegen pakage and file in the project
func NewGeneratedFile(gen *protogen.Plugin, file *protogen.File, pkg string) *protogen.GeneratedFile {
	pkgName := filepath.Base(pkg)
	importPath := filepath.Join(string(file.GoImportPath), "codegen", pkg)
	pathSuffix := filepath.Base(file.GeneratedFilenamePrefix)
	fileName := filepath.Join(strings.TrimSuffix(file.GeneratedFilenamePrefix, pathSuffix), "codegen", pkg, fmt.Sprintf("%s.pb.go", pkgName))
	logger.GetLogger().Infof("%s", fileName)

	g := gen.NewGeneratedFile(fileName, protogen.GoImportPath(importPath))
	g.P(TetragonCopyrightHeader)
	g.P()

	g.P("// Code generated by protoc-gen-go-tetragon. DO NOT EDIT")
	g.P()

	g.P("package ", pkgName)
	g.P()

	return g
}

// GoIdent is a convenience helper that returns a qualified go ident as a string for
// a given import package and name
func GoIdent(g *protogen.GeneratedFile, importPath string, name string) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

// TetragonApiIdent is a convenience helper that calls GoIdent with the path to the
// Tetragon API package.
func TetragonApiIdent(g *protogen.GeneratedFile, name string) string {
	return TetragonIdent(g, TetragonApiPackageName, name)
}

// TetragonIdent is a convenience helper that calls GoIdent with the path to the
// Tetragon package.
func TetragonIdent(g *protogen.GeneratedFile, importPath string, name string) string {
	importPath = filepath.Join(TetragonPackageName, importPath)
	return GoIdent(g, importPath, name)
}

// GeneratedIdent is a convenience helper that returns a qualified go ident as a string for
// a given import package and name within the codegen package
func GeneratedIdent(g *protogen.GeneratedFile, importPath string, name string) string {
	importPath = filepath.Join(TetragonPackageName, TetragonApiPackageName, "codegen", importPath)
	return GoIdent(g, importPath, name)
}

// Logger is a convenience helper that generates a call to logger.GetLogger()
func Logger(g *protogen.GeneratedFile) string {
	return fmt.Sprintf("%s()", GoIdent(g, "github.com/cilium/tetragon/pkg/logger", "GetLogger"))
}

// FmtErrorf is a convenience helper that generates a call to fmt.Errorf
func FmtErrorf(g *protogen.GeneratedFile, fmt_ string, args ...string) string {
	args = append([]string{fmt.Sprintf("\"%s\"", fmt_)}, args...)
	return fmt.Sprintf("%s(%s)", GoIdent(g, "fmt", "Errorf"), strings.Join(args, ", "))
}

// EventFieldCheck returns true if the event has the field
func EventFieldCheck(msg *protogen.Message, field string) bool {
	if msg.Desc.Fields().ByName(protoreflect.Name(field)) != nil {
		return true
	}

	return false
}

// IsProcessEvent returns true if the message is an Tetragon event that has a process field
func IsProcessEvent(msg *protogen.Message) bool {
	return EventFieldCheck(msg, "process")
}

// IsParentEvent returns true if the message is an Tetragon event that has a parent field
func IsParentEvent(msg *protogen.Message) bool {
	return EventFieldCheck(msg, "parent")
}

// StructTag is a convenience helper that formats a struct tag
func StructTag(tag string) string {
	return fmt.Sprintf("`%s`", tag)
}

var eventsCache []*protogen.Message

// GetEvents returns a list of all messages that are events
func GetEvents(f *protogen.File) ([]*protogen.Message, error) {
	if len(eventsCache) == 0 {
		var getEventsResponse *protogen.Message
		for _, msg := range f.Messages {
			if msg.GoIdent.GoName == "GetEventsResponse" {
				getEventsResponse = msg
				break
			}
		}
		if getEventsResponse == nil {
			return nil, fmt.Errorf("Unable to find GetEventsResponse message")
		}

		var eventOneof *protogen.Oneof
		for _, oneof := range getEventsResponse.Oneofs {
			if oneof.Desc.Name() == "event" {
				eventOneof = oneof
				break
			}
		}
		if eventOneof == nil {
			return nil, fmt.Errorf("Unable to find GetEventsResponse.event")
		}

		validNames := make(map[string]struct{})
		for _, type_ := range eventOneof.Fields {
			name := strings.TrimPrefix(type_.GoIdent.GoName, "GetEventsResponse_")
			validNames[name] = struct{}{}
		}

		for _, msg := range f.Messages {
			if _, ok := validNames[string(msg.Desc.Name())]; ok {
				eventsCache = append(eventsCache, msg)
			}
		}
	}

	return eventsCache, nil
}

var fieldsCache []*protogen.Message

// GetFields returns a list of all messages that are fields
func GetFields(f *protogen.File) ([]*protogen.Message, error) {
	if len(fieldsCache) == 0 {
		events, err := GetEvents(f)
		if err != nil {
			return nil, err
		}

		validFields := make(map[string]struct{})
		for _, event := range events {
			fields := getFieldsForMessage(event)
			for _, field := range fields {
				if field.Message == nil {
					continue
				}
				validFields[field.Message.GoIdent.GoName] = struct{}{}
			}
		}

		for _, msg := range f.Messages {
			if _, ok := validFields[string(msg.Desc.Name())]; ok {
				fieldsCache = append(fieldsCache, msg)
			}
		}
	}

	return fieldsCache, nil
}

// getFieldsForMessage recursively looks up all the fields for a given message
func getFieldsForMessage(msg *protogen.Message) []*protogen.Field {
	seen := make(map[string]struct{})
	return __getFieldsForMessage(msg, seen)
}

// __getFieldsForMessage is the underlying recusion logic of getFieldsForMessage
func __getFieldsForMessage(msg *protogen.Message, seen map[string]struct{}) []*protogen.Field {
	var fields []*protogen.Field

	for _, field := range msg.Fields {
		if field.Message == nil {
			continue
		}
		fieldType := field.Message.GoIdent.GoName
		if _, ok := seen[fieldType]; ok {
			continue
		}
		seen[fieldType] = struct{}{}
		fields = append(fields, field)
		fields = append(fields, __getFieldsForMessage(field.Message, seen)...)
	}

	return fields
}

var enumsCache []*protogen.Enum

// GetEnums returns a list of all enums that are message fields
func GetEnums(f *protogen.File) ([]*protogen.Enum, error) {
	if len(enumsCache) == 0 {
		events, err := GetEvents(f)
		if err != nil {
			return nil, err
		}

		fields, err := GetFields(f)
		if err != nil {
			return nil, err
		}

		eventsAndFields := append(events, fields...)

		validNames := make(map[string]struct{})
		for _, msg := range eventsAndFields {
			for _, field := range msg.Fields {
				enum := field.Enum
				if enum == nil {
					continue
				}
				validNames[string(enum.Desc.Name())] = struct{}{}
			}
		}

		for _, enum := range f.Enums {
			if _, ok := validNames[string(enum.Desc.Name())]; ok {
				enumsCache = append(enumsCache, enum)
			}
		}
	}

	return enumsCache, nil
}
