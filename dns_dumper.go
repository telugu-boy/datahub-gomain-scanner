package main

import (
	"fmt"

	"github.com/miekg/dns"
)

var dnsclient = new(dns.Client)

func EnumerateDnsRecords(domain string) map[string][]string {
	ret := map[string][]string{
		"A": {}, "AAAA": {},
		"MX": {}}

	// https://en.wikipedia.org/wiki/List_of_DNS_record_types
	// A, AAAA, MX
	dnsmsg := new(dns.Msg)
	for _, qtyp := range []uint16{1, 28, 15} {
		dnsmsg.SetQuestion(dns.Fqdn(domain), qtyp)

		in, _, _ := dnsclient.Exchange(dnsmsg, DNSServer)

		if in == nil || len(in.Answer) == 0 {
			continue
		}

		for _, answer := range in.Answer {
			typ_name := dns.TypeToString[answer.Header().Rrtype]
			l := ret[typ_name]
			switch vv := answer.(type) {
			case *dns.A:
				l = append(l, vv.A.String())
			case *dns.AAAA:
				l = append(l, vv.AAAA.String())
			case *dns.MX:
				l = append(l, vv.Mx+" "+fmt.Sprint(vv.Preference))
			default:
				continue
			}

			ret[typ_name] = l
		}
	}

	return ret
}
