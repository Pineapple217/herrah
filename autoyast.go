package main

import (
	"encoding/xml"
)

type Profile struct {
	XMLName      xml.Name        `xml:"profile"`
	Xmlns        string          `xml:"xmlns,attr"`
	XmlnsConf    string          `xml:"xmlns:config,attr"`
	General      General         `xml:"general"`
	Networking   Networking      `xml:"networking"`
	Users        Users           `xml:"users"`
	Services     ServicesManager `xml:"services-manager"`
	Firewall     Firewall        `xml:"firewall"`
	Partitioning Partitioning    `xml:"partitioning"`
	Software     Software        `xml:"software"`
}

type General struct {
	Clock      Clock `xml:"clock"`
	Networking struct {
		Hostname string `xml:"hostname"`
	} `xml:"networking"`
}

type Clock struct {
	HWClock  string `xml:"hwclock"`
	Timezone string `xml:"timezone"`
}

type Networking struct {
	Interfaces Interfaces `xml:"interfaces"`
	Routing    Routing    `xml:"routing"`
}

type Interfaces struct {
	Type       string      `xml:"config:type,attr"`
	Interfaces []Interface `xml:"interface"`
}

type Interface struct {
	Bootproto string `xml:"bootproto"`
	Device    string `xml:"device"`
	IPAddr    string `xml:"ipaddr"`
	Netmask   string `xml:"netmask"`
}

type Routing struct {
	Routes Routes `xml:"routes"`
}

type Routes struct {
	Type   string  `xml:"config:type,attr"`
	Routes []Route `xml:"route"`
}

type Route struct {
	Destination string `xml:"destination"`
	Gateway     string `xml:"gateway"`
}

type Users struct {
	Type  string `xml:"config:type,attr"`
	Users []User `xml:"user"`
}

type User struct {
	Username string  `xml:"username"`
	Password string  `xml:"user_password"`
	FullName string  `xml:"fullname,omitempty"`
	GID      int     `xml:"gid,omitempty"`
	Home     string  `xml:"home,omitempty"`
	Shell    string  `xml:"shell,omitempty"`
	Groups   *Groups `xml:"groups,omitempty"`
}

type Groups struct {
	Type   string   `xml:"config:type,attr"`
	Groups []string `xml:"group"`
}

type ServicesManager struct {
	Services Services `xml:"services"`
}

type Services struct {
	Type     string    `xml:"config:type,attr"`
	Services []Service `xml:"service"`
}

type Service struct {
	Name   string `xml:"service_name"`
	Action string `xml:"service_action"`
}

type Firewall struct {
	ServicesToOpen ServicesToOpen `xml:"services_to_open"`
}

type ServicesToOpen struct {
	Type     string   `xml:"config:type,attr"`
	Services []string `xml:"service"`
}

type Partitioning struct {
	Type   string  `xml:"config:type,attr"`
	Drives []Drive `xml:"drive"`
}

type Drive struct {
	Device     string      `xml:"device"`
	Initialize *BoolAttr   `xml:"initialize,omitempty"`
	Use        string      `xml:"use"`
	Partitions *Partitions `xml:"partitions,omitempty"`
}

type BoolAttr struct {
	Type  string `xml:"config:type,attr"`
	Value bool   `xml:",chardata"`
}

type Partitions struct {
	Type       string      `xml:"config:type,attr"`
	Partitions []Partition `xml:"partition"`
}

type Partition struct {
	Mount       string     `xml:"mount"`
	Size        string     `xml:"size"`
	Format      BoolAttr   `xml:"format"`
	Filesystem  SymbolAttr `xml:"filesystem"`
	PartitionID string     `xml:"partition_id,omitempty"`
	PartType    string     `xml:"partition_type"`
}

type SymbolAttr struct {
	Type  string `xml:"config:type,attr"`
	Value string `xml:",chardata"`
}

type Software struct {
	Patterns Patterns `xml:"patterns"`
}

type Patterns struct {
	Type     string   `xml:"config:type,attr"`
	Patterns []string `xml:"pattern"`
}

func GetAutoyast(m machine) ([]byte, error) {
	profile := Profile{
		Xmlns:     "http://www.suse.com/1.0/yast2ns",
		XmlnsConf: "http://www.suse.com/1.0/configns",
		General: General{
			Clock: Clock{HWClock: "UTC", Timezone: "Europe/Brussels"},
			Networking: struct {
				Hostname string "xml:\"hostname\""
			}{Hostname: m.Name},
		},
		Networking: Networking{
			Interfaces: Interfaces{
				Type: "list",
				Interfaces: []Interface{{
					Bootproto: "static",
					Device:    "eth0",
					IPAddr:    m.IP,
					Netmask:   "255.255.255.0",
				}},
			},
			Routing: Routing{
				Routes: Routes{
					Type: "list",
					Routes: []Route{{
						Destination: "default",
						Gateway:     "10.10.10.1",
					}},
				},
			},
		},
		Users: Users{
			Type: "list",
			Users: []User{
				{Username: "root", Password: "woorden"},
				{
					Username: "admin",
					Password: "woorden",
					FullName: "Admin User",
					GID:      100,
					Home:     "/home/admin",
					Shell:    "/bin/bash",
					Groups: &Groups{
						Type:   "list",
						Groups: []string{"wheel"},
					},
				},
			},
		},
		Services: ServicesManager{
			Services: Services{
				Type: "list",
				Services: []Service{{
					Name:   "sshd",
					Action: "enable",
				}},
			},
		},
		Firewall: Firewall{
			ServicesToOpen: ServicesToOpen{
				Type:     "list",
				Services: []string{"ssh"},
			},
		},
		Partitioning: Partitioning{
			Type: "list",
			Drives: []Drive{
				{
					Device:     "/dev/sda",
					Initialize: &BoolAttr{Type: "boolean", Value: true},
					Use:        "all",
					Partitions: &Partitions{
						Type: "list",
						Partitions: []Partition{
							{
								Mount:       "/boot/efi",
								Size:        "512M",
								Format:      BoolAttr{Type: "boolean", Value: true},
								Filesystem:  SymbolAttr{Type: "symbol", Value: "vfat"},
								PartitionID: "259",
								PartType:    "primary",
							},
							{
								Mount:      "swap",
								Size:       "4G",
								Format:     BoolAttr{Type: "boolean", Value: true},
								Filesystem: SymbolAttr{Type: "symbol", Value: "swap"},
								PartType:   "primary",
							},
							{
								Mount:      "/",
								Size:       "max",
								Format:     BoolAttr{Type: "boolean", Value: true},
								Filesystem: SymbolAttr{Type: "symbol", Value: "btrfs"},
								PartType:   "primary",
							},
						},
					},
				},
				{Device: "/dev/sdb", Use: "none"},
			},
		},
		Software: Software{
			Patterns: Patterns{
				Type:     "list",
				Patterns: []string{"base", "enhanced_base", "microos"},
			},
		},
	}

	xmlBytes, err := xml.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}

	result := append([]byte(`<?xml version="1.0"?>`+"\n"+"<!DOCTYPE profile>\n"), xmlBytes...)
	return result, nil
}
