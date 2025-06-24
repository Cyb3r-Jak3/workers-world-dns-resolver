package main

import (
	"fmt"
	"time"
)

type DNSServer struct {
	Name    string
	Address string
	Port    int
}

type DNSServerResponse struct {
	DNSServer      string        `json:"server"`
	Values         []string      `json:"values"`
	Address        string        `json:"server_address"`
	TTL            int           `json:"ttl"`
	Duration       time.Duration `json:"duration"`
	DurationString string        `json:"duration_string"`
}

type LookupResponse struct {
	Question            string              `json:"question"`
	Type                string              `json:"type"`
	Answers             []DNSServerResponse `json:"answers"`
	Location            string              `json:"location"`
	Region              string              `json:"region"`
	Country             string              `json:"country"`
	TotalDuration       time.Duration       `json:"total_duration"`
	TotalDurationString string              `json:"total_duration_string"`
}

func (s *DNSServer) String() string {
	return s.Name + " (" + s.Address + ":" + fmt.Sprint(s.Port) + ")"
}
func (s *DNSServer) AddressString() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

var dnsServers = []DNSServer{
	{
		Name:    "Cloudflare",
		Address: "1.1.1.1",
		Port:    53,
	},
	{
		Name:    "Google",
		Address: "8.8.8.8",
		Port:    53,
	},
	{
		Name:    "OpenDNS",
		Address: "208.67.222.222",
		Port:    53,
	},
	{
		Name:    "Quad9",
		Address: "9.9.9.9",
		Port:    53,
	},
	{
		Name:    "Oracle",
		Address: "216.146.35.35",
		Port:    53,
	},
	{
		Name:    "WholeSale Internet",
		Address: "204.12.225.227",
		Port:    53,
	},
	{
		Name:    "Fortinet",
		Address: "208.91.112.53",
		Port:    53,
	},
	{
		Name:    "SkyDNS",
		Address: "195.46.39.39",
		Port:    53,
	},
	{
		Name:    "Liquid Telecommunications Ltd",
		Address: "5.11.11.5",
		Port:    53,
	},
	{
		Name:    "Tele2 Nederland B.V.",
		Address: "87.213.100.113",
		Port:    53,
	},
	{
		Name:    "Completel SAS",
		Address: "83.145.86.7",
		Port:    53,
	},
	{
		Name:    "Prioritytelecom Spain S.A",
		Address: "212.230.255.1",
		Port:    53,
	},
	{
		Name:    "nemox.net",
		Address: "83.137.41.9",
		Port:    53,
	},
	{
		Name:    "Universitaet Leipzig",
		Address: "139.18.25.33",
		Port:    53,
	},
	{
		Name:    "Vogel Solucoes em Telecom e Informatica S/A",
		Address: "189.126.192.4",
		Port:    53,
	},
	{
		Name:    "TT Dotcom Sdn Bhd",
		Address: "211.25.206.147",
		Port:    53,
	},
	{
		Name:    "Telstra Internet",
		Address: "139.130.4.4",
		Port:    53,
	},
	{
		Name:    "Global-Gateway Internet",
		Address: "122.56.107.86",
		Port:    53,
	},
	{
		Name:    "DigitalOcean LLC",
		Address: "139.59.219.245",
		Port:    53,
	},
	{
		Name:    "LG Dacom Corporation",
		Address: "164.124.101.2",
		Port:    53,
	},
	{
		Name:    "Kappa Internet Services Private Limited",
		Address: "115.178.96.2",
		Port:    53,
	},
	{
		Name:    "CMPak Limited",
		Address: "209.150.154.1",
		Port:    53,
	},
	{
		Name:    "Daniel Cid",
		Address: "185.228.168.9",
		Port:    53,
	},
	{
		Name:    "SS Online",
		Address: "103.80.1.2",
		Port:    53,
	},
	{
		Name:    "Alternate DNS",
		Address: "76.76.19.19",
		Port:    53,
	},
	{
		Name:    "CleanBrowsing",
		Address: "185.228.168.9",
		Port:    53,
	},
	{
		Name:    "Comodo Secure",
		Address: "8.26.56.26",
		Port:    53,
	},
	{
		Name:  "Comcast Xfinity DNS Servers",
		Address: "75.75.75.75",
		Port:    53,
	},
}
