package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
	yml "sigs.k8s.io/yaml"
)

type List struct {
	ConfigServer   string `json:"server"`
	DownloadServer string `json:"downloadserver"`
	Maclist        []struct {
		Macaddress string `json:"macaddress"`
		Values     struct {
			Ipaddress string      `json:"ipaddress,omitempty"`
			Netmask   string      `json:"netmask,omitempty"`
			Gateway   string      `json:"gateway,omitempty"`
			Interface string      `json:"interface,omitempty"`
			Version   string      `json:"version,omitempty"`
			Mode      string      `json:"type,omitempty"`
			Config    interface{} `json:"config,omitempty"`
		} `json:"values"`
	} `json:"maclist"`
	Config []struct {
		Cluster      string      `json:"cluster,omitempty"`
		ConfigCreate interface{} `json:"config_create,omitempty"`
		ConfigJoin   interface{} `json:"config_join,omitempty"`
	} `json:"config,omitempty"`
	Indices map[string]int
}

var MacData List

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")
	fmt.Fprintf(w, "Welcome to this Harvester iPXE bootstrap server!\n\n")
	fmt.Fprintf(w, "This server provides 3 endpoints:\n\n")
	fmt.Fprintf(w, "\t- /all\t sends entire configuration\n")
	fmt.Fprintf(w, "\t- /ipxe/<mac address>\t sends an ipxe script specific for that mac address\n")
	fmt.Fprintf(w, "\t- /config/<mac address>\t sends a harvester bootstrap configuration for that specific server\n\n")
	// examples
	fmt.Fprintf(w, "Examples:\n\n")
	fmt.Fprintf(w, "\t- http://example.org/all\n")
	fmt.Fprintf(w, "\t- http://example.org/ipxe/AB:cd:Ef:01:F5:c9\n")
	fmt.Fprintf(w, "\t- http://example.org/config/CD:Ef:01:F5:c9:01\n")
}

func returnAllMacIPs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllMacIPs")
	json.NewEncoder(w).Encode(MacData)
}

func returnIPXEScript(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnIPXEScript")
	vars := mux.Vars(r)
	key := vars["macaddr"]

	// lookup in indices, and access it.
	if match, _ := regexp.MatchString("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$", key); match == true {
		if m, ok := MacData.Indices[key]; ok {
			fmt.Println("Got mac-address from index; ", m)
			fmt.Println("Found match: ", MacData.Maclist[m].Macaddress, MacData.Maclist[m].Values.Ipaddress, MacData.Maclist[m].Values.Mode)
			fmt.Fprintf(w, "#!ipxe\n")
			fmt.Fprintf(w, "kernel "+MacData.DownloadServer+"/harvester/"+MacData.Maclist[m].Values.Version+"/harvester-"+MacData.Maclist[m].Values.Version+"-vmlinuz-amd64 ip="+MacData.Maclist[m].Values.Ipaddress+"::"+MacData.Maclist[m].Values.Gateway+":"+MacData.Maclist[m].Values.Netmask+":harvester:"+MacData.Maclist[m].Values.Interface+":off rd.cos.disable rd.noverifyssl root=live:"+MacData.DownloadServer+"/harvester/"+MacData.Maclist[m].Values.Version+"/harvester-"+MacData.Maclist[m].Values.Version+"-rootfs-amd64.squashfs console=ttyS0 console=tty1 harvester.install.automatic=true harvester.install.config_url="+MacData.ConfigServer+":10000/config/"+MacData.Maclist[m].Macaddress+" net.ifnames=1\n")
			fmt.Fprintf(w, "initrd "+MacData.DownloadServer+"/harvester/"+MacData.Maclist[m].Values.Version+"/harvester-"+MacData.Maclist[m].Values.Version+"-initrd-amd64\n")
			fmt.Fprintf(w, "boot\n")
		} else {
			fmt.Println("No match found for: ", key)
			fmt.Fprintf(w, "No match found for: "+key)
		}
	}
}

func returnConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnConfig")
	vars := mux.Vars(r)
	key := vars["macaddr"]

	// lookup in indices, and access it.
	if match, _ := regexp.MatchString("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$", key); match == true {
		if m, ok := MacData.Indices[key]; ok {
			fmt.Println("Got mac-address from index; ", m)
			fmt.Println("Found match: ", MacData.Maclist[m].Macaddress, "Sending config...")
			json.NewEncoder(w).Encode(MacData.Maclist[m].Values.Config)
		} else {
			fmt.Println("No match found for: ", key)
			fmt.Fprintf(w, "No match found for: "+key)
		}
	}
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/all", returnAllMacIPs)
	myRouter.HandleFunc("/config/{macaddr}", returnConfig)
	myRouter.HandleFunc("/ipxe/{macaddr}", returnIPXEScript)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	fmt.Println("Rest API v2.0 - Mux Routers")

	// Get location of the configuration data, else exit
	configfile := os.Getenv("CONFIG_FILE")
	if configfile == "" {
		log.Fatalf("Please, set the CONFIG_FILE environment variable to a valid value")
		return
	}

	// read data from file
	buf, err := ioutil.ReadFile(filepath.Join("/config/", configfile))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	j, err := yml.YAMLToJSON(buf)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	// var MacData List
	json.Unmarshal([]byte(j), &MacData)
	if err != nil {
		fmt.Printf("Error parsing JSON string - %s", err)
	}

	// Initiate Indices (empty)
	MacData.Indices = make(map[string]int)
	// fill indice
	for i, v := range MacData.Maclist {
		m := v.Macaddress
		MacData.Indices[m] = i
	}

	handleRequests()
}
