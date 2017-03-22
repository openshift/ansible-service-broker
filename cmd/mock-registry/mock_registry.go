package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/fusor/ansible-service-broker/pkg/apb"
	flags "github.com/jessevdk/go-flags"
	yaml "gopkg.in/yaml.v1"
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

	http.HandleFunc("/apbs", handler(args, apps, GetAnsibleApps))

	fmt.Println("Listening on localhost:1337")
	http.ListenAndServe(":1337", nil)
}

func GetAnsibleApps(w http.ResponseWriter, r *http.Request, args *Args, pApps *[]apb.Spec) {
	apps := *pApps
	fmt.Printf("Amount of apbs %d\n", len(apps))

	for i, app := range apps {
		fmt.Printf("%d | ID: %s\n", i, app.Id)
		fmt.Printf("%d | Name: %s\n", i, app.Name)
		fmt.Printf("%d | Description: %s\n", i, app.Description)
		fmt.Printf("%d | Bindable: %t\n", i, app.Bindable)
		fmt.Printf("%d | Async: %s\n", i, app.Async)
		fmt.Println("===")
	}

	json.NewEncoder(w).Encode(pApps)
}

func LoadAnsibleApps(appFile string) *[]apb.Spec {
	// TODO: Is this required just to unwrap the root key and get the array?
	// Load just an array without a root key to wrap it?
	var parsedDat struct {
		AnsibleApps []apb.Spec
	}

	fmt.Println(appFile)
	dat, _ := ioutil.ReadFile(appFile)
	yaml.Unmarshal(dat, &parsedDat)

	return &parsedDat.AnsibleApps
}

type VanillaHandler func(http.ResponseWriter, *http.Request)
type InjectedHandler func(http.ResponseWriter, *http.Request, *Args, *[]apb.Spec)

func handler(args *Args, apps *[]apb.Spec, r InjectedHandler) VanillaHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		r(writer, request, args, apps)
	}
}
