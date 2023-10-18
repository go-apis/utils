package xsheet

import (
	"path"
	"reflect"
	"strings"
)

var excelOldTypes = []string{
	"application/vnd.ms-excel",
	"text/csv",
}

var excelTypes = []string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template",
	"application/vnd.ms-excel.sheet.macroEnabled.12",
	"application/vnd.ms-excel.template.macroEnabled.12",
	"application/vnd.ms-excel.addin.macroEnabled.12",
	"application/vnd.ms-excel.sheet.binary.macroEnabled.12",
}

var excelExt = []string{
	"xls",
	"xlsx",
}

func IsCsvFilename(fileName string) bool {
	ext := path.Ext(fileName)
	return strings.EqualFold(ext, ".csv")
}

func IsExcelFileType(fileType string) bool {
	for _, t := range excelTypes {
		if strings.EqualFold(t, fileType) {
			return true
		}
	}

	// TODO verify this.
	for _, t := range excelExt {
		if strings.EqualFold(t, fileType) {
			return true
		}
	}

	return false
}

func IsExcelOldFileType(fileType string) bool {
	for _, t := range excelOldTypes {
		if strings.EqualFold(t, fileType) {
			return true
		}
	}
	return false
}

func ToType[T any]() reflect.Type {
	var n T
	t := reflect.TypeOf(n)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
