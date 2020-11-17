// +build windows

package wmi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/pkg/errors"

	"project/internal/xpanic"
)

// references:
//
// IWbemServices:
// https://docs.microsoft.com/en-us/windows/win32/api/wbemcli/nf-wbemcli-iwbemservices-execquery
// https://docs.microsoft.com/en-us/windows/win32/api/wbemcli/nf-wbemcli-iwbemservices-getobject
//
// SWbemObject object:
// https://docs.microsoft.com/en-us/windows/win32/wmisdk/swbemobject
// https://docs.microsoft.com/en-us/windows/win32/wmisdk/swbemobject-execmethod-
// https://docs.microsoft.com/en-us/windows/win32/wmisdk/swbemobject-methods-
// https://docs.microsoft.com/en-us/windows/win32/wmisdk/swbemobject-path-
// https://docs.microsoft.com/en-us/windows/win32/wmisdk/swbemmethodset
//
// CIM Win32 Provider:
// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-process
// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem

const defaultInitTimeout = 15 * time.Second

// Options contains options about WMI client.
type Options struct {
	Host        string        `toml:"host"`
	Username    string        `toml:"username"`
	Password    string        `toml:"password"`
	InitTimeout time.Duration `toml:"init_timeout"`
}

type execQuery struct {
	WQL string
	Dst interface{}
	Err chan<- error
}

type getObject struct {
	Path   string
	Args   []interface{}
	Result chan<- *getObjectResult
}

type getObjectResult struct {
	Object *Object
	Err    error
}

type execMethod struct {
	Path   string
	Method string
	Input  interface{}
	Output interface{}
	Err    chan<- error
}

// Client is a WMI client.
type Client struct {
	namespace string
	opts      *Options

	initErr chan error

	unknown *ole.IUnknown
	query   *ole.IDispatch
	rawSvc  *ole.VARIANT
	wmi     *ole.IDispatch

	queryQueue chan *execQuery
	getQueue   chan *getObject
	execQueue  chan *execMethod

	stopSignal chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// NewClient is used to create a WMI client.
func NewClient(namespace string, opts *Options) (*Client, error) {
	const queueSize = 16
	if opts == nil {
		opts = new(Options)
	}
	// set connect server arguments
	client := Client{
		namespace:  namespace,
		opts:       opts,
		initErr:    make(chan error, 1),
		queryQueue: make(chan *execQuery, queueSize),
		getQueue:   make(chan *getObject, queueSize),
		execQueue:  make(chan *execMethod, queueSize),
		stopSignal: make(chan struct{}),
	}
	client.wg.Add(1)
	go client.work()
	// wait WMI client initialize
	timeout := opts.InitTimeout
	if timeout < time.Second {
		timeout = defaultInitTimeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-client.initErr:
		if err != nil {
			return nil, err
		}
		return &client, nil
	case <-timer.C:
		client.Close()
		return nil, errors.New("initialize client timeout")
	}
}

func (client *Client) work() {
	defer client.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "Client.work")
		}
	}()
	defer client.close()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := client.init()
	select {
	case client.initErr <- err:
	case <-client.stopSignal:
		return
	}
	if err != nil {
		return
	}
	defer client.release()

	client.handleLoop()
}

func (client *Client) init() error {
	var ok bool
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		errCode := err.(*ole.OleError).Code()
		// CoInitialize already called
		// https://msdn.microsoft.com/en-us/library/windows/desktop/ms695279%28v=vs.85%29.aspx
		if errCode != ole.S_OK && errCode != 0x00000001 { // S_FALSE
			return err
		}
	}
	defer func() {
		if !ok {
			ole.CoUninitialize()
		}
	}()
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return errors.Wrap(err, "failed to create object")
	}
	defer func() {
		if !ok {
			unknown.Release()
		}
	}()
	query, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return errors.Wrap(err, "failed to query interface")
	}
	defer func() {
		if !ok {
			query.Release()
			query.Release()
		}
	}()
	// start connect server, service is a SWbemServices.
	opts := client.opts
	args := []interface{}{opts.Host, client.namespace, opts.Username, opts.Password}
	rawSvc, err := oleutil.CallMethod(query, "ConnectServer", args...)
	if err != nil {
		return errors.Wrap(err, "failed to connect server")
	}
	client.unknown = unknown
	client.query = query
	client.rawSvc = rawSvc
	client.wmi = rawSvc.ToIDispatch()
	ok = true
	return nil
}

func (client *Client) release() {
	client.wmi.Release()
	_ = client.rawSvc.Clear()
	client.query.Release()
	client.query.Release()
	client.unknown.Release()
	ole.CoUninitialize()
}

const clientClosed = "wmi client is closed"

// Query is used to query with WQL, dst is used to save query result.
// destination interface must be slice pointer like *[]*Type or *[]Type.
func (client *Client) Query(wql string, dst interface{}) error {
	errCh := make(chan error, 1)
	query := execQuery{
		WQL: wql,
		Dst: dst,
		Err: errCh,
	}
	select {
	case client.queryQueue <- &query:
	case <-client.stopSignal:
		return errors.New("failed to query: " + clientClosed)
	}
	var err error
	select {
	case err = <-errCh:
	case <-client.stopSignal:
		return errors.New("failed to receive query error: " + clientClosed)
	}
	return err
}

// GetObject is used to retrieves a class or instance.
func (client *Client) GetObject(path string, args ...interface{}) (*Object, error) {
	result := make(chan *getObjectResult, 1)
	get := getObject{
		Path:   path,
		Args:   args,
		Result: result,
	}
	select {
	case client.getQueue <- &get:
	case <-client.stopSignal:
		return nil, errors.New("failed to get object: " + clientClosed)
	}
	var getResult *getObjectResult
	select {
	case getResult = <-result:
	case <-client.stopSignal:
		return nil, errors.New("failed to receive get object result: " + clientClosed)
	}
	if getResult.Err != nil {
		return nil, getResult.Err
	}
	return getResult.Object, nil
}

// ExecMethod is used to execute a method about object, dst is used to save execute result.
// destination interface must be structure pointer like *struct.
func (client *Client) ExecMethod(path, method string, input, output interface{}) error {
	errCh := make(chan error, 1)
	exec := execMethod{
		Path:   path,
		Method: method,
		Input:  input,
		Output: output,
		Err:    errCh,
	}
	select {
	case client.execQueue <- &exec:
	case <-client.stopSignal:
		return errors.New("failed to execute method: " + clientClosed)
	}
	var err error
	select {
	case err = <-errCh:
	case <-client.stopSignal:
		return errors.New("failed to receive execute method error: " + clientClosed)
	}
	return err
}

func (client *Client) handleLoop() {
	var (
		query *execQuery
		get   *getObject
		exec  *execMethod
	)
	for {
		select {
		case query = <-client.queryQueue:
			client.handleExecQuery(query)
		case get = <-client.getQueue:
			client.handleGetObject(get)
		case exec = <-client.execQueue:
			client.handleExecMethod(exec)
		case <-client.stopSignal:
			return
		}
	}
}

func (client *Client) handleExecQuery(query *execQuery) {
	var err error
	defer func() { query.Err <- err }()

	result, err := oleutil.CallMethod(client.wmi, "ExecQuery", query.WQL)
	if err != nil {
		err = errors.Wrap(err, "failed to exec query")
		return
	}
	object := Object{raw: result}
	defer object.Clear()

	objects, err := object.objects()
	if err != nil {
		return
	}
	defer func() {
		for i := 0; i < len(objects); i++ {
			objects[i].Clear()
		}
	}()
	err = parseExecQueryResult(objects, query.Dst)
}

func (client *Client) handleGetObject(get *getObject) {
	var result getObjectResult
	defer func() { get.Result <- &result }()

	params := append([]interface{}{get.Path}, get.Args...)
	instance, err := oleutil.CallMethod(client.wmi, "Get", params...)
	if err != nil {
		result.Err = errors.Wrapf(err, "failed to get object \"%s\"", get.Path)
		return
	}
	result.Object = &Object{raw: instance}
}

func (client *Client) handleExecMethod(exec *execMethod) {
	var err error
	defer func() { exec.Err <- err }()

	// get class, don't use client.GetObject() or will block
	class, err := oleutil.CallMethod(client.wmi, "Get", exec.Path)
	if err != nil {
		err = errors.Wrapf(err, "failed to get class \"%s\"", exec.Path)
		return
	}
	object := Object{raw: class}
	defer object.Clear()
	// execute method
	var output *Object
	if exec.Input != nil { // set input parameters
		var input *Object
		input, err = object.GetMethodInputParameters(exec.Method)
		if err != nil {
			return
		}
		defer input.Clear()
		err = client.setExecMethodInput(input, exec.Input)
		if err != nil {
			return
		}
		iDispatch := input.ToIDispatch()
		iDispatch.AddRef()
		defer iDispatch.Release()
		output, err = object.ExecMethod("ExecMethod_", exec.Method, iDispatch)
	} else {
		output, err = object.ExecMethod("ExecMethod_", exec.Method)
	}
	if err != nil {
		return
	}
	defer output.Clear()
	err = parseExecMethodOutput(output, exec.Output)
}

// setExecMethodInput is used to set input parameters to object.
func (client *Client) setExecMethodInput(obj *Object, input interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "setExecMethodInput")
		}
	}()
	// check input type
	typ := reflect.TypeOf(input)
	val := reflect.ValueOf(input)
	switch typ.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		if val.IsNil() {
			panic("input pointer is nil")
		}
		typ = typ.Elem()
		if typ.Kind() != reflect.Struct {
			panic("input pointer is not point to structure")
		}
		val = val.Elem()
	default:
		panic("input interface is not structure or pointer")
	}
	return client.walkStruct(obj, typ, val)
}

func (client *Client) walkStruct(obj *Object, typ reflect.Type, val reflect.Value) error {
	fields := getStructFields(typ)
	for i := 0; i < len(fields); i++ {
		// skipped field
		if fields[i] == "" {
			continue
		}
		err := client.setValue(obj, fields[i], typ.Field(i).Type, val.Field(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// setValue is used to set value to property.
func (client *Client) setValue(obj *Object, name string, typ reflect.Type, val reflect.Value) error {
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() { // skip nil point
			return nil
		}
		typ = typ.Elem()
		val = val.Elem()
	}
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
		return obj.SetProperty(name, val.Interface())
	case reflect.Slice: // []string, []byte will panic.
		switch typ.Elem().Kind() {
		case reflect.String: // ,reflect.Uint8
			return obj.SetProperty(name, val.Interface())
		default:
			const format = "unsupported type about slice element, name: %s type: %s"
			panic(fmt.Sprintf(format, name, typ.Name()))
		}
	case reflect.Struct:
		switch typ {
		case timeType:
			return obj.SetProperty(name, val.Interface())
		default:
			return client.setStruct(obj, name, typ, val)
		}
	default:
		const format = "unsupported field type, name: %s type: %s"
		panic(fmt.Sprintf(format, name, typ.Name()))
	}
}

// setStruct is used to set structure to property.
func (client *Client) setStruct(obj *Object, name string, typ reflect.Type, val reflect.Value) error {
	// get class name from structure field
	classField, ok := typ.FieldByName("Class")
	if !ok {
		const format = "\"class\" field is not in structure %s"
		panic(fmt.Sprintf(format, typ.Name()))
	}
	if classField.Type.Kind() != reflect.String {
		const format = "\"class\" field is not string type in structure %s"
		panic(fmt.Sprintf(format, typ.Name()))
	}
	class := val.FieldByName("Class").Interface().(string)
	if class == "" {
		const format = "\"class\" field is empty in structure %s"
		panic(fmt.Sprintf(format, typ.Name()))
	}
	// create instance
	instance, err := oleutil.CallMethod(client.wmi, "Get", class)
	if err != nil {
		return errors.Wrapf(err, "failed to create instance from class \"%s\"", class)
	}
	object := Object{raw: instance}
	defer object.Clear()
	// set fields
	err = client.walkStruct(&object, typ, val)
	if err != nil {
		return err
	}
	// set instance
	return obj.SetProperty(name, instance)
}

// Close is used to close WMI client.
func (client *Client) Close() {
	client.close()
	client.wg.Wait()
}

func (client *Client) close() {
	client.closeOnce.Do(func() {
		close(client.stopSignal)
	})
}

// BuildWQLStatement is used to build WQL from structure.
//
// type win32Process struct {
//     Name   string
//     PID    uint32 `wmi:"ProcessId"`
//     Ignore string `wmi:"-"`
// }
//
// to "select Name, ProcessId from Win32_Process"
func BuildWQLStatement(structure interface{}, form string) string {
	fields := getStructFields(reflect.TypeOf(structure))
	fieldsLen := len(fields)
	// remove empty string
	f := make([]string, 0, fieldsLen)
	for i := 0; i < fieldsLen; i++ {
		if fields[i] != "" {
			f = append(f, fields[i])
		}
	}
	return "select " + strings.Join(f, ", ") + " from " + form
}
