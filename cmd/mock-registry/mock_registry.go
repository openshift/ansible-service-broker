package main

import (
	"encoding/json"
	"fmt"
	"github.com/fusor/ansible-service-broker/pkg/ansibleapp"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
)

type Args struct {
	AppFile string `short:"a" long:"appfile" description:"Mock Ansible Apps yaml file"`
}

func main() {
	args := &Args{}
	flags.Parse(args)

	if args.AppFile == "" {
		fmt.Println("Must provide --appfile $FILE")
		os.Exit(1)
	}

	fmt.Printf("Reading appfile: [ %s ]\n", args.AppFile)

	apps := LoadAnsibleApps(args.AppFile)

	http.HandleFunc("/ansibleapps", handler(args, apps, GetAnsibleApps))

	fmt.Println("Listening on localhost:1337")
	http.ListenAndServe(":1337", nil)
}

func GetAnsibleApps(w http.ResponseWriter, r *http.Request, args *Args, pApps *[]ansibleapp.Spec) {
	apps := *pApps
	fmt.Printf("Amount of ansibleapps %d\n", len(apps))

	for i, app := range apps {
		fmt.Printf("%d | ID: %s\n", i, app.Id)
		fmt.Printf("%d | Name: %s\n", i, app.Name)
		fmt.Printf("%d | Bindable: %t\n", i, app.Bindable)
		fmt.Printf("%d | Async: %s\n", i, app.Async)
		fmt.Println("===")
	}

	json.NewEncoder(w).Encode(pApps)
}

func LoadAnsibleApps(appFile string) *[]ansibleapp.Spec {
	// TODO: Is this required just to unwrap the root key and get the array?
	// Load just an array without a root key to wrap it?
	var parsedDat struct {
		AnsibleApps []ansibleapp.Spec
	}

	fmt.Println(appFile)
	dat, _ := ioutil.ReadFile(appFile)
	yaml.Unmarshal(dat, &parsedDat)

	return &parsedDat.AnsibleApps
}

type VanillaHandler func(http.ResponseWriter, *http.Request)
type InjectedHandler func(http.ResponseWriter, *http.Request, *Args, *[]ansibleapp.Spec)

func handler(args *Args, apps *[]ansibleapp.Spec, r InjectedHandler) VanillaHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		r(writer, request, args, apps)
	}
}
