// +build windows

package wmi

import (
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/pkg/errors"
)

// property types about objects
//
// Find it from https://github.com/angelcolmenares/pash/blob/master/
// External/System.Management/System.Management/tag_CIMTYPE_ENUMERATION.cs
//
// after some time find it from microsoft
// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-wmio/
// e137e6c6-c1cc-449e-a0b4-76fabf534480
//
// [!shit mountain!]
const (
	CIMTypeInt8      uint8 = 16
	CIMTypeInt16     uint8 = 2
	CIMTypeInt32     uint8 = 3
	CIMTypeInt64     uint8 = 20
	CIMTypeUint8     uint8 = 17
	CIMTypeUint16    uint8 = 18
	CIMTypeUint32    uint8 = 19
	CIMTypeUint64    uint8 = 21
	CIMTypeFloat32   uint8 = 4
	CIMTypeFloat64   uint8 = 5
	CIMTypeString    uint8 = 8
	CIMTypeBool      uint8 = 11
	CIMTypeDateTime  uint8 = 101
	CIMTypeReference uint8 = 102
	CIMTypeChar16    uint8 = 103
	CIMTypeObject    uint8 = 13
)

// Object returned by Client.Get().
type Object struct {
	raw *ole.VARIANT
}

// count is used to get the number of objects.
func (obj *Object) count() (int, error) {
	iDispatch := obj.raw.ToIDispatch()
	if iDispatch == nil {
		return 0, nil
	}
	iDispatch.AddRef()
	defer iDispatch.Release()
	count, err := oleutil.GetProperty(iDispatch, "Count")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get Count property")
	}
	defer func() { _ = count.Clear() }()
	return int(count.Val), nil
}

// need clear object.
func (obj *Object) itemIndex(i int) (*Object, error) {
	iDispatch := obj.raw.ToIDispatch()
	if iDispatch == nil {
		return nil, errors.New("object is not callable")
	}
	iDispatch.AddRef()
	defer iDispatch.Release()
	itemRaw, err := oleutil.CallMethod(iDispatch, "ItemIndex", i)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call ItemIndex")
	}
	return &Object{raw: itemRaw}, nil
}

// need clear each object, after use.
func (obj *Object) objects() ([]*Object, error) {
	count, err := obj.count()
	if err != nil {
		return nil, err
	}
	objects := make([]*Object, count)
	for i := 0; i < count; i++ {
		objects[i], err = obj.itemIndex(i)
		if err != nil {
			// clear objects
			for j := 0; j < i; j++ {
				objects[j].Clear()
			}
			return nil, err
		}
	}
	return objects, nil
}

// ExecMethod is used to execute a method on the object.
// need call Clear after use.
func (obj *Object) ExecMethod(method string, args ...interface{}) (*Object, error) {
	iDispatch := obj.raw.ToIDispatch()
	if iDispatch == nil {
		return nil, errors.New("object is not callable")
	}
	iDispatch.AddRef()
	defer iDispatch.Release()
	returnValue, err := oleutil.CallMethod(iDispatch, method, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call method \"%s\"", method)
	}
	return &Object{raw: returnValue}, nil
}

// GetProperty is used to get property of this object, need clear object.
// need call Clear after use.
func (obj *Object) GetProperty(name string) (*Object, error) {
	iDispatch := obj.raw.ToIDispatch()
	if iDispatch == nil {
		return nil, errors.New("object is not callable")
	}
	iDispatch.AddRef()
	defer iDispatch.Release()
	prop, err := oleutil.GetProperty(iDispatch, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get property \"%s\"", name)
	}
	return &Object{raw: prop}, nil
}

// SetProperty is used to set property of this object.
func (obj *Object) SetProperty(name string, args ...interface{}) error {
	iDispatch := obj.raw.ToIDispatch()
	if iDispatch == nil {
		return errors.New("object is not callable")
	}
	iDispatch.AddRef()
	defer iDispatch.Release()
	switch arg := args[0].(type) {
	case time.Time: // process time.Time to WMI date time string
		args[0] = timeToWMIDateTime(arg)
	case *time.Time:
		args[0] = timeToWMIDateTime(*arg)
	case *Object: // save CIM_Object
		iDispatch := arg.raw.ToIDispatch()
		iDispatch.AddRef()
		defer iDispatch.Release()
		args[0] = iDispatch
	}
	result, err := oleutil.PutProperty(iDispatch, name, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to set property \"%s\"", name)
	}
	defer func() { _ = result.Clear() }()
	return nil
}

// AddProperty is used to add a property to object.
func (obj *Object) AddProperty(name string, typ uint8, isArray bool) error {
	properties, err := obj.GetProperty("Properties_")
	if err != nil {
		return err
	}
	defer properties.Clear()
	result, err := properties.ExecMethod("Add", name, typ, isArray)
	if err != nil {
		return errors.Wrapf(err, "failed to add property \"%s\"", name)
	}
	result.Clear()
	return nil
}

// RemoveProperty is used to remove property.
func (obj *Object) RemoveProperty(name string) error {
	properties, err := obj.GetProperty("Properties_")
	if err != nil {
		return err
	}
	defer properties.Clear()
	result, err := properties.ExecMethod("Remove", name)
	if err != nil {
		return errors.Wrapf(err, "failed to remove property \"%s\"", name)
	}
	result.Clear()
	return nil
}

// GetMethodInputParameters is used to get input parameters about a method.
// need call Clear after use.
func (obj *Object) GetMethodInputParameters(name string) (*Object, error) {
	methods, err := obj.GetProperty("Methods_")
	if err != nil {
		return nil, err
	}
	defer methods.Clear()
	method, err := methods.ExecMethod("Item", name)
	if err != nil {
		return nil, err
	}
	defer method.Clear()
	input, err := method.GetProperty("InParameters")
	if err != nil {
		return nil, err
	}
	return input, nil
}

// Path is used to get path about this object.
func (obj *Object) Path() (string, error) {
	prop, err := obj.GetProperty("Path_")
	if err != nil {
		return "", err
	}
	defer prop.Clear()
	path, err := prop.GetProperty("Path")
	if err != nil {
		return "", err
	}
	defer path.Clear()
	return path.Value().(string), nil
}

// Value is used to return the value of a result as an interface.
func (obj *Object) Value() interface{} {
	return obj.raw.Value()
}

// ToArray is used to return array values in a []interface.
func (obj *Object) ToArray() []interface{} {
	return obj.raw.ToArray().ToValueArray()
}

// ToIDispatch is used to convert object to *ole.IDispatch.
// need call Release after use.
func (obj *Object) ToIDispatch() *ole.IDispatch {
	return obj.raw.ToIDispatch()
}

// Clear is used to clear the memory of variant object.
func (obj *Object) Clear() {
	_ = obj.raw.Clear()
}
