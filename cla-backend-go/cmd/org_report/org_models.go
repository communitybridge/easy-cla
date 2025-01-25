// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package main

import "encoding/xml"

type DocusignEnvelopeInformation struct {
	XMLName        xml.Name       `xml:"DocuSignEnvelopeInformation"`
	EnvelopeStatus EnvelopeStatus `xml:"EnvelopeStatus"`
}

type EnvelopeStatus struct {
	RecipientStatuses RecipientStatuses `xml:"RecipientStatuses"`
}

type RecipientStatuses struct {
	RecipientStatus []RecipientStatus `xml:"RecipientStatus"`
}

type RecipientStatus struct {
	FormData FormData `xml:"FormData"`
}

type FormData struct {
	XFDF XFDF `xml:"xfdf"`
}

type XFDF struct {
	Fields Fields `xml:"fields"`
}

type Fields struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

// Struct to hold the final data
type CompanyData struct {
	CompanyID         string
	CompanyName       string
	CompanySFID       string
	CCLASignatures    int
	DateFirstSigned   string
	DateLastSigned    string
	CoporationAddress []string
	ClaManagers       []string
	ClaGroupNames     []string
}
