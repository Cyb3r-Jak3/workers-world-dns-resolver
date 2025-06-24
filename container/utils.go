package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/miekg/dns"
)

const JSONApplicationType = "application/json; charset=utf-8"

func JSONResponse(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", JSONApplicationType)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error. Unable to decode JSON response", http.StatusInternalServerError)
	}
}

type ParsedQuestion struct {
	Domain string
	Type   uint16
}

func ParseURLQuery(url *url.URL) (*ParsedQuestion, error) {
	parsed := &ParsedQuestion{}
	query := url.Query()
	if domain := query.Get("domain"); domain != "" {
		parsed.Domain = domain
	} else {
		return nil, fmt.Errorf("missing 'domain' parameter in query")
	}
	if typeStr := query.Get("type"); typeStr != "" {
		parsed.Type = dns.StringToType[typeStr]
		if parsed.Type == 0 {
			return nil, fmt.Errorf("invalid DNS type: %s", typeStr)
		}
	} else {
		return nil, fmt.Errorf("missing 'type' parameter in query")
	}
	return parsed, nil
}
