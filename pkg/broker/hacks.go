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
