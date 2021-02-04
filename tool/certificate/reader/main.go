package main

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"os"

	"project/internal/cert"
	"project/internal/cert/certpool"
	"project/internal/system"
)

func main() {
	// load certificates
	pool, err := certpool.System()
	system.CheckError(err)
	certs := pool.Certs()
	l := len(certs)
	// print certificate information
	for i := 0; i < l; i++ {
		cert.Dump(certs[i])
		fmt.Println("================================================")
	}
	// encode certificates to pem
	fmt.Println("------------------------------------------------")
	buf := new(bytes.Buffer)
	for i := 0; i < l; i++ {
		block := pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certs[i].Raw,
		}
		err = pem.Encode(buf, &block)
		system.CheckError(err)
		// print certificate version and subject
		crt := certs[i]
		const format = "V%d %s\n"
		switch {
		case crt.Subject.CommonName != "":
			fmt.Printf(format, crt.Version, crt.Subject.CommonName)
		case len(crt.Subject.Organization) != 0:
			fmt.Printf(format, crt.Version, crt.Subject.Organization[0])
		default:
			fmt.Printf(format, crt.Version, crt.Subject)
		}
	}
	fmt.Println("------------------------------------------------")
	fmt.Println("the number of the system CA certificates:", l)
	// write pem
	err = system.WriteFile("system.pem", buf.Bytes())
	system.CheckError(err)
	// test certificates
	pemData, err := os.ReadFile("system.pem")
	system.CheckError(err)
	certs, err = cert.ParseCertificatesPEM(pemData)
	system.CheckError(err)
	// compare
	loadNum := len(certs)
	if loadNum == l {
		fmt.Println("export System CA certificates successfully")
	} else {
		fmt.Printf("warning: system: %d, test load: %d\n", l, loadNum)
	}
}
