//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	certificates "github.com/arangodb-helper/go-certificates"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// TLS valid for defaults
	defaultTLSValidFor   = time.Hour * 24 * 365 * 2  // 2 years
	defaultTLSCAValidFor = time.Hour * 24 * 365 * 15 // 15 years
	// Client authentication valid for defaults
	defaultClientAuthValidFor   = time.Hour * 24 * 365 * 1  // 1 years
	defaultClientAuthCAValidFor = time.Hour * 24 * 365 * 15 // 15 years
)

var (
	logFatal  func(error, string)
	showUsage func(cmd *cobra.Command, args []string)
	cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "Create certificates",
		Run:   cmdShowUsage,
	}
	cmdCreateJWTSecret = &cobra.Command{
		Use:   "jwt-secret",
		Short: "Create a JWT secret and save it in a plain text file",
		Run:   cmdCreateJWTSecretRun,
	}
	cmdCreateTLS = &cobra.Command{
		Use:   "tls",
		Short: "Create TLS certificates",
		Run:   cmdShowUsage,
	}
	cmdCreateTLSCA = &cobra.Command{
		Use:   "ca",
		Short: "Create a CA certificate used to sign TLS certificates",
		Run:   cmdCreateTLSCARun,
	}
	cmdCreateTLSKeyFile = &cobra.Command{
		Use:   "keyfile",
		Short: "Create a TLS certificate and save it as keyfile",
		Run:   cmdCreateTLSKeyFileRun,
	}
	cmdCreateTLSCertificate = &cobra.Command{
		Use:   "certificate",
		Short: "Create a TLS certificate and save it as crt, key files",
		Run:   cmdCreateTLSCertificateRun,
	}
	cmdCreateTLSKeystore = &cobra.Command{
		Use:   "keystore",
		Short: "Create a TLS certificate and save it as java keystore files",
		Run:   cmdCreateTLSKeystoreRun,
	}
	cmdCreateClientAuth = &cobra.Command{
		Use:   "client-auth",
		Short: "Create client authentication certificates",
		Run:   cmdShowUsage,
	}
	cmdCreateClientAuthCA = &cobra.Command{
		Use:   "ca",
		Short: "Create a CA certificate used to sign client authentication certificates",
		Run:   cmdCreateClientAuthCARun,
	}
	cmdCreateClientAuthKeyFile = &cobra.Command{
		Use:   "keyfile",
		Short: "Create a client authentication certificate and save it as keyfile",
		Run:   cmdCreateClientAuthKeyFileRun,
	}

	createOptions struct {
		jwtsecret createJWTSecretOptions
		tls       struct {
			ca          createCAOptions
			keyFile     createKeyFileOptions
			certificate createCertificateOptions
			keystore    createKeystoreOptions
		}
		clientAuth struct {
			ca      createCAOptions
			keyFile createKeyFileOptions
		}
	}
)

type createJWTSecretOptions struct {
	secretFile   string
	secretLength int
}

func (o *createJWTSecretOptions) ConfigureFlags(f *pflag.FlagSet) {
	f.StringVar(&o.secretFile, "secret", "secret.jwt", "Filename of the generated JWT secret")
	f.IntVar(&o.secretLength, "length", 32, "Number of bytes in the secret")
}

func (o *createJWTSecretOptions) CreateSecret() {
	if o.secretFile == "" {
		logFatal(nil, "--secret missing")
	}

	// Create secrey
	secret := make([]byte, o.secretLength)
	rand.Read(secret)

	// Encode as hex & store
	encoded := hex.EncodeToString(secret)
	mustWriteFile(encoded, o.secretFile, 0600, "secret")

	fmt.Printf("Created JWT secret in %s\n", o.secretFile)
	fmt.Println("Make sure to store this files in a secure location.")
}

type createCAOptions struct {
	certFile   string
	keyFile    string
	validFor   time.Duration
	ecdsaCurve string
}

func (o *createCAOptions) ConfigureFlags(f *pflag.FlagSet, defaultFName string, defaultValidFor time.Duration) {
	f.StringVar(&o.certFile, "cert", defaultFName+".crt", "Filename of the generated CA certificate")
	f.StringVar(&o.keyFile, "key", defaultFName+".key", "Filename of the generated CA private key")
	f.DurationVar(&o.validFor, "validfor", defaultValidFor, "Lifetime of the certificate until expiration")
	f.StringVar(&o.ecdsaCurve, "curve", "P521", "ECDSA curve used for private key")
}

func (o *createCAOptions) CreateCA() {
	// Create certificate
	options := certificates.CreateCertificateOptions{
		ValidFor:   o.validFor,
		ECDSACurve: o.ecdsaCurve,
		IsCA:       true,
	}
	cert, key, err := certificates.CreateCertificate(options, nil)
	if err != nil {
		logFatal(err, "Failed to create certificate")
	}

	// Save certificate
	mustWriteFile(cert, o.certFile, 0644, "cert")
	mustWriteFile(key, o.keyFile, 0600, "key")

	fmt.Printf("Created CA certificate & key in %s, %s\n", o.certFile, o.keyFile)
	fmt.Println("Make sure to store these files in a secure location.")
}

type createCertificateBaseOptions struct {
	caCertFile     string
	caKeyFile      string
	hosts          []string
	emailAddresses []string
	validFor       time.Duration
	ecdsaCurve     string
}

func (o *createCertificateBaseOptions) ConfigureFlags(f *pflag.FlagSet, defaultCAFName, defaultFName string, defaultValidFor time.Duration) {
	f.StringVar(&o.caCertFile, "cacert", defaultCAFName+".crt", "File containing TLS CA certificate")
	f.StringVar(&o.caKeyFile, "cakey", defaultCAFName+".key", "File containing TLS CA private key")
	f.StringSliceVar(&o.hosts, "host", nil, "Host name to include in the certificate")
	f.StringSliceVar(&o.emailAddresses, "email", nil, "Email address to include in the certificate")
	f.DurationVar(&o.validFor, "validfor", defaultValidFor, "Lifetime of the certificate until expiration")
	f.StringVar(&o.ecdsaCurve, "curve", "P521", "ECDSA curve used for private key")
}

// Create a certificate from given options.
// Returns: certificate content, key content, ca-certificate content
func (o *createCertificateBaseOptions) CreateCertificate(isClientAuth bool) (string, string, string) {
	// Load data
	caCert := mustReadFile(o.caCertFile, "cacert")
	caKey := mustReadFile(o.caKeyFile, "cakey")
	ca, err := certificates.LoadCAFromPEM(caCert, caKey)
	if err != nil {
		logFatal(err, "Failed to load CA")
	}

	// Create certificate
	options := certificates.CreateCertificateOptions{
		Hosts:          o.hosts,
		EmailAddresses: o.emailAddresses,
		ValidFor:       o.validFor,
		ECDSACurve:     o.ecdsaCurve,
		IsClientAuth:   isClientAuth,
	}
	cert, key, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		logFatal(err, "Failed to create certificate")
	}

	return cert, key, caCert
}

type createKeyFileOptions struct {
	createCertificateBaseOptions
	keyFile string
}

func (o *createKeyFileOptions) ConfigureFlags(f *pflag.FlagSet, defaultCAFName, defaultFName string, defaultValidFor time.Duration) {
	o.createCertificateBaseOptions.ConfigureFlags(f, defaultCAFName, defaultFName, defaultValidFor)
	f.StringVar(&o.keyFile, "keyfile", defaultFName+".keyfile", "Filename of keyfile to generate")
}

func (o *createKeyFileOptions) CreateKeyFile(isClientAuth bool) {
	// Create certificate + key
	cert, key, _ := o.createCertificateBaseOptions.CreateCertificate(isClientAuth)

	// Save certificate
	mustWriteKeyFile(cert, key, o.keyFile, "keyfile")

	fmt.Printf("Created keyfile in %s\n", o.keyFile)
	fmt.Println("Make sure to store this file in a secure location.")
}

type createCertificateOptions struct {
	createCertificateBaseOptions
	certFile string
	keyFile  string
}

func (o *createCertificateOptions) ConfigureFlags(f *pflag.FlagSet, defaultCAFName, defaultFName string, defaultValidFor time.Duration) {
	o.createCertificateBaseOptions.ConfigureFlags(f, defaultCAFName, defaultFName, defaultValidFor)
	f.StringVar(&o.certFile, "cert", defaultFName+".crt", "Filename of the generated certificate")
	f.StringVar(&o.keyFile, "key", defaultFName+".key", "Filename of the generated private key")
}

func (o *createCertificateOptions) CreateCertificate(isClientAuth bool) {
	// Create certificate + key
	cert, key, _ := o.createCertificateBaseOptions.CreateCertificate(isClientAuth)

	// Save certificate
	mustWriteFile(cert, o.certFile, 0644, "cert")
	mustWriteFile(key, o.keyFile, 0600, "key")

	fmt.Printf("Created certificate & key in %s, %s\n", o.certFile, o.keyFile)
	fmt.Println("Make sure to store these files in a secure location.")
}

type createKeystoreOptions struct {
	createCertificateBaseOptions
	keystoreFile     string
	keystorePassword string
	alias            string
}

func (o *createKeystoreOptions) ConfigureFlags(f *pflag.FlagSet, defaultCAFName, defaultFName string, defaultValidFor time.Duration) {
	o.createCertificateBaseOptions.ConfigureFlags(f, defaultCAFName, defaultFName, defaultValidFor)
	f.StringVar(&o.keystoreFile, "keystore", defaultFName+".jks", "Filename of the generated keystore")
	f.StringVar(&o.keystorePassword, "keystore-password", "", "Password of the generated keystore")
	f.StringVar(&o.alias, "alias", "", "Aliases use to store the certificate under in the keystore")
}

func (o *createKeystoreOptions) CreateCertificate(isClientAuth bool) {
	if o.alias == "" {
		logFatal(nil, "--alias missing")
	}
	if o.keystorePassword == "" {
		logFatal(nil, "--keystore-password missing")
	}
	// Create certificate + key
	cert, key, caCert := o.createCertificateBaseOptions.CreateCertificate(isClientAuth)

	// Encode as keystore
	ks, err := certificates.CreateKeystore(cert, key, caCert, o.alias, []byte(o.keystorePassword))
	if err != nil {
		logFatal(err, "Failed to create keystore")
	}
	mustWriteFile(string(ks), o.keystoreFile, 0600, "keystore")

	fmt.Printf("Created keystore in %s\n", o.keystoreFile)
	fmt.Println("Make sure to store this files in a secure location.")
}

// AddCommands adds all creations commands to the given root command.
func AddCommands(cmd *cobra.Command, logFatalFunc func(error, string), showUsageFunc func(cmd *cobra.Command, args []string)) {
	logFatal = logFatalFunc
	showUsage = showUsageFunc

	cmd.AddCommand(cmdCreate)
	cmdCreate.AddCommand(cmdCreateJWTSecret)
	cmdCreate.AddCommand(cmdCreateTLS)
	cmdCreateTLS.AddCommand(cmdCreateTLSCA)
	cmdCreateTLS.AddCommand(cmdCreateTLSKeyFile)
	cmdCreateTLS.AddCommand(cmdCreateTLSCertificate)
	cmdCreateTLS.AddCommand(cmdCreateTLSKeystore)
	cmdCreate.AddCommand(cmdCreateClientAuth)
	cmdCreateClientAuth.AddCommand(cmdCreateClientAuthCA)
	cmdCreateClientAuth.AddCommand(cmdCreateClientAuthKeyFile)

	createOptions.jwtsecret.ConfigureFlags(cmdCreateJWTSecret.Flags())
	createOptions.tls.ca.ConfigureFlags(cmdCreateTLSCA.Flags(), "tls-ca", defaultTLSCAValidFor)
	createOptions.tls.keyFile.ConfigureFlags(cmdCreateTLSKeyFile.Flags(), "tls-ca", "tls", defaultTLSValidFor)
	createOptions.tls.certificate.ConfigureFlags(cmdCreateTLSCertificate.Flags(), "tls-ca", "tls", defaultTLSValidFor)
	createOptions.tls.keystore.ConfigureFlags(cmdCreateTLSKeystore.Flags(), "tls-ca", "tls", defaultTLSValidFor)
	createOptions.clientAuth.ca.ConfigureFlags(cmdCreateClientAuthCA.Flags(), "client-auth-ca", defaultClientAuthCAValidFor)
	createOptions.clientAuth.keyFile.ConfigureFlags(cmdCreateClientAuthKeyFile.Flags(), "client-auth-ca", "client-auth", defaultClientAuthValidFor)
}

// Cobra run function using the usage of the given command
func cmdShowUsage(cmd *cobra.Command, args []string) {
	showUsage(cmd, args)
}

// cmdCreateJWTSecretRun creates a JWT secret and writes it to file
func cmdCreateJWTSecretRun(cmd *cobra.Command, args []string) {
	createOptions.jwtsecret.CreateSecret()
}

// cmdCreateTLSCARun creates a CA used to sign TLS certificates
func cmdCreateTLSCARun(cmd *cobra.Command, args []string) {
	createOptions.tls.ca.CreateCA()
}

// cmdCreateTLSKeyFileRun creates a TLS certificate and save it as keyfile
func cmdCreateTLSKeyFileRun(cmd *cobra.Command, args []string) {
	isClientAuth := false
	createOptions.tls.keyFile.CreateKeyFile(isClientAuth)
}

// cmdCreateTLSCertificateRun creates a TLS certificate and save it as crt+key file
func cmdCreateTLSCertificateRun(cmd *cobra.Command, args []string) {
	isClientAuth := false
	createOptions.tls.certificate.CreateCertificate(isClientAuth)
}

// cmdCreateTLSKeystoreRun creates a TLS certificate and save it as java keystore file.
func cmdCreateTLSKeystoreRun(cmd *cobra.Command, args []string) {
	isClientAuth := false
	createOptions.tls.keystore.CreateCertificate(isClientAuth)
}

// cmdCreateClientAuthCARun creates a CA used to sign client authentication certificates
func cmdCreateClientAuthCARun(cmd *cobra.Command, args []string) {
	createOptions.clientAuth.ca.CreateCA()
}

// cmdCreateClientAuthKeyFileRun creates a client authentication certificate and save it as keyfile
func cmdCreateClientAuthKeyFileRun(cmd *cobra.Command, args []string) {
	isClientAuth := true
	createOptions.clientAuth.keyFile.CreateKeyFile(isClientAuth)
}

func mustWriteKeyFile(cert, key string, filename string, flagName string) {
	if filename == "" {
		logFatal(nil, fmt.Sprintf("Missing filename for option --%s", flagName))
	}
	if err := certificates.SaveKeyFile(cert, key, filename); err != nil {
		logFatal(err, fmt.Sprintf("Failed to write %s", filename))
	}
}

func mustWriteFile(content string, filename string, mode os.FileMode, flagName string) {
	if filename == "" {
		logFatal(nil, fmt.Sprintf("Missing filename for option --%s", flagName))
	}
	folder := filepath.Dir(filename)
	if folder != "" {
		os.MkdirAll(folder, 0755)
	}
	if err := ioutil.WriteFile(filename, []byte(content), mode); err != nil {
		logFatal(err, fmt.Sprintf("Failed to write %s", filename))
	}
}

func mustReadFile(filename string, flagName string) string {
	if filename == "" {
		logFatal(nil, fmt.Sprintf("Missing filename for option --%s", flagName))
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		logFatal(err, fmt.Sprintf("Failed to read %s", filename))
	}
	return string(content)
}
