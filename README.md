scratch
=======

A bunch of random scratchings written in go.

# counter 

Counter is a simple stateful webserver. It was written to demonstrate one
possible use case for Checkpoint/Restore in Userspace [CRIU](https://criu.org)

## Set up 

The CRIU website has a very useful [article](https://criu.org/Docker) on
integration with docker. Doesn't require that much setup, but does rely on an
experimental feature that you'll have to manually enable on the docker daemon.
See the article for more details! 

## Instructions 

Run `make counter` to build the docker image and tag it as `counter:test`.

```
docker run --name counter -p 9993:9993 counter:test 
```

Once it's running, hit the endpoint a coupe of times to increment the counter.
When you've hit your favorite number, create a checkpoint:

```
docker checkpoint create counter checkpoint1
```

This command will step through the process of creating a checkpoint for the
container. More or less this seems to include: 
1. Seizing the processes with `ptrace()`
2. Collecting the details from `/proc/<PID>/*`
3. Injecting parasite code into the process, runs as a daemon in the address
   space of process. It essentially collects info about the types of pages that
   need to be written. It gets removed after usage.
4. All of the required information gets written to disk and the process gets
   killed.

To restart the frozen container, you can run:

```
docker start --checkpoint checkpoint1 counter
```

Note that this is the most trivial example for stoping and restarting the same
container, the CRIU article contains more information on using the checkpoint to
start different containers.

