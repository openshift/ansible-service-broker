# Switch to using Fan-in & Fan-out / Observer pattern for subscribers

This is not quite the same as a fan-in / fan-out out pattern but not a
million miles away either.

## Problem Description

Our current subscriber pattern has some issues that have been outlined in the following
[github issue](https://github.com/openshift/ansible-service-broker/issues/638)
To surmise:

- **Shared Channel with single subscriber**: There is a single buffered channel per topic (topics being broker
actions such as provision, bind etc). All jobs, of the given type, write to
this channel. This has the potential to cause a bottleneck, particularly
as there is currently only a single subscriber reading from this channel and
this subscriber does not read the next message until it has completed
all of its work. Currently, there is not too much happening in these
subscribers, but if the work done in these subscribers were to increase
or become computationally complex, it could begin to slow down the throughput
of messages from the jobs and potentially block them. The likely hood
of this happening has been increased by the fact that we will soon allow
fine-grained status updates to sent by the APB developer via last operation
updates.

- **Separation of concerns**: While adding more subscribers would alleviate
the previous issue, the current pattern means only one subscriber can
deal with a given message at any one time. This makes it difficult to have a
separation of concerns.  Ideally, we would be able to have each subscriber
responsible for doing just one thing (recording metrics adding various
logs, persisting state etc all based on the msg it had been passed). This
would make it simple to add more functionality based on these messages
along with making testing simple and isolated.


## Goal

- Outline a design for passing messages, both user-generated and system
generated from executing APB jobs to each of a set of registered subscribers
that removes the outlined problems and improves the overall design of the broker.

## Subscriber design principal

This proposal outlines changes to the subscribers and for allowing
the addition of more subscribers. Each of these subscribers will act on the same message.
It is vital that the order these subscribers are called in should not matter, their work
should be independent of any other subscribers work.

## Proposal


### Change the subscriber interface

Currently, the subscribers implement the WorkSubscriber interface, we
would change this interface and the current subscribers.
Rather than passing the channel into the subscriber, we would
change this interface to accept a ```JobMsg```, value not a pointer to stop
unexpected mutation.
The current subscribers would handle state persistence and would be
renamed to reflect their role.
 
```go
#Note I call it observer here see terminology
type WorkSubscriber interface{
   Notify(msg JobMsg)
}
```

It is intended that each subscriber would handle any errors that occurred internally
either with logic or simply by logging the error in a clear fashion.


#### Change how subscribers are referenced

Currently, when a subscriber is added, we take the existing channel or
create one if not present and hand that channel to the subscriber.
Instead, we would now have a registry of subscribers to topics:

```go
type WorkEngine struct {
	subscribers map[WorkTopic][]WorkSubscriber
	...
} 

```

When a new subscriber is added, it would be added to the slice associated with a topic.

### Channel changes
There a couple of options available as the goal here is to avoid
any blocking:

#### 1) Keep the single channel but remove potential for blocking

Currently, the topic channel is handed to the single subscriber which
performs all the work. As mentioned this means no more messages can be
read from the channel until this unit of work is complete. A change to
how the messages are handled could alleviate any potential issue. The
reader would change to be the work engine which would start a read loop
for each of the topics when the broker started and distribute incoming
messages to any number of subscribers async:
As long as the principal that subscribers shouldn't depend on each other
is maintained there shouldn't be issues with this approach.
```go 
for{

   select{
       case msg := <-provisionChan:
           for _,s := range e.subscribers[ProvisionTopic]{
                go s(msg)              
           }
       case <-stop:
       break
   }
}


```

The stop channel here would be tied to os signals
[example](https://gist.github.com/reiki4040/be3705f307d3cd136e85).

In this case, we would still have a single channel per topic, but as we
hand off the message immediately and don't block waiting for the
subscribers to finish their work, there is little to no chance for the
channel to become full and so block the running jobs.

#### Add work engine start method

The work engine will now be the one reading messages from the channel
and passing it on to the Observers. To do this I propose we add a  public
``Start(stop <-chan struct{})`` method, that would be called in its own
go routine during ASB start up. I also propose adding a signal channel
that will send a message to anything running  in the background
(could also use context.Context). This would signal for the background
loops etc to stop.
This start method would start a loop for all of the topics as shown
above and close them down once a message was received.

#### 2) Create a channel per job

Remove the single channel per topic and replace it with a channel per
job. The work engine would be updated to create a channel per Job
when starting a new job. It would store active channels in a map where
the key would be the job token.  The work engine would be responsible for
the lifecycle of these channels. It would ensure they were started, stopped 
and resources cleaned up. 

As above, the subscribers would be registered in a map and called in a non-blocking way when a message was received.
While more complex, this method would allow us to leverage more control in the future.
For example adding a [WaitGroup](https://golang.org/pkg/sync/#WaitGroup) could allow us to limit how much 
parallel work we do per job. This would allow us to have some control over a "noisey" job that created many messages
while still allowing all of the subscribers to act on the message asynchronously 
(note this may not be needed and may be early optimisation right now and so should be seen as an example).

Sudo Code:

```go 

StartNewJob(token string, work Work, topic WorkTopic){
  a.jobs[token] = make(chan JobMsg)
  go func(){
      //optional used as an example to show limiting the work while still acting on the message async
      wait := &sync.WaitGroup{}
      //make sure we always close the channel  
      defer close(e.jobs[token])
      // read messages sent to the channel 
      go func(){
      // will auto stop once the channel is closed
        for msg := range a.jobs[token]{
            wait.Add(len(a.subscribers[topic]))
            for _, s := range a.subscribers[topic]{
               // Note we could choose to run these sync but again if there was a slow
               // subscriber it would block receiving on the channel.
               go s(msg,wait)
            }
            wait.Wait()
        } 
      }()
      // Start the actual job and block until done.
      work.Run()
  }()
}

```

Each job that was started would create approx 5 go routines and 1 channel.
I don't have too much concern here as go is designed to be able to spawn
many thousands of go routines without issue.



## Considerations

- perhaps want a config option to limit the number of in-progress jobs