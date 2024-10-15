package main

const SCAN_REPORT_VERSION = 1

// NOTE: we will need to change datahub's server code
//		because it depends a lot on the presences of keys (but in golang we will have to set them to nil)

type ScanReport struct {
	Domain      string                `json:"domain"`
	Version     int                   `json:"version"`
	Tags        []string              `json:"tags"`
	Services    map[string]any        `json:"services"`
	HttpReports map[string]HttpReport `json:"http_reports"`
	Records     map[string][]string   `json:"records"`
	Whois       string                `json:"whois"`
	Meta        map[string]any        `json:"meta"`
}

type HttpReport struct {
	Tags         []string          `json:"tags"`
	Certificate  *CertificateDump  `json:"certificate"`
	Path         string            `json:"path"` // actual path after redirection (like login page or something like that)
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"` // actually we are losing headers with multiple value but the server doesn't support it
	Title        string            `json:"title"`
	HtmlMeta     []HtmlMeta        `json:"html_meta"`
	RobotTxt     []RobotDirective  `json:"robot_txt"`
	NodeInfoList []any             `json:"node_info_list"`
	NodeInfo     any               `json:"node_info"`
	Matrix       *MatrixMeta       `json:"matrix"`
}

type MatrixMeta struct {
	WellknownClient string   `json:"wellknown_client"`
	ClientVersions  []string `json:"client_versions"`
	ClientFeatures  []string `json:"client_features"`
}

type RobotDirective struct {
	UserAgent string `json:"useragent"`
	Directive string `json:"directive"`
	Data      string `json:"data"`
}

type HtmlMeta struct {
	Property string `json:"property"`
	Content  string `json:"content"`
}

type CertificateEntity struct {
	Rfc4514 string              `json:"rfc4514"`
	Attrs   map[string][]string `json:"attrs"`
}
type CertificateDump struct {
	Version struct {
		VersionName  string `json:"name"`
		VersionValue int    `json:"number"`
	} `json:"version"`
	Issuer      CertificateEntity `json:"issuer"`
	Subject     CertificateEntity `json:"subject"`
	ValidAfter  string            `json:"valid_after"`  // python's isoformat()[:19]
	ValidBefore string            `json:"valid_before"` // python's isoformat()[:19]
	PublicKey   string            `json:"public_key"`   // OpenSSH format
	DNSNames    []string          `json:"dns_names"`    // (SUBJECT_ALTERNATIVE_NAME)
	Raw         string            `json:"raw"`          // base64-encoded cert (without newlines and header/footer)
}

type OutdatedDomainResponse struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	ErrorMessage string   `json:"error_message"`
	Domains      []string `json:"data"`
}
