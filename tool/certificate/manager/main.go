package main

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/term"

	"project/internal/cert"
	"project/internal/cert/certmgr"
	"project/internal/security"
	"project/internal/system"
)

const backupFilePath = certmgr.CertPoolFilePath + ".bak"

var stdinFD = int(syscall.Stdin)

func main() {
	var (
		init  bool
		reset bool
	)
	flag.BoolVar(&init, "init", false, "initialize certificate manager")
	flag.BoolVar(&reset, "reset", false, "reset certificate manager password")
	flag.Parse()
	switch {
	case init:
		initialize()
	case reset:
		resetPassword()
	default:
		manage()
	}
}

func initialize() {
	// check data file is exists
	exist, err := system.IsExist(certmgr.CertPoolFilePath)
	checkError(err, true)
	if exist {
		const format = "certificate pool file \"%s\" is already exists\n"
		fmt.Printf(format, certmgr.CertPoolFilePath)
		os.Exit(1)
	}
	// input password
	fmt.Print("password: ")
	password, err := term.ReadPassword(stdinFD)
	checkError(err, true)
	for {
		fmt.Print("\nretype: ")
		retype, err := term.ReadPassword(stdinFD)
		checkError(err, true)
		if !bytes.Equal(password, retype) {
			fmt.Print("\ndifferent password")
		} else {
			fmt.Println()
			break
		}
	}
	// load system certificates
	pool, err := cert.NewPoolWithSystemCerts()
	checkError(err, true)
	// create Root CA certificate
	rootCA, err := cert.GenerateCA(nil)
	checkError(err, true)
	err = pool.AddPrivateRootCAPair(rootCA.Encode())
	checkError(err, true)
	// create Client CA certificate
	clientCA, err := cert.GenerateCA(nil)
	checkError(err, true)
	err = pool.AddPrivateClientCAPair(clientCA.Encode())
	checkError(err, true)
	// generate a client certificate and use client CA sign it
	clientCert, err := cert.Generate(clientCA.Certificate, clientCA.PrivateKey, nil)
	checkError(err, true)
	err = pool.AddPrivateClientPair(clientCert.Encode())
	checkError(err, true)
	// save certificate pool
	err = certmgr.SaveCtrlCertPool(pool, password)
	checkError(err, true)
	fmt.Println("initialize certificate manager successfully")
}

func resetPassword() {
	// input old password
	fmt.Print("input old password: ")
	oldPwd, err := term.ReadPassword(stdinFD)
	checkError(err, true)
	fmt.Println()
	defer security.CoverBytes(oldPwd)
	// input new password
	fmt.Print("input new password: ")
	newPwd1, err := term.ReadPassword(stdinFD)
	checkError(err, true)
	fmt.Println()
	defer security.CoverBytes(newPwd1)
	fmt.Print("retype: ")
	newPwd2, err := term.ReadPassword(stdinFD)
	checkError(err, true)
	fmt.Println()
	defer security.CoverBytes(newPwd2)
	if !bytes.Equal(newPwd1, newPwd2) {
		fmt.Println("different password")
		os.Exit(1)
	}
	// load certificate pool
	certPool, err := ioutil.ReadFile(certmgr.CertPoolFilePath)
	checkError(err, true)
	pool := cert.NewPool()
	err = certmgr.LoadCtrlCertPool(pool, certPool, oldPwd)
	checkError(err, true)
	// save certificate pool
	err = certmgr.SaveCtrlCertPool(pool, newPwd1)
	checkError(err, true)
	fmt.Println("reset certificate manager password successfully")
}

func manage() {
	// input password
	fmt.Print("password: ")
	password, err := term.ReadPassword(stdinFD)
	checkError(err, true)
	fmt.Println()
	// backup
	createBackup()
	// start manage
	mgr := manager{
		password: security.NewBytes(password),
	}
	security.CoverBytes(password)
	mgr.Manage()
}

func createBackup() {
	data, err := ioutil.ReadFile(certmgr.CertPoolFilePath)
	checkError(err, true)
	err = system.WriteFile(backupFilePath, data)
	checkError(err, true)
}

func deleteBackup() {
	err := os.Remove(backupFilePath)
	checkError(err, true)
}

func printCertificate(id int, c *x509.Certificate) {
	fmt.Printf("ID: %d\n%s\n\n", id, cert.Print(c))
}

func loadPairs(certFile, keyFile string) ([]*x509.Certificate, []interface{}) {
	certPEM, err := ioutil.ReadFile(certFile) // #nosec
	if checkError(err, false) {
		return nil, nil
	}
	keyPEM, err := ioutil.ReadFile(keyFile) // #nosec
	if checkError(err, false) {
		return nil, nil
	}
	certs, err := cert.ParseCertificates(certPEM)
	if checkError(err, false) {
		return nil, nil
	}
	keys, err := cert.ParsePrivateKeys(keyPEM)
	if checkError(err, false) {
		return nil, nil
	}
	certsNum := len(certs)
	keysNum := len(keys)
	if certsNum != keysNum {
		const format = "%d certificates in %s and %d private keys in %s\n"
		fmt.Printf(format, certsNum, certFile, keysNum, keyFile)
		return nil, nil
	}
	return certs, keys
}

func checkError(err error, exit bool) bool {
	if err != nil {
		if err != io.EOF {
			fmt.Println(err)
		}
		if exit {
			os.Exit(1)
		}
		return true
	}
	return false
}

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

const locationHelpTemplate = `
help about manager/%s:
  
  root-ca      switch to %s/root-ca mode
  client-ca    switch to %s/client-ca mode
  client       switch to %s/client mode
  help         print help
  save         save certificate pool
  reload       reload certificate pool
  return       return to the manager
  exit         close certificate manager
  
`

const certHelpTemplate = `
help about manager/%s:
  
  list         list all %s certificates
  add          add a certificate
                command: add "cert.pem" ["key.pem"]
  delete       delete a certificate with ID
                command: delete 0
  export       export certificate and private key with ID
                command: export 0 "cert1.pem" ["key1.pem"]
  help         print help
  save         save certificate pool
  reload       reload certificate pool
  return       return to the %s mode
  exit         close certificate manager

`

type manager struct {
	password *security.Bytes
	pool     *cert.Pool
	prefix   string
	scanner  *bufio.Scanner
}

func (mgr *manager) Manage() {
	// interrupt input
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
	}()
	mgr.reload()
	mgr.prefix = prefixManager
	mgr.scanner = bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s> ", mgr.prefix)
		if !mgr.scanner.Scan() {
			mgr.scanner = bufio.NewScanner(os.Stdin)
			fmt.Println()
			continue
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
			fmt.Printf("unknown prefix: %s\n", mgr.prefix)
			os.Exit(1)
		}
	}
}

func (mgr *manager) reload() {
	// read certificate pool file
	certPool, err := ioutil.ReadFile(certmgr.CertPoolFilePath)
	checkError(err, true)
	// get password
	password := mgr.password.Get()
	defer mgr.password.Put(password)
	// load certificate
	pool := cert.NewPool()
	err = certmgr.LoadCtrlCertPool(pool, certPool, password)
	checkError(err, true)
	mgr.pool = pool
}

func (mgr *manager) save() {
	// get password
	password := mgr.password.Get()
	defer mgr.password.Put(password)
	// save certificate
	err := certmgr.SaveCtrlCertPool(mgr.pool, password)
	checkError(err, false)
}

func (mgr *manager) exit() {
	deleteBackup()
	fmt.Println("Bye!")
	os.Exit(0)
}

func (mgr *manager) manager() {
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
		mgr.managerHelp()
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

func (mgr *manager) managerHelp() {
	const help = `
help about manager:
  
  public       switch to public mode
  private      switch to private mode
  help         print help
  save         save certificate pool
  reload       reload certificate pool
  exit         close certificate manager
  
`
	fmt.Print(help)
}

func (mgr *manager) public() {
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

func (mgr *manager) private() {
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

func (mgr *manager) publicRootCA() {
	cmd := mgr.scanner.Text()
	args := system.CommandLineToArgv(cmd)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "list":
		mgr.publicRootCAList()
	case "add":
		if len(args) < 2 {
			fmt.Println("no certificate file")
			return
		}
		mgr.publicRootCAAdd(args[1])
	case "delete":
		if len(args) < 2 {
			fmt.Println("no certificate ID")
			return
		}
		mgr.publicRootCADelete(args[1])
	case "export":
		if len(args) < 3 {
			fmt.Println("no certificate ID or export file name")
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

func (mgr *manager) publicRootCAList() {
	fmt.Println()
	certs := mgr.pool.GetPublicRootCACerts()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i])
	}
}

func (mgr *manager) publicRootCAAdd(certFile string) {
	pemData, err := ioutil.ReadFile(certFile) // #nosec
	if checkError(err, false) {
		return
	}
	certs, err := cert.ParseCertificates(pemData)
	if checkError(err, false) {
		return
	}
	for i := 0; i < len(certs); i++ {
		err = mgr.pool.AddPublicRootCACert(certs[i].Raw)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) publicRootCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePublicRootCACert(i)
	checkError(err, false)
}

func (mgr *manager) publicRootCAExport(id, file string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, err := mgr.pool.ExportPublicRootCACert(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(file, certPEM)
	checkError(err, false)
}

// ----------------------------------------Public Client CA----------------------------------------

func (mgr *manager) publicClientCA() {
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
			fmt.Println("no certificate ID")
			return
		}
		mgr.publicClientCADelete(args[1])
	case "export":
		if len(args) < 3 {
			fmt.Println("no certificate ID or export file name")
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

func (mgr *manager) publicClientCAList() {
	fmt.Println()
	certs := mgr.pool.GetPublicClientCACerts()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i])
	}
}

func (mgr *manager) publicClientCAAdd(certFile string) {
	pemData, err := ioutil.ReadFile(certFile) // #nosec
	if checkError(err, false) {
		return
	}
	certs, err := cert.ParseCertificates(pemData)
	if checkError(err, false) {
		return
	}
	for i := 0; i < len(certs); i++ {
		err = mgr.pool.AddPublicClientCACert(certs[i].Raw)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) publicClientCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePublicClientCACert(i)
	checkError(err, false)
}

func (mgr *manager) publicClientCAExport(id, file string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, err := mgr.pool.ExportPublicClientCACert(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(file, certPEM)
	checkError(err, false)
}

// -----------------------------------------Public Client------------------------------------------

func (mgr *manager) publicClient() {
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
			fmt.Println("no certificate ID")
			return
		}
		mgr.publicClientDelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate ID or two export file name")
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

func (mgr *manager) publicClientList() {
	fmt.Println()
	certs := mgr.pool.GetPublicClientPairs()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i].Certificate)
	}
}

func (mgr *manager) publicClientAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPublicClientPair(certs[i].Raw, keyData)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) publicClientDelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePublicClientCert(i)
	checkError(err, false)
}

func (mgr *manager) publicClientExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPublicClientPair(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err, false)
}

// ----------------------------------------Private Root CA-----------------------------------------

func (mgr *manager) privateRootCA() {
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
			fmt.Println("no certificate ID")
			return
		}
		mgr.privateRootCADelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate ID or two export file name")
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

func (mgr *manager) privateRootCAList() {
	fmt.Println()
	certs := mgr.pool.GetPrivateRootCACerts()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i])
	}
}

func (mgr *manager) privateRootCAAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateRootCAPair(certs[i].Raw, keyData)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) privateRootCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePrivateRootCACert(i)
	checkError(err, false)
}

func (mgr *manager) privateRootCAExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateRootCAPair(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err, false)
}

// ---------------------------------------Private Client CA----------------------------------------

func (mgr *manager) privateClientCA() {
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
			fmt.Println("no certificate ID")
			return
		}
		mgr.privateClientCADelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate ID or two export file name")
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

func (mgr *manager) privateClientCAList() {
	fmt.Println()
	certs := mgr.pool.GetPrivateClientCACerts()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i])
	}
}

func (mgr *manager) privateClientCAAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateClientCAPair(certs[i].Raw, keyData)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) privateClientCADelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePrivateClientCACert(i)
	checkError(err, false)
}

func (mgr *manager) privateClientCAExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateClientCAPair(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err, false)
}

// ----------------------------------------Private Client------------------------------------------

func (mgr *manager) privateClient() {
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
			fmt.Println("no certificate ID")
			return
		}
		mgr.privateClientDelete(args[1])
	case "export":
		if len(args) < 4 {
			fmt.Println("no certificate ID or two export file name")
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

func (mgr *manager) privateClientList() {
	fmt.Println()
	certs := mgr.pool.GetPrivateClientPairs()
	for i := 0; i < len(certs); i++ {
		printCertificate(i, certs[i].Certificate)
	}
}

func (mgr *manager) privateClientAdd(certFile, keyFile string) {
	certs, keys := loadPairs(certFile, keyFile)
	for i := 0; i < len(certs); i++ {
		keyData, _ := x509.MarshalPKCS8PrivateKey(keys[i])
		err := mgr.pool.AddPrivateClientPair(certs[i].Raw, keyData)
		checkError(err, false)
		fmt.Printf("\n%s\n\n", cert.Print(certs[i]))
	}
}

func (mgr *manager) privateClientDelete(id string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	err = mgr.pool.DeletePrivateClientCert(i)
	checkError(err, false)
}

func (mgr *manager) privateClientExport(id, cert, key string) {
	i, err := strconv.Atoi(id)
	if checkError(err, false) {
		return
	}
	certPEM, keyPEM, err := mgr.pool.ExportPrivateClientPair(i)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(cert, certPEM)
	if checkError(err, false) {
		return
	}
	err = system.WriteFile(key, keyPEM)
	checkError(err, false)
}
