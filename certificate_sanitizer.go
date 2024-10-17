package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"time"
)

func CertName2Entity(name pkix.Name) CertificateEntity {
	return CertificateEntity{
		Rfc4514: name.String(),
		Attrs: func() map[string][]string {
			ret := map[string][]string{}
			for _, name := range name.Names {
				oid := name.Type.String()
				if t, e := x509OidNames[oid]; e {
					oid = t
				}
				ret[oid] = []string{name.Value.(string)}
			}
			return ret
		}()}
}

func SanitizeCertificate(cert x509.Certificate) CertificateDump {
	return CertificateDump{
		Issuer:      CertName2Entity(cert.Issuer),
		Subject:     CertName2Entity(cert.Subject),
		Version:     cert.Version,
		DNSNames:    cert.DNSNames,
		Raw:         base64.StdEncoding.EncodeToString(cert.Raw),
		ValidAfter:  cert.NotBefore.Format(time.RFC3339),
		ValidBefore: cert.NotAfter.Format(time.RFC3339),
		// TODO: i don't know how to get the pubkey hash
		PublicKey: cert.PublicKeyAlgorithm.String() + " ",
	}
}
