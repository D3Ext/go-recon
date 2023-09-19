package main

import (
	"fmt"
	"github.com/D3Ext/go-recon/pkg/gorecon"
	"log"
)

func main() {

	results, err := gorecon.Whois("example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(results.Domain.Name)
	fmt.Println(results.Domain.NameServers)
	fmt.Println(results.Domain.ID)

	// print other fields

	/*
		type WhoisInfo struct {
			Domain         *Domain  `json:"domain,omitempty"`
			Registrar      *Contact `json:"registrar,omitempty"`
			Registrant     *Contact `json:"registrant,omitempty"`
			Administrative *Contact `json:"administrative,omitempty"`
			Technical      *Contact `json:"technical,omitempty"`
			Billing        *Contact `json:"billing,omitempty"`
		}

		type Domain struct {
			ID                   string     `json:"id,omitempty"`
			Domain               string     `json:"domain,omitempty"`
			Punycode             string     `json:"punycode,omitempty"`
			Name                 string     `json:"name,omitempty"`
			Extension            string     `json:"extension,omitempty"`
			WhoisServer          string     `json:"whois_server,omitempty"`
			Status               []string   `json:"status,omitempty"`
			NameServers          []string   `json:"name_servers,omitempty"`
			DNSSec               bool       `json:"dnssec,omitempty"`
			CreatedDate          string     `json:"created_date,omitempty"`
			CreatedDateInTime    *time.Time `json:"created_date_in_time,omitempty"`
			UpdatedDate          string     `json:"updated_date,omitempty"`
			UpdatedDateInTime    *time.Time `json:"updated_date_in_time,omitempty"`
			ExpirationDate       string     `json:"expiration_date,omitempty"`
			ExpirationDateInTime *time.Time `json:"expiration_date_in_time,omitempty"`
		}

		type Contact struct {
			ID           string `json:"id,omitempty"`
			Name         string `json:"name,omitempty"`
			Organization string `json:"organization,omitempty"`
			Street       string `json:"street,omitempty"`
			City         string `json:"city,omitempty"`
			Province     string `json:"province,omitempty"`
			PostalCode   string `json:"postal_code,omitempty"`
			Country      string `json:"country,omitempty"`
			Phone        string `json:"phone,omitempty"`
			PhoneExt     string `json:"phone_ext,omitempty"`
			Fax          string `json:"fax,omitempty"`
			FaxExt       string `json:"fax_ext,omitempty"`
			Email        string `json:"email,omitempty"`
			ReferralURL  string `json:"referral_url,omitempty"`
		}
	*/
}
