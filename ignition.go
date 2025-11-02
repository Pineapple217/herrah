package main

import "golang.org/x/crypto/bcrypt"

type IgnitionConfig struct {
	Ignition Ignition `json:"ignition"`
	Passwd   Passwd   `json:"passwd"`
	Storage  Storage  `json:"storage"`
	Systemd  Systemd  `json:"systemd"`
}

type Ignition struct {
	Version string `json:"version"`
}

type Passwd struct {
	Users []User `json:"users"`
}

type User struct {
	Name         string `json:"name"`
	PasswordHash string `json:"passwordHash"`
}

type Storage struct {
	Filesystems []Filesystem `json:"filesystems"`
	Files       []File       `json:"files"`
}

type Filesystem struct {
	Device         string   `json:"device"`
	Format         string   `json:"format"`
	MountOptions   []string `json:"mountOptions,omitempty"`
	Path           string   `json:"path"`
	WipeFilesystem bool     `json:"wipeFilesystem"`
}

type File struct {
	Path      string   `json:"path"`
	Mode      int      `json:"mode"`
	Overwrite bool     `json:"overwrite"`
	Contents  Contents `json:"contents"`
}

type Contents struct {
	Source    string `json:"source"`
	HumanRead string `json:"human_read,omitempty"`
}

type Systemd struct {
	Units []Unit `json:"units"`
}

type Unit struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func GetIgnitionConfig(m machine) (IgnitionConfig, error) {
	rootPasswordHash, err := bcrypt.GenerateFromPassword([]byte("root"), 10)
	if err != nil {
		return IgnitionConfig{}, err
	}
	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte("admin"), 10)
	if err != nil {
		return IgnitionConfig{}, err
	}

	root := IgnitionConfig{
		Ignition: Ignition{Version: "3.2.0"},
		Passwd: Passwd{
			Users: []User{
				{Name: "root", PasswordHash: string(rootPasswordHash)},
				{Name: "admin", PasswordHash: string(adminPasswordHash)},
			},
		},
		Storage: Storage{
			Filesystems: []Filesystem{
				{
					Device:         "/dev/disk/by-label/ROOT",
					Format:         "btrfs",
					MountOptions:   []string{"subvol=/@/home"},
					Path:           "/home",
					WipeFilesystem: false,
				},
			},
			Files: []File{
				{
					Path:      "/etc/hostname",
					Mode:      420,
					Overwrite: true,
					Contents:  Contents{Source: "data:," + m.Name},
				},
				{
					Path:      "/etc/locale.conf",
					Mode:      420,
					Overwrite: true,
					Contents: Contents{
						Source:    "data:text/plain;charset=utf-8;base64,TEFORz1lbl9HQi5VVEYtOAo=",
						HumanRead: "LANG=en_GB.UTF-8\n",
					},
				},
				{
					Path:      "/etc/NetworkManager/system-connections/em1.nmconnection",
					Mode:      384,
					Overwrite: true,
					Contents: Contents{
						Source:    "data:text/plain;charset=utf-8;base64,Cltjb25uZWN0aW9uXQppZD1lbTEKdHlwZT1ldGhlcm5ldAppbnRlcmZhY2UtbmFtZT1lbTEKCltpcHY0XQpkbnMtc2VhcmNoPQptZXRob2Q9bWFudWFsCmFkZHJlc3MxPTEwLjEwLjEwLjIwMS8yNCwxMC4xMC4xMC4xCmRucz0xLjEuMS4xCmlnbm9yZV9hdXRvX2Rucz10cnVlCgpbaXB2Nl0KZG5zLXNlYXJjaD0KYWRkci1nZW4tbW9kZT1ldWk2NAptZXRob2Q9aWdub3JlCg==",
						HumanRead: "\n[connection]\nid=em1\ntype=ethernet\ninterface-name=em1\n\n[ipv4]\ndns-search=\nmethod=manual\naddress1=10.10.10.201/24,10.10.10.1\ndns=1.1.1.1\nignore_auto_dns=true\n\n[ipv6]\ndns-search=\naddr-gen-mode=eui64\nmethod=ignore\n",
					},
				},
				{
					Path:      "/etc/NetworkManager/conf.d/noauto.conf",
					Mode:      420,
					Overwrite: true,
					Contents: Contents{
						Source:    "data:text/plain;charset=utf-8;base64,W21haW5dCiMgRG8gbm90IGRvIGF1dG9tYXRpYyAoREhDUC9TTEFBQykgY29uZmlndXJhdGlvbiBvbiBldGhlcm5ldCBkZXZpY2VzCiMgd2l0aCBubyBvdGhlciBtYXRjaGluZyBjb25uZWN0aW9ucy4Kbm8tYXV0by1kZWZhdWx0PSoK",
						HumanRead: "[main]\n# Do not do automatic (DHCP/SLAAC) configuration on ethernet devices\n# with no other matching connections.\nno-auto-default=*\n",
					},
				},
			},
		},
		Systemd: Systemd{
			Units: []Unit{
				{Name: "sshd.service", Enabled: true},
			},
		},
	}

	return root, nil
}
