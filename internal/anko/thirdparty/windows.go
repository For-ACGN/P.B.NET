// +build windows

// Package thirdparty generate by script/code/anko/package.go, don't edit it.
package thirdparty

import (
	"reflect"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/mattn/anko/env"
)

func init() {
	initGithubComGoOLEGoOLE()
	initGithubComGoOLEGoOLEOLEUtil()
}

func initGithubComGoOLEGoOLE() {
	env.Packages["github.com/go-ole/go-ole"] = map[string]reflect.Value{
		// define constants
		"CC_CDECL":                 reflect.ValueOf(ole.CC_CDECL),
		"CC_FASTCALL":              reflect.ValueOf(ole.CC_FASTCALL),
		"CC_FPFASTCALL":            reflect.ValueOf(ole.CC_FPFASTCALL),
		"CC_MACPASCAL":             reflect.ValueOf(ole.CC_MACPASCAL),
		"CC_MAX":                   reflect.ValueOf(ole.CC_MAX),
		"CC_MPWCDECL":              reflect.ValueOf(ole.CC_MPWCDECL),
		"CC_MPWPASCAL":             reflect.ValueOf(ole.CC_MPWPASCAL),
		"CC_MSCPASCAL":             reflect.ValueOf(ole.CC_MSCPASCAL),
		"CC_PASCAL":                reflect.ValueOf(ole.CC_PASCAL),
		"CC_STDCALL":               reflect.ValueOf(ole.CC_STDCALL),
		"CC_SYSCALL":               reflect.ValueOf(ole.CC_SYSCALL),
		"CLSCTX_ALL":               reflect.ValueOf(ole.CLSCTX_ALL),
		"CLSCTX_INPROC":            reflect.ValueOf(ole.CLSCTX_INPROC),
		"CLSCTX_INPROC_HANDLER":    reflect.ValueOf(ole.CLSCTX_INPROC_HANDLER),
		"CLSCTX_INPROC_SERVER":     reflect.ValueOf(ole.CLSCTX_INPROC_SERVER),
		"CLSCTX_INPROC_SERVER16":   reflect.ValueOf(ole.CLSCTX_INPROC_SERVER16),
		"CLSCTX_LOCAL_SERVER":      reflect.ValueOf(ole.CLSCTX_LOCAL_SERVER),
		"CLSCTX_REMOTE_SERVER":     reflect.ValueOf(ole.CLSCTX_REMOTE_SERVER),
		"CLSCTX_SERVER":            reflect.ValueOf(ole.CLSCTX_SERVER),
		"COINIT_APARTMENTTHREADED": reflect.ValueOf(ole.COINIT_APARTMENTTHREADED),
		"COINIT_DISABLE_OLE1DDE":   reflect.ValueOf(ole.COINIT_DISABLE_OLE1DDE),
		"COINIT_MULTITHREADED":     reflect.ValueOf(ole.COINIT_MULTITHREADED),
		"COINIT_SPEED_OVER_MEMORY": reflect.ValueOf(ole.COINIT_SPEED_OVER_MEMORY),
		"CO_E_CLASSSTRING":         reflect.ValueOf(uint32(ole.CO_E_CLASSSTRING)),
		"DISPATCH_METHOD":          reflect.ValueOf(ole.DISPATCH_METHOD),
		"DISPATCH_PROPERTYGET":     reflect.ValueOf(ole.DISPATCH_PROPERTYGET),
		"DISPATCH_PROPERTYPUT":     reflect.ValueOf(ole.DISPATCH_PROPERTYPUT),
		"DISPATCH_PROPERTYPUTREF":  reflect.ValueOf(ole.DISPATCH_PROPERTYPUTREF),
		"DISPID_COLLECT":           reflect.ValueOf(ole.DISPID_COLLECT),
		"DISPID_CONSTRUCTOR":       reflect.ValueOf(ole.DISPID_CONSTRUCTOR),
		"DISPID_DESTRUCTOR":        reflect.ValueOf(ole.DISPID_DESTRUCTOR),
		"DISPID_EVALUATE":          reflect.ValueOf(ole.DISPID_EVALUATE),
		"DISPID_NEWENUM":           reflect.ValueOf(ole.DISPID_NEWENUM),
		"DISPID_PROPERTYPUT":       reflect.ValueOf(ole.DISPID_PROPERTYPUT),
		"DISPID_UNKNOWN":           reflect.ValueOf(ole.DISPID_UNKNOWN),
		"DISPID_VALUE":             reflect.ValueOf(ole.DISPID_VALUE),
		"E_ABORT":                  reflect.ValueOf(uint32(ole.E_ABORT)),
		"E_ACCESSDENIED":           reflect.ValueOf(uint32(ole.E_ACCESSDENIED)),
		"E_FAIL":                   reflect.ValueOf(uint32(ole.E_FAIL)),
		"E_HANDLE":                 reflect.ValueOf(uint32(ole.E_HANDLE)),
		"E_INVALIDARG":             reflect.ValueOf(uint32(ole.E_INVALIDARG)),
		"E_NOINTERFACE":            reflect.ValueOf(uint32(ole.E_NOINTERFACE)),
		"E_NOTIMPL":                reflect.ValueOf(uint32(ole.E_NOTIMPL)),
		"E_OUTOFMEMORY":            reflect.ValueOf(uint32(ole.E_OUTOFMEMORY)),
		"E_PENDING":                reflect.ValueOf(uint32(ole.E_PENDING)),
		"E_POINTER":                reflect.ValueOf(uint32(ole.E_POINTER)),
		"E_UNEXPECTED":             reflect.ValueOf(uint32(ole.E_UNEXPECTED)),
		"FADF_AUTO":                reflect.ValueOf(ole.FADF_AUTO),
		"FADF_BSTR":                reflect.ValueOf(ole.FADF_BSTR),
		"FADF_DISPATCH":            reflect.ValueOf(ole.FADF_DISPATCH),
		"FADF_EMBEDDED":            reflect.ValueOf(ole.FADF_EMBEDDED),
		"FADF_FIXEDSIZE":           reflect.ValueOf(ole.FADF_FIXEDSIZE),
		"FADF_HAVEIID":             reflect.ValueOf(ole.FADF_HAVEIID),
		"FADF_HAVEVARTYPE":         reflect.ValueOf(ole.FADF_HAVEVARTYPE),
		"FADF_RECORD":              reflect.ValueOf(ole.FADF_RECORD),
		"FADF_RESERVED":            reflect.ValueOf(ole.FADF_RESERVED),
		"FADF_STATIC":              reflect.ValueOf(ole.FADF_STATIC),
		"FADF_UNKNOWN":             reflect.ValueOf(ole.FADF_UNKNOWN),
		"FADF_VARIANT":             reflect.ValueOf(ole.FADF_VARIANT),
		"S_OK":                     reflect.ValueOf(ole.S_OK),
		"TKIND_ALIAS":              reflect.ValueOf(ole.TKIND_ALIAS),
		"TKIND_COCLASS":            reflect.ValueOf(ole.TKIND_COCLASS),
		"TKIND_DISPATCH":           reflect.ValueOf(ole.TKIND_DISPATCH),
		"TKIND_ENUM":               reflect.ValueOf(ole.TKIND_ENUM),
		"TKIND_INTERFACE":          reflect.ValueOf(ole.TKIND_INTERFACE),
		"TKIND_MAX":                reflect.ValueOf(ole.TKIND_MAX),
		"TKIND_MODULE":             reflect.ValueOf(ole.TKIND_MODULE),
		"TKIND_RECORD":             reflect.ValueOf(ole.TKIND_RECORD),
		"TKIND_UNION":              reflect.ValueOf(ole.TKIND_UNION),
		"VT_ARRAY":                 reflect.ValueOf(ole.VT_ARRAY),
		"VT_BLOB":                  reflect.ValueOf(ole.VT_BLOB),
		"VT_BLOB_OBJECT":           reflect.ValueOf(ole.VT_BLOB_OBJECT),
		"VT_BOOL":                  reflect.ValueOf(ole.VT_BOOL),
		"VT_BSTR":                  reflect.ValueOf(ole.VT_BSTR),
		"VT_BSTR_BLOB":             reflect.ValueOf(ole.VT_BSTR_BLOB),
		"VT_BYREF":                 reflect.ValueOf(ole.VT_BYREF),
		"VT_CARRAY":                reflect.ValueOf(ole.VT_CARRAY),
		"VT_CF":                    reflect.ValueOf(ole.VT_CF),
		"VT_CLSID":                 reflect.ValueOf(ole.VT_CLSID),
		"VT_CY":                    reflect.ValueOf(ole.VT_CY),
		"VT_DATE":                  reflect.ValueOf(ole.VT_DATE),
		"VT_DECIMAL":               reflect.ValueOf(ole.VT_DECIMAL),
		"VT_DISPATCH":              reflect.ValueOf(ole.VT_DISPATCH),
		"VT_EMPTY":                 reflect.ValueOf(ole.VT_EMPTY),
		"VT_ERROR":                 reflect.ValueOf(ole.VT_ERROR),
		"VT_FILETIME":              reflect.ValueOf(ole.VT_FILETIME),
		"VT_HRESULT":               reflect.ValueOf(ole.VT_HRESULT),
		"VT_I1":                    reflect.ValueOf(ole.VT_I1),
		"VT_I2":                    reflect.ValueOf(ole.VT_I2),
		"VT_I4":                    reflect.ValueOf(ole.VT_I4),
		"VT_I8":                    reflect.ValueOf(ole.VT_I8),
		"VT_ILLEGAL":               reflect.ValueOf(ole.VT_ILLEGAL),
		"VT_ILLEGALMASKED":         reflect.ValueOf(ole.VT_ILLEGALMASKED),
		"VT_INT":                   reflect.ValueOf(ole.VT_INT),
		"VT_INT_PTR":               reflect.ValueOf(ole.VT_INT_PTR),
		"VT_LPSTR":                 reflect.ValueOf(ole.VT_LPSTR),
		"VT_LPWSTR":                reflect.ValueOf(ole.VT_LPWSTR),
		"VT_NULL":                  reflect.ValueOf(ole.VT_NULL),
		"VT_PTR":                   reflect.ValueOf(ole.VT_PTR),
		"VT_R4":                    reflect.ValueOf(ole.VT_R4),
		"VT_R8":                    reflect.ValueOf(ole.VT_R8),
		"VT_RECORD":                reflect.ValueOf(ole.VT_RECORD),
		"VT_RESERVED":              reflect.ValueOf(ole.VT_RESERVED),
		"VT_SAFEARRAY":             reflect.ValueOf(ole.VT_SAFEARRAY),
		"VT_STORAGE":               reflect.ValueOf(ole.VT_STORAGE),
		"VT_STORED_OBJECT":         reflect.ValueOf(ole.VT_STORED_OBJECT),
		"VT_STREAM":                reflect.ValueOf(ole.VT_STREAM),
		"VT_STREAMED_OBJECT":       reflect.ValueOf(ole.VT_STREAMED_OBJECT),
		"VT_TYPEMASK":              reflect.ValueOf(ole.VT_TYPEMASK),
		"VT_UI1":                   reflect.ValueOf(ole.VT_UI1),
		"VT_UI2":                   reflect.ValueOf(ole.VT_UI2),
		"VT_UI4":                   reflect.ValueOf(ole.VT_UI4),
		"VT_UI8":                   reflect.ValueOf(ole.VT_UI8),
		"VT_UINT":                  reflect.ValueOf(ole.VT_UINT),
		"VT_UINT_PTR":              reflect.ValueOf(ole.VT_UINT_PTR),
		"VT_UNKNOWN":               reflect.ValueOf(ole.VT_UNKNOWN),
		"VT_USERDEFINED":           reflect.ValueOf(ole.VT_USERDEFINED),
		"VT_VARIANT":               reflect.ValueOf(ole.VT_VARIANT),
		"VT_VECTOR":                reflect.ValueOf(ole.VT_VECTOR),
		"VT_VOID":                  reflect.ValueOf(ole.VT_VOID),

		// define variables
		"CLSID_COMEchoTestObject":       reflect.ValueOf(ole.CLSID_COMEchoTestObject),
		"CLSID_COMTestScalarClass":      reflect.ValueOf(ole.CLSID_COMTestScalarClass),
		"IID_ICOMEchoTestObject":        reflect.ValueOf(ole.IID_ICOMEchoTestObject),
		"IID_ICOMTestBoolean":           reflect.ValueOf(ole.IID_ICOMTestBoolean),
		"IID_ICOMTestDouble":            reflect.ValueOf(ole.IID_ICOMTestDouble),
		"IID_ICOMTestFloat":             reflect.ValueOf(ole.IID_ICOMTestFloat),
		"IID_ICOMTestInt16":             reflect.ValueOf(ole.IID_ICOMTestInt16),
		"IID_ICOMTestInt32":             reflect.ValueOf(ole.IID_ICOMTestInt32),
		"IID_ICOMTestInt64":             reflect.ValueOf(ole.IID_ICOMTestInt64),
		"IID_ICOMTestInt8":              reflect.ValueOf(ole.IID_ICOMTestInt8),
		"IID_ICOMTestString":            reflect.ValueOf(ole.IID_ICOMTestString),
		"IID_ICOMTestTypes":             reflect.ValueOf(ole.IID_ICOMTestTypes),
		"IID_IConnectionPoint":          reflect.ValueOf(ole.IID_IConnectionPoint),
		"IID_IConnectionPointContainer": reflect.ValueOf(ole.IID_IConnectionPointContainer),
		"IID_IDispatch":                 reflect.ValueOf(ole.IID_IDispatch),
		"IID_IEnumVariant":              reflect.ValueOf(ole.IID_IEnumVariant),
		"IID_IInspectable":              reflect.ValueOf(ole.IID_IInspectable),
		"IID_IProvideClassInfo":         reflect.ValueOf(ole.IID_IProvideClassInfo),
		"IID_IUnknown":                  reflect.ValueOf(ole.IID_IUnknown),
		"IID_NULL":                      reflect.ValueOf(ole.IID_NULL),

		// define functions
		"BstrToString":            reflect.ValueOf(ole.BstrToString),
		"BytePtrToString":         reflect.ValueOf(ole.BytePtrToString),
		"CLSIDFromProgID":         reflect.ValueOf(ole.CLSIDFromProgID),
		"CLSIDFromString":         reflect.ValueOf(ole.CLSIDFromString),
		"ClassIDFrom":             reflect.ValueOf(ole.ClassIDFrom),
		"CoInitialize":            reflect.ValueOf(ole.CoInitialize),
		"CoInitializeEx":          reflect.ValueOf(ole.CoInitializeEx),
		"CoTaskMemFree":           reflect.ValueOf(ole.CoTaskMemFree),
		"CoUninitialize":          reflect.ValueOf(ole.CoUninitialize),
		"Connect":                 reflect.ValueOf(ole.Connect),
		"CreateDispTypeInfo":      reflect.ValueOf(ole.CreateDispTypeInfo),
		"CreateInstance":          reflect.ValueOf(ole.CreateInstance),
		"CreateStdDispatch":       reflect.ValueOf(ole.CreateStdDispatch),
		"DeleteHString":           reflect.ValueOf(ole.DeleteHString),
		"DispatchMessage":         reflect.ValueOf(ole.DispatchMessage),
		"GetActiveObject":         reflect.ValueOf(ole.GetActiveObject),
		"GetMessage":              reflect.ValueOf(ole.GetMessage),
		"GetObject":               reflect.ValueOf(ole.GetObject),
		"GetUserDefaultLCID":      reflect.ValueOf(ole.GetUserDefaultLCID),
		"GetVariantDate":          reflect.ValueOf(ole.GetVariantDate),
		"IIDFromString":           reflect.ValueOf(ole.IIDFromString),
		"IsEqualGUID":             reflect.ValueOf(ole.IsEqualGUID),
		"LpOleStrToString":        reflect.ValueOf(ole.LpOleStrToString),
		"NewError":                reflect.ValueOf(ole.NewError),
		"NewErrorWithDescription": reflect.ValueOf(ole.NewErrorWithDescription),
		"NewErrorWithSubError":    reflect.ValueOf(ole.NewErrorWithSubError),
		"NewGUID":                 reflect.ValueOf(ole.NewGUID),
		"NewHString":              reflect.ValueOf(ole.NewHString),
		"NewVariant":              reflect.ValueOf(ole.NewVariant),
		"RoActivateInstance":      reflect.ValueOf(ole.RoActivateInstance),
		"RoGetActivationFactory":  reflect.ValueOf(ole.RoGetActivationFactory),
		"RoInitialize":            reflect.ValueOf(ole.RoInitialize),
		"StringFromCLSID":         reflect.ValueOf(ole.StringFromCLSID),
		"StringFromIID":           reflect.ValueOf(ole.StringFromIID),
		"SysAllocString":          reflect.ValueOf(ole.SysAllocString),
		"SysAllocStringLen":       reflect.ValueOf(ole.SysAllocStringLen),
		"SysFreeString":           reflect.ValueOf(ole.SysFreeString),
		"SysStringLen":            reflect.ValueOf(ole.SysStringLen),
		"UTF16PtrToString":        reflect.ValueOf(ole.UTF16PtrToString),
		"VariantClear":            reflect.ValueOf(ole.VariantClear),
		"VariantInit":             reflect.ValueOf(ole.VariantInit),
	}
	var (
		bindOpts                      ole.BindOpts
		dISPPARAMS                    ole.DISPPARAMS
		dispatch                      ole.Dispatch
		eXCEPINFO                     ole.EXCEPINFO
		gUID                          ole.GUID
		hString                       ole.HString
		iConnectionPoint              ole.IConnectionPoint
		iConnectionPointContainer     ole.IConnectionPointContainer
		iConnectionPointContainerVtbl ole.IConnectionPointContainerVtbl
		iConnectionPointVtbl          ole.IConnectionPointVtbl
		iDLDESC                       ole.IDLDESC
		iDispatch                     ole.IDispatch
		iDispatchVtbl                 ole.IDispatchVtbl
		iEnumVARIANT                  ole.IEnumVARIANT
		iEnumVARIANTVtbl              ole.IEnumVARIANTVtbl
		iInspectable                  ole.IInspectable
		iInspectableVtbl              ole.IInspectableVtbl
		iNTERFACEDATA                 ole.INTERFACEDATA
		iProvideClassInfo             ole.IProvideClassInfo
		iProvideClassInfoVtbl         ole.IProvideClassInfoVtbl
		iTypeInfo                     ole.ITypeInfo
		iTypeInfoVtbl                 ole.ITypeInfoVtbl
		iUnknown                      ole.IUnknown
		iUnknownVtbl                  ole.IUnknownVtbl
		mETHODDATA                    ole.METHODDATA
		msg                           ole.Msg
		oleError                      ole.OleError
		pARAMDATA                     ole.PARAMDATA
		point                         ole.Point
		sAFEARRAY                     ole.SAFEARRAY
		sAFEARRAYBOUND                ole.SAFEARRAYBOUND
		safeArray                     ole.SafeArray
		safeArrayBound                ole.SafeArrayBound
		safeArrayConversion           ole.SafeArrayConversion
		tYPEATTR                      ole.TYPEATTR
		tYPEDESC                      ole.TYPEDESC
		unknownLike                   ole.UnknownLike
		vARIANT                       ole.VARIANT
		vT                            ole.VT
	)
	env.PackageTypes["github.com/go-ole/go-ole"] = map[string]reflect.Type{
		"BindOpts":                      reflect.TypeOf(&bindOpts).Elem(),
		"DISPPARAMS":                    reflect.TypeOf(&dISPPARAMS).Elem(),
		"Dispatch":                      reflect.TypeOf(&dispatch).Elem(),
		"EXCEPINFO":                     reflect.TypeOf(&eXCEPINFO).Elem(),
		"GUID":                          reflect.TypeOf(&gUID).Elem(),
		"HString":                       reflect.TypeOf(&hString).Elem(),
		"IConnectionPoint":              reflect.TypeOf(&iConnectionPoint).Elem(),
		"IConnectionPointContainer":     reflect.TypeOf(&iConnectionPointContainer).Elem(),
		"IConnectionPointContainerVtbl": reflect.TypeOf(&iConnectionPointContainerVtbl).Elem(),
		"IConnectionPointVtbl":          reflect.TypeOf(&iConnectionPointVtbl).Elem(),
		"IDLDESC":                       reflect.TypeOf(&iDLDESC).Elem(),
		"IDispatch":                     reflect.TypeOf(&iDispatch).Elem(),
		"IDispatchVtbl":                 reflect.TypeOf(&iDispatchVtbl).Elem(),
		"IEnumVARIANT":                  reflect.TypeOf(&iEnumVARIANT).Elem(),
		"IEnumVARIANTVtbl":              reflect.TypeOf(&iEnumVARIANTVtbl).Elem(),
		"IInspectable":                  reflect.TypeOf(&iInspectable).Elem(),
		"IInspectableVtbl":              reflect.TypeOf(&iInspectableVtbl).Elem(),
		"INTERFACEDATA":                 reflect.TypeOf(&iNTERFACEDATA).Elem(),
		"IProvideClassInfo":             reflect.TypeOf(&iProvideClassInfo).Elem(),
		"IProvideClassInfoVtbl":         reflect.TypeOf(&iProvideClassInfoVtbl).Elem(),
		"ITypeInfo":                     reflect.TypeOf(&iTypeInfo).Elem(),
		"ITypeInfoVtbl":                 reflect.TypeOf(&iTypeInfoVtbl).Elem(),
		"IUnknown":                      reflect.TypeOf(&iUnknown).Elem(),
		"IUnknownVtbl":                  reflect.TypeOf(&iUnknownVtbl).Elem(),
		"METHODDATA":                    reflect.TypeOf(&mETHODDATA).Elem(),
		"Msg":                           reflect.TypeOf(&msg).Elem(),
		"OleError":                      reflect.TypeOf(&oleError).Elem(),
		"PARAMDATA":                     reflect.TypeOf(&pARAMDATA).Elem(),
		"Point":                         reflect.TypeOf(&point).Elem(),
		"SAFEARRAY":                     reflect.TypeOf(&sAFEARRAY).Elem(),
		"SAFEARRAYBOUND":                reflect.TypeOf(&sAFEARRAYBOUND).Elem(),
		"SafeArray":                     reflect.TypeOf(&safeArray).Elem(),
		"SafeArrayBound":                reflect.TypeOf(&safeArrayBound).Elem(),
		"SafeArrayConversion":           reflect.TypeOf(&safeArrayConversion).Elem(),
		"TYPEATTR":                      reflect.TypeOf(&tYPEATTR).Elem(),
		"TYPEDESC":                      reflect.TypeOf(&tYPEDESC).Elem(),
		"UnknownLike":                   reflect.TypeOf(&unknownLike).Elem(),
		"VARIANT":                       reflect.TypeOf(&vARIANT).Elem(),
		"VT":                            reflect.TypeOf(&vT).Elem(),
	}
}

func initGithubComGoOLEGoOLEOLEUtil() {
	env.Packages["github.com/go-ole/go-ole"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CallMethod":         reflect.ValueOf(oleutil.CallMethod),
		"ClassIDFrom":        reflect.ValueOf(oleutil.ClassIDFrom),
		"ConnectObject":      reflect.ValueOf(oleutil.ConnectObject),
		"CreateObject":       reflect.ValueOf(oleutil.CreateObject),
		"ForEach":            reflect.ValueOf(oleutil.ForEach),
		"GetActiveObject":    reflect.ValueOf(oleutil.GetActiveObject),
		"GetProperty":        reflect.ValueOf(oleutil.GetProperty),
		"MustCallMethod":     reflect.ValueOf(oleutil.MustCallMethod),
		"MustGetProperty":    reflect.ValueOf(oleutil.MustGetProperty),
		"MustPutProperty":    reflect.ValueOf(oleutil.MustPutProperty),
		"MustPutPropertyRef": reflect.ValueOf(oleutil.MustPutPropertyRef),
		"PutProperty":        reflect.ValueOf(oleutil.PutProperty),
		"PutPropertyRef":     reflect.ValueOf(oleutil.PutPropertyRef),
	}
	var ()
	env.PackageTypes["github.com/go-ole/go-ole"] = map[string]reflect.Type{}
}
