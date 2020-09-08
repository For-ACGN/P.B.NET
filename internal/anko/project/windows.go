// +build windows

// Package project generate by script/code/anko/package.go, don't edit it.
package project

import (
	"reflect"

	"github.com/mattn/anko/env"

	"project/internal/module/windows/privilege"
	"project/internal/module/windows/wmi"
)

func init() {
	initInternalModuleWindowsWMI()
	initInternalModuleWindowsPrivilege()
}

func initInternalModuleWindowsWMI() {
	env.Packages["project/internal/module/windows/wmi"] = map[string]reflect.Value{
		// define constants
		"CIMTypeBool":      reflect.ValueOf(wmi.CIMTypeBool),
		"CIMTypeChar16":    reflect.ValueOf(wmi.CIMTypeChar16),
		"CIMTypeDateTime":  reflect.ValueOf(wmi.CIMTypeDateTime),
		"CIMTypeFloat32":   reflect.ValueOf(wmi.CIMTypeFloat32),
		"CIMTypeFloat64":   reflect.ValueOf(wmi.CIMTypeFloat64),
		"CIMTypeInt16":     reflect.ValueOf(wmi.CIMTypeInt16),
		"CIMTypeInt32":     reflect.ValueOf(wmi.CIMTypeInt32),
		"CIMTypeInt64":     reflect.ValueOf(wmi.CIMTypeInt64),
		"CIMTypeInt8":      reflect.ValueOf(wmi.CIMTypeInt8),
		"CIMTypeObject":    reflect.ValueOf(wmi.CIMTypeObject),
		"CIMTypeReference": reflect.ValueOf(wmi.CIMTypeReference),
		"CIMTypeString":    reflect.ValueOf(wmi.CIMTypeString),
		"CIMTypeUint16":    reflect.ValueOf(wmi.CIMTypeUint16),
		"CIMTypeUint32":    reflect.ValueOf(wmi.CIMTypeUint32),
		"CIMTypeUint64":    reflect.ValueOf(wmi.CIMTypeUint64),
		"CIMTypeUint8":     reflect.ValueOf(wmi.CIMTypeUint8),

		// define variables

		// define functions
		"BuildWQLStatement": reflect.ValueOf(wmi.BuildWQLStatement),
		"NewClient":         reflect.ValueOf(wmi.NewClient),
	}
	var (
		client           wmi.Client
		errFieldMismatch wmi.ErrFieldMismatch
		object           wmi.Object
		options          wmi.Options
	)
	env.PackageTypes["project/internal/module/windows/wmi"] = map[string]reflect.Type{
		"Client":           reflect.TypeOf(&client).Elem(),
		"ErrFieldMismatch": reflect.TypeOf(&errFieldMismatch).Elem(),
		"Object":           reflect.TypeOf(&object).Elem(),
		"Options":          reflect.TypeOf(&options).Elem(),
	}
}

func initInternalModuleWindowsPrivilege() {
	env.Packages["project/internal/module/windows/privilege"] = map[string]reflect.Value{
		// define constants
		"SEBackup":         reflect.ValueOf(privilege.SEBackup),
		"SEDebug":          reflect.ValueOf(privilege.SEDebug),
		"SELoadDriver":     reflect.ValueOf(privilege.SELoadDriver),
		"SERemoteShutdown": reflect.ValueOf(privilege.SERemoteShutdown),
		"SESecurity":       reflect.ValueOf(privilege.SESecurity),
		"SEShutdown":       reflect.ValueOf(privilege.SEShutdown),
		"SESystemEnv":      reflect.ValueOf(privilege.SESystemEnv),
		"SESystemProf":     reflect.ValueOf(privilege.SESystemProf),
		"SESystemTime":     reflect.ValueOf(privilege.SESystemTime),
		"SeDebug":          reflect.ValueOf(privilege.SeDebug),
		"SeShutdown":       reflect.ValueOf(privilege.SeShutdown),

		// define variables

		// define functions
		"EnableDebug":              reflect.ValueOf(privilege.EnableDebug),
		"EnablePrivilege":          reflect.ValueOf(privilege.EnablePrivilege),
		"EnableShutdown":           reflect.ValueOf(privilege.EnableShutdown),
		"RtlAdjustPrivilege":       reflect.ValueOf(privilege.RtlAdjustPrivilege),
		"RtlDisableBackup":         reflect.ValueOf(privilege.RtlDisableBackup),
		"RtlDisableDebug":          reflect.ValueOf(privilege.RtlDisableDebug),
		"RtlDisableLoadDriver":     reflect.ValueOf(privilege.RtlDisableLoadDriver),
		"RtlDisableRemoteShutdown": reflect.ValueOf(privilege.RtlDisableRemoteShutdown),
		"RtlDisableSecurity":       reflect.ValueOf(privilege.RtlDisableSecurity),
		"RtlDisableShutdown":       reflect.ValueOf(privilege.RtlDisableShutdown),
		"RtlDisableSystemEnv":      reflect.ValueOf(privilege.RtlDisableSystemEnv),
		"RtlDisableSystemProf":     reflect.ValueOf(privilege.RtlDisableSystemProf),
		"RtlDisableSystemTime":     reflect.ValueOf(privilege.RtlDisableSystemTime),
		"RtlEnableBackup":          reflect.ValueOf(privilege.RtlEnableBackup),
		"RtlEnableDebug":           reflect.ValueOf(privilege.RtlEnableDebug),
		"RtlEnableLoadDriver":      reflect.ValueOf(privilege.RtlEnableLoadDriver),
		"RtlEnableRemoteShutdown":  reflect.ValueOf(privilege.RtlEnableRemoteShutdown),
		"RtlEnableSecurity":        reflect.ValueOf(privilege.RtlEnableSecurity),
		"RtlEnableShutdown":        reflect.ValueOf(privilege.RtlEnableShutdown),
		"RtlEnableSystemEnv":       reflect.ValueOf(privilege.RtlEnableSystemEnv),
		"RtlEnableSystemProf":      reflect.ValueOf(privilege.RtlEnableSystemProf),
		"RtlEnableSystemTime":      reflect.ValueOf(privilege.RtlEnableSystemTime),
	}
	var ()
	env.PackageTypes["project/internal/module/windows/privilege"] = map[string]reflect.Type{}
}
