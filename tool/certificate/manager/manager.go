package manager

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"project/internal/cert"
	"project/internal/cert/certmgr"
	"project/internal/cert/certpool"
	"project/internal/security"
	"project/internal/system"
)

const (
	prefixManager         = "manager"
	prefixPublic          = "manager/public"
	prefixPublicRootCA    = "manager/public/root-ca"
	prefixPublicClientCA  = "manager/public/client-ca"
	prefixPublicClient    = "manager/public/client"
	prefixPrivate         = "manager/private"
	prefixPrivateRootCA   = "manager/private/root-ca"
	prefixPrivateClientCA = "manager/private/client-ca"
	prefixPrivateClient   = "manager/private/client"
)

const managerHelp = `
help about manager:
  
  public       switch to public mode
  private      switch to private mode
  help         print help information
  save         save certificate pool
  reload       reload certificate pool
  exit         close certificate manager
  
`

const locationHelpTemplate = `
help about manager/%s:
  
  root-ca      switch to %s/root-ca mode
  client-ca    switch to %s/client-ca mode
  client       switch to %s/client mode
  help         print help information
  save         save certificate pool
  reload       reload certificate pool
  return       return to the manager
  exit         close certificate manager
  
`

const certHelpTemplate = `
help about manager/%s:

  print        print certificate information with ID
                 example: print 0
  add          add a certificate with private key
                 example: add "certs.pem" ["keys.pem"]
  delete       delete a certificate with ID
                 example: delete 0
  export       export certificate and private key with ID
                 example: export 0 "cert.pem" ["key.pem"]
  list         list %s certificates with simple information
  help         print help information
  save         save certificate pool
  reload       reload certificate pool
  return       return to the %s mode
  exit         close certificate manager

`

// Manager is the certificate manager CUI program.
type Manager struct {
	stdin    io.Reader
	dataPath string
	bakPath  string
	password *security.Bytes
	pool     *certpool.Pool
	prefix   string
	scanner  *bufio.Scanner
	closed   bool
	testMode bool
}

// New is used to create a certificate manager.
func New(stdin io.Reader, path string) *Manager {
	return &Manager{
		stdin:    stdin,
		dataPath: path,
		bakPath:  path + ".bak",
	}
}

// Initialize is used to initialize certificate manager.
func (mgr *Manager) Initialize(password []byte) error {
	// check data file is exists
	exist, err := system.IsExist(mgr.dataPath)
	if err != nil {
		return err
	}
	if exist {
		const format = "certificate pool file \"%s\" is already exists\n"
		return errors.Errorf(format, mgr.dataPath)
	}
	// load system certificates
	pool, err := certpool.NewPoolWithSystem()
	if err != nil {
		return err
	}
	// save certificate pool
	data, err := certmgr.SaveCtrlCertPool(pool, password)
	if err != nil {
		return err
	}
	fmt.Println("add certificates from system")
	err = system.WriteFile(mgr.dataPath, data)
	if err != nil {
		return err
	}
	fmt.Println("initialize certificate manager successfully")
	return nil
}

// ResetPassword is used to reset certificate manager password.
func (mgr *Manager) ResetPassword(oldPwd, newPwd []byte) error {
	// load certificate pool with old password
	data, err := os.ReadFile(mgr.dataPath)
	if err != nil {
		return err
	}
	pool := certpool.NewPool()
	err = certmgr.LoadCtrlCertPool(pool, data, oldPwd)
	if err != nil {
		return err
	}
	// save certificate pool with new password
	data, err = certmgr.SaveCtrlCertPool(pool, newPwd)
	if err != nil {
		return err
	}
	err = system.WriteFile(mgr.dataPath, data)
	if err != nil {
		return err
	}
	fmt.Println("reset certificate manager password successfully")
	return nil
}

// Manage is used to manage certificate, it will cover password slice.
func (mgr *Manager) Manage(password []byte) error {
	defer security.CoverBytes(password)
	// check data file is exists
	exist, err := system.IsExist(mgr.dataPath)
	if err != nil {
		return err
	}
	if !exist {
		const format = "certificate pool file \"%s\" is not exist\n"
		return errors.Errorf(format, mgr.dataPath)
	}
	// store password
	mgr.password = security.NewBytes(password)
	security.CoverBytes(password)
	// create backup
	err = mgr.createBackup()
	if err != nil {
		return errors.WithMessage(err, "failed to create backup")
	}
	err = mgr.load()
	if err != nil {
		return errors.WithMessage(err, "failed to load certificate pool")
	}
	return mgr.readCommandLoop()
}

func (mgr *Manager) readCommandLoop() error {
	mgr.prefix = prefixManager
	mgr.scanner = bufio.NewScanner(mgr.stdin)
	for {
		if mgr.closed {
			return nil
		}
		fmt.Printf("%s> ", mgr.prefix)
		// handle CTRL+C
		if !mgr.scanner.Scan() {
			mgr.scanner = bufio.NewScanner(mgr.stdin)
			fmt.Println()
			continue
		}
		// print test input content
		if mgr.testMode {
			fmt.Println(mgr.scanner.Text())
		}
		switch mgr.prefix {
		case prefixManager:
			mgr.manager()
		case prefixPublic:
			mgr.public()
		case prefixPrivate:
			mgr.private()
		case prefixPublicRootCA:
			mgr.publicRootCA()
		case prefixPublicClientCA:
			mgr.publicClientCA()
		case prefixPublicClient:
			mgr.publicClient()
		case prefixPrivateRootCA:
			mgr.privateRootCA()
		case prefixPrivateClientCA:
			mgr.privateClientCA()
		case prefixPrivateClient:
			mgr.privateClient()
		default:
			panic(fmt.Sprintf("unknown prefix: %s\n", mgr.prefix))
		}
	}
}

func (mgr *Manager) createBackup() error {
	data, err := os.ReadFile(mgr.dataPath)
	if err != nil {
		return err
	}
	return system.WriteFile(mgr.bakPath, data)
}

func (mgr *Manager) deleteBackup() error {
	return os.Remove(mgr.bakPath)
}

func (mgr *Manager) load() error {
	// read certificate pool file
	data, err := os.ReadFile(mgr.dataPath)
	if err != nil {
		return err
	}
	// get password
	password := mgr.password.Get()
	defer mgr.password.Put(password)
	// load certificate
	pool := certpool.NewPool()
	err = certmgr.LoadCtrlCertPool(pool, data, password)
	if err != nil {
		return err
	}
	mgr.pool = pool
	return nil
}

func (mgr *Manager) reload() {
	err := mgr.load()
	if err != nil {
		fmt.Printf("failed to reload certificate pool: %s\n", err)
	}
}

func (mgr *Manager) save() {
	err := mgr.saveCertPool()
	if err != nil {
		fmt.Printf("failed to save certificate pool: %s\n", err)
	}
}

func (mgr *Manager) saveCertPool() error {
	password := mgr.password.Get()
	defer mgr.password.Put(password)
	data, err := certmgr.SaveCtrlCertPool(mgr.pool, password)
	if err != nil {
		return err
	}
	return system.WriteFile(mgr.dataPath, data)
}

func (mgr *Manager) exit() {
	err := mgr.deleteBackup()
	if err != nil {
		fmt.Printf("failed to delete backup: %s\n", err)
	}
	mgr.closed = true
	fmt.Println("Bye!")
}

func (mgr *Manager) manager() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	if len(args) > 1 {
		fmt.Printf("unknown command: \"%s\"\n", cmd)
		return
	}
	switch args[0] {
	case "public":
		mgr.prefix = prefixPublic
	case "private":
		mgr.prefix = prefixPrivate
	case "help":
		fmt.Print(managerHelp)
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) public() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	if len(args) > 1 {
		fmt.Printf("unknown command: \"%s\"\n", cmd)
		return
	}
	switch args[0] {
	case "root-ca":
		mgr.prefix = prefixPublicRootCA
	case "client-ca":
		mgr.prefix = prefixPublicClientCA
	case "client":
		mgr.prefix = prefixPublicClient
	case "help":
		a := make([]interface{}, 4)
		for i := 0; i < 4; i++ {
			a[i] = "public"
		}
		fmt.Printf(locationHelpTemplate, a...)
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixManager
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) private() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	if len(args) > 1 {
		fmt.Printf("unknown command: \"%s\"\n", cmd)
		return
	}
	switch args[0] {
	case "root-ca":
		mgr.prefix = prefixPrivateRootCA
	case "client-ca":
		mgr.prefix = prefixPrivateClientCA
	case "client":
		mgr.prefix = prefixPrivateClient
	case "help":
		a := make([]interface{}, 4)
		for i := 0; i < 4; i++ {
			a[i] = "private"
		}
		fmt.Printf(locationHelpTemplate, a...)
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixManager
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

// -----------------------------------------Public Root CA-----------------------------------------

func (mgr *Manager) publicRootCA() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.publicRootCAList()
	case "print":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.publicRootCAPrint(args[1])
	case "add":
		if len(args) < 2 {
			fmt.Println("no certificate file")
			return
		}
		mgr.publicRootCAAdd(args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.publicRootCADelete(args[1])
	case "export":
		if len(args) < 3 {
			fmt.Println("no certificate id or export file name")
			return
		}
		mgr.publicRootCAExport(args[1], args[2])
	case "help":
		fmt.Printf(certHelpTemplate, "public/root-ca", "Root CA", "public")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPublic
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) publicRootCAPrint(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certs := mgr.pool.GetPublicRootCACerts()
	if i < 1 || i > len(certs) {
		fmt.Println("invalid certificate id")
		return
	}
	dumpCert(i, certs[i-1])
}

func (mgr *Manager) publicRootCAAdd(certFile string) {
	block, err := os.ReadFile(certFile) // #nosec
	if checkError(err) {
		return
	}
	certs, err := cert.ParseCertificatesPEM(block)
	if checkError(err) {
		return
	}
	for i := 0; i < len(certs); i++ {
		err = mgr.pool.AddPublicRootCACert(certs[i].Raw)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) publicRootCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePublicRootCACert(i)
	checkError(err)
}

func (mgr *Manager) publicRootCAExport(id, file string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, err := mgr.pool.ExportPublicRootCACert(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(file, certPEM)
	checkError(err)
}

func (mgr *Manager) publicRootCAList() {
	certs := mgr.pool.GetPublicRootCACerts()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i])
	}
}

// ----------------------------------------Public Client CA----------------------------------------

func (mgr *Manager) publicClientCA() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.publicClientCAList()
	case "add":
		if len(args) < 2 {
			fmt.Println("no certificate file")
			return
		}
		mgr.publicClientCAAdd(args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.publicClientCADelete(args[1])
	case "export":
		if len(args) < 3 {
			fmt.Println("no certificate id or export file name")
			return
		}
		mgr.publicClientCAExport(args[1], args[2])
	case "help":
		fmt.Printf(certHelpTemplate, "public/client-ca", "Client CA", "public")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPublic
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) publicClientCAList() {
	certs := mgr.pool.GetPublicClientCACerts()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i])
	}
}

func (mgr *Manager) publicClientCAAdd(certFile string) {
	pemData, err := os.ReadFile(certFile) // #nosec
	if checkError(err) {
		return
	}
	certs, err := cert.ParseCertificatesPEM(pemData)
	if checkError(err) {
		return
	}
	for i := 0; i < len(certs); i++ {
		err = mgr.pool.AddPublicClientCACert(certs[i].Raw)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) publicClientCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePublicClientCACert(i)
	checkError(err)
}

func (mgr *Manager) publicClientCAExport(id, file string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, err := mgr.pool.ExportPublicClientCACert(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(file, certPEM)
	checkError(err)
}

// -----------------------------------------Public Client------------------------------------------

func (mgr *Manager) publicClient() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.publicClientList()
	case "add":
		if len(args) < 3 {
			fmt.Println("no certificate file or private key file")
			return
		}
		mgr.publicClientAdd(args[1], args[2])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.publicClientDelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate id or two export file name")
			return
		}
		mgr.publicClientExport(args[1], args[2], args[3])
	case "help":
		fmt.Printf(certHelpTemplate, "public/client", "Client", "public")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPublic
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) publicClientList() {
	certs := mgr.pool.GetPublicClientPairs()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i].Certificate)
	}
}

func (mgr *Manager) publicClientAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPublicClientPair(certs[i].Raw, keyData)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) publicClientDelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePublicClientCert(i)
	checkError(err)
}

func (mgr *Manager) publicClientExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPublicClientPair(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err)
}

// ----------------------------------------Private Root CA-----------------------------------------

func (mgr *Manager) privateRootCA() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.privateRootCAList()
	case "add":
		if len(args) < 3 {
			fmt.Println("no certificate file or private key file")
			return
		}
		mgr.privateRootCAAdd(args[1], args[2])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.privateRootCADelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate id or two export file name")
			return
		}
		mgr.privateRootCAExport(args[1], args[2], args[3])
	case "help":
		fmt.Printf(certHelpTemplate, "private/root-ca", "Root CA", "private")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPrivate
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) privateRootCAList() {
	certs := mgr.pool.GetPrivateRootCACerts()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i])
	}
}

func (mgr *Manager) privateRootCAAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateRootCAPair(certs[i].Raw, keyData)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) privateRootCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePrivateRootCACert(i)
	checkError(err)
}

func (mgr *Manager) privateRootCAExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateRootCAPair(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err)
}

// ---------------------------------------Private Client CA----------------------------------------

func (mgr *Manager) privateClientCA() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.privateClientCAList()
	case "add":
		if len(args) < 3 {
			fmt.Println("no certificate file or private key file")
			return
		}
		mgr.privateClientCAAdd(args[1], args[2])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.privateClientCADelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate id or two export file name")
			return
		}
		mgr.privateClientCAExport(args[1], args[2], args[3])
	case "help":
		fmt.Printf(certHelpTemplate, "private/client-ca", "Client CA", "private")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPrivate
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) privateClientCAList() {
	certs := mgr.pool.GetPrivateClientCACerts()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i])
	}
}

func (mgr *Manager) privateClientCAAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateClientCAPair(certs[i].Raw, keyData)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) privateClientCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePrivateClientCACert(i)
	checkError(err)
}

func (mgr *Manager) privateClientCAExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateClientCAPair(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err)
}

// ----------------------------------------Private Client------------------------------------------

func (mgr *Manager) privateClient() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.privateClientList()
	case "add":
		if len(args) < 3 {
			fmt.Println("no certificate file or private key file")
			return
		}
		mgr.privateClientAdd(args[1], args[2])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate id")
			return
		}
		mgr.privateClientDelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate id or two export file name")
			return
		}
		mgr.privateClientExport(args[1], args[2], args[3])
	case "help":
		fmt.Printf(certHelpTemplate, "private/client", "Client", "private")
	case "save":
		mgr.save()
	case "reload":
		mgr.reload()
	case "return":
		mgr.prefix = prefixPrivate
	case "exit":
		mgr.exit()
	default:
		fmt.Printf("unknown command: \"%s\"\n", cmd)
	}
}

func (mgr *Manager) privateClientList() {
	certs := mgr.pool.GetPrivateClientPairs()
	for i := 0; i < len(certs); i++ {
		printCert(i, certs[i].Certificate)
	}
}

func (mgr *Manager) privateClientAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateClientPair(certs[i].Raw, keyData)
		if checkError(err) {
			return
		}
		fmt.Printf("\n%s\n\n", cert.Sdump(certs[i]))
	}
}

func (mgr *Manager) privateClientDelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	err = mgr.pool.DeletePrivateClientCert(i)
	checkError(err)
}

func (mgr *Manager) privateClientExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateClientPair(i)
	if checkError(err) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err)
}

func loadPairs(certFile, keyFile string) ([]*x509.Certificate, []interface{}) {
	certPEM, err := os.ReadFile(certFile) // #nosec
	if checkError(err) {
		return nil, nil
	}
	keyPEM, err := os.ReadFile(keyFile) // #nosec
	if checkError(err) {
		return nil, nil
	}
	certs, err := cert.ParseCertificatesPEM(certPEM)
	if checkError(err) {
		return nil, nil
	}
	keys, err := cert.ParsePrivateKeysPEM(keyPEM)
	if checkError(err) {
		return nil, nil
	}
	certsNum := len(certs)
	keysNum := len(keys)
	if certsNum != keysNum {
		const format = "%d certificates in %s but %d private keys in %s\n"
		fmt.Printf(format, certsNum, certFile, keysNum, keyFile)
		return nil, nil
	}
	return certs, keys
}

func printCert(id int, crt *x509.Certificate) {
	const format = "ID: %-3d %s\n"
	switch {
	case crt.Subject.CommonName != "":
		fmt.Printf(format, id, crt.Subject.CommonName)
	case len(crt.Subject.Organization) != 0:
		fmt.Printf(format, id, crt.Subject.Organization[0])
	default:
		fmt.Printf(format, id, crt.Subject)
	}
}

func dumpCert(id int, crt *x509.Certificate) {
	const format = "========================ID: %d========================\n%s\n"
	fmt.Printf(format, id, cert.Sdump(crt))
}

func checkError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}
