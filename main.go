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
	Indices map[string]int
	Maclist []struct {
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
}

var MacData List

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")
	fmt.Fprintf(w, "<html><title>iPXE Server</title>")
	fmt.Fprintf(w, "<h1>")
	fmt.Fprintf(w, "Welcome to the HomePage!\n\n")
	fmt.Fprintf(w, "</h1>")

	var ipxe_template = "\t#!ipxe\n" +
		"\tkernel vmlinuz ip={ipaddr}::{gateway}:{netmask}::{interface}:off rd.cos.disable rd.noverifyssl root=live:http://192.168.178.7/harvester/rootfs.squashfs console=ttyS0 console=tty1 harvester.install.automatic=true harvester.install.config_url=http://192.168.178.7/harvester/config-{create|join}.yaml net.ifnames=1\n" +
		"\tinitrd initrd\n" +
		"\tboot\n"

	fmt.Fprintf(w, "<div>")
	fmt.Fprintf(w, "This server generates a iPXE script for a specific machine based on the request parameter.</br>")
	fmt.Fprintf(w, "For instance; http://192.168.178.7/harvester/ipxe/fe:c6:0b:44:98:03</br>")
	fmt.Fprintf(w, "</div>")
	fmt.Fprintf(w, "returns:\n\n")
	fmt.Fprintf(w, "<pre>")
	fmt.Fprintf(w, ipxe_template)
	fmt.Fprintf(w, "</pre>")

	fmt.Fprintf(w, "Source: <a href='https://linuxlink.timesys.com/docs/static_ip' target='_blank'>https://linuxlink.timesys.com/docs/static_ip</a>")
	fmt.Fprintf(w, "</html>")
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
		i := MacData.Indices[key]
		fmt.Println("Found match: ", MacData.Maclist[i].Macaddress, MacData.Maclist[i].Values.Ipaddress, MacData.Maclist[i].Values.Mode)
		fmt.Fprintf(w, "#!ipxe\n")
		fmt.Fprintf(w, "kernel http://192.168.178.7/harvester/"+MacData.Maclist[i].Values.Version+"/vmlinuz ip="+MacData.Maclist[i].Values.Ipaddress+"::"+MacData.Maclist[i].Values.Gateway+":"+MacData.Maclist[i].Values.Netmask+":harvester:"+MacData.Maclist[i].Values.Interface+":off rd.cos.disable rd.noverifyssl root=live:http://192.168.178.7/harvester/"+MacData.Maclist[i].Values.Version+"/rootfs.squashfs console=ttyS0 console=tty1 harvester.install.automatic=true harvester.install.config_url=http://192.168.178.7:10000/config/"+MacData.Maclist[i].Macaddress+" net.ifnames=1\n")
		fmt.Fprintf(w, "initrd http://192.168.178.7/harvester/"+MacData.Maclist[i].Values.Version+"/initrd\n")
		fmt.Fprintf(w, "boot\n")
	} else {
		fmt.Println("No match found for: ", key)
		fmt.Fprintf(w, "No match found for: "+key)
	}
}

func returnConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnConfig")
	vars := mux.Vars(r)
	key := vars["macaddr"]

	// lookup in indices, and access it.
	if match, _ := regexp.MatchString("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$", key); match == true {
		i := MacData.Indices[key]
		fmt.Println("Found match: ", MacData.Maclist[i].Macaddress, "Sending config...")
		json.NewEncoder(w).Encode(MacData.Maclist[i].Values.Config)
	} else {
		fmt.Println("No match found for: ", key)
		fmt.Fprintf(w, "No match found for: "+key)
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
