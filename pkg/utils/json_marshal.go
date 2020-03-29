package utils

import (
	"encoding/json"
	"gopkg.in/ldap.v3"
	"unicode/utf8"
)



type LDAPEntryJSON map[string]interface{}

func SearchResultToJSON(result *ldap.SearchResult) (jResponse []byte, err error) {
	var ldapResponsesJSON []LDAPEntryJSON
	for _, entry := range result.Entries {
		jEntry := make(LDAPEntryJSON)
		for _, attribute := range entry.Attributes {
			if len(attribute.Values) == 1 {
				jEntry[attribute.Name] = HandleLDAPBytes(attribute.Name, attribute.ByteValues[0])
			} else {
				var vals []interface{}
				for _, val := range attribute.ByteValues {
					vals = append(vals, HandleLDAPBytes(attribute.Name, val))
				}
				jEntry[attribute.Name] = vals
			}
		}
		ldapResponsesJSON = append(ldapResponsesJSON, jEntry)
	}
	return json.Marshal(ldapResponsesJSON)
}

// HandleLDAPBytes takes a byte slice from a raw attribute value and returns either a UTF8 string (if it's a string),
// or GUID or timestamp
func HandleLDAPBytes(name string, b []byte) interface{} {
	if name == "objectGUID" {
		g, err := WindowsGuidFromBytes(b); if err != nil {
			return b
		}
		return g
	}
	if name == "objectSid" {
		s, err := WindowsSIDFromBytes(b); if err != nil {
			return b
		}
		return s
	}

	if utf8.Valid(b) {
		s := string(b)
		if s == "9223372036854775807" { //max int64 size
			return nil //basically a no-value (e.g. never expires)
		}
		if NTFileTimeRegex.Match(b) {
			timeStamp, err := NTFileTimeToTimestamp(s)
			if err != nil {
				return s
			}
			return timeStamp
		}
		if ADLdapTimeRegex.Match(b) {
			timeStamp, err := ADLdapTimeToTimestamp(s)
			if err != nil {
				return s
			}
			return timeStamp
		}
		return s
	}
	return b
}

