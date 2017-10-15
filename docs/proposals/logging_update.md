## Logging Updates

### Problem Description
The current state is that we are setting loggers on objects and passing loggers around. This means that each object that we create has a logger and that if we make a function, to log from that function we need to pass in the logger. The solution is to allow for logging to take place at any location in the project. This will allow developers in the future to not have to worry about where the logger is included and instead just be able to log.

The broker will, instead of having  `*logger` being passed around, will instead use helper methods on a new package. This means that from now on when the broker would log something it will call the package like `log.Errorf(...)`.

The APBs will not be changed.

### Work Items
We need to decide how we are going to change this. I have put together 3 options that I think are all pretty good options that significantly help our current logging woes.

1. We can keep go-logging and use package level loggers. An example is the example go program [here](https://github.com/op/go-logging/blob/master/examples/example.go). Or an example that would look more like what we would do is select a file in each package and initialize the log.  example changes to `broker.go`
Example:
```go
package broker

import (
...
logging "github.com/op/go-logging"
...
)

var (
...
    log = lgging.MustGetLogger("broker")
)

...

func (a AnsibleBroker) getServiceInstance(instanceUUID uuid.UUID) (*apb.ServiceInstance, error) {
        instance, err := a.dao.GetServiceInstance(instanceUUID.String())
        if err != nil {
                if client.IsKeyNotFound(err) {
                    log.Errorf("Could not find a service instance in dao - %v", err)
                        return nil, ErrorNotFound
                }
                log.Error("Couldn't find a service instance: ", err)
                return nil, err
        }
        return instance, nil
}

...

```

2. We could create our own package and implement our own logging methods.
Example:
```go
package log

import (
    logger "...go-logging"
    ...
)

var log = logging.MustGetLogger("base-log")

var format = logging.....


func Errorf(formatString string, args ...interface{}) {
    log.Errorf(formatString, args)
}

func Error(args ...interface{}) {]
    log.Error(args)
}
```

3. We could abandon go-logging and move to something more familiar to [kubernetes](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go#L132) and use the [glog](https://github.com/golang/glog) package.


The next steps will be a combination of removing logging from all the structs, updating each log call to use the new log syntax.




