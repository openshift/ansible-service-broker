package broker

import (
	"github.com/pborman/uuid"
)

var plans = []Plan{
	{
		ID:          uuid.Parse("4c10ff42-be89-420a-9bab-27a9bef9aed8"),
		Name:        "default",
		Description: "Default plan",
		Free:        true,
	},
}

// copied from github.com/openshift/origin/pkg/cmd/util/clientcmd/shortcut_restmapper.go
//var userResourcesCopy = []string{
//"buildconfigs", "builds",
//"imagestreams",
//"deploymentconfigs", "replicationcontrollers",
//"routes", "services",
//"pods",
//}

//var userResources = append(userResourcesCopy, "configmaps")
