control groups and thread limits
================================

I came across a problem recently at work where a service dependency of an
application failed to start due to missing some targets on it's bootstrap tests.
Specifically, it required that the max number of threads per user be set to at
least 4096 (the system defaults were set to 3288). The user in this case is
named according to the application, the application gets started by the user in
a container. Note that this limit was read from initially read from the error
logs of the application and confirmed through running the `ulimit` command on a
root container on the node. It would be interesting to see if this limit is
specific to the container, or is the default on the host. If it is the latter,
then it appears to be set to quite a low value. 

## Some more context

For the time being, let's focus on recreating the context in which the original
error message presented itself. The application in question here is
elasticsearch, it's being run as a service dependency of temporal. We're
planning to run temporal as a workflow engine at a HPC center. After some
initial tests running a containerized version of temporal on a VM, we're moving
it to Kubernetes for the production deployment.

So, when temporal starts up, it tries to bring up the version of elasticsearch,
on which it relies for keeping track of the various workloads. As the
elasticsearch pods start, they run something like bootstrap check to see if they
have enough system resources available to do their job. It's at this point that
we got the error (they need a max of at least 4096 for NPROC per user). The
elasticsearch application is getting run by the elasticsearch user in a container
that is being run by the rke2 service.

The specific errors can be seen in the following output from the container logs:

```sh
> kubectl logs -n temporal elasticsearch-master-0
Defaulted container "elasticsearch" out of: elasticsearch, configure-sysctl (init)
...
ERROR: [1] bootstrap checks failed. You must address the points described in the following [1] lines before starting Elasticsearch.
bootstrap check failure [1] of [1]: max number of threads [3288] for user [elasticsearch] is too low, increase to at least [4096]
ERROR: Elasticsearch did not exit normally - check the logs at /usr/share/elasticsearch/logs/elasticsearch.log
```

Here are the limits as visible from the container: 

```sh
> kubectl debug node/dev-rke2-agent-1 -it --image=busybox
Creating debugging pod node-debugger-dev-rke2-agent-1-vkclx with container debugger on node dev-rke2-agent-1.
If you don't see a command prompt, try pressing enter.
/ # ulimit -u
3288
```

Which is the same as the view as root from the node:

```sh
[rocky@dev-rke2-agent-1 ~]$ ulimit -u
3288
```

One question that emerges at this point is what is setting this limit? So,
without going through all the specifics, it turned out not to be anything that
was getting set in the traditional pam/systemd/cloud-init config steps, but
rather a setting in the Proxmox template that was causing the VMs to be created
with hotplug enabled. There's an interesting discussion of the problem on the
proxmox mailing list[^4]. So, there was a long an meandering road of testing
various config settings that eventually landed at the hotplugging/numa issue.
And one of the things that got tested along the way was how & if cgroups were
affecting the setting. That's pretty much what I want to write about in this 
post.



A look at `/etc/security/limits.conf` shows the file where these limits have
traditionally been set and managed. As is seemingly the trend and tendency these
days, systemd also provides an interface to viewing and setting various limits
for the user. Searching for the default limit of number of processes allowed returns 
both soft and hard limits.

I was kind of surprised to see that this default limit was set at roughly 4
times larger than the value that is shown when queried from within a container.


### So what is setting the limits?

#### Is it the pam_limits config files


#### Is it rlimit?

#### Is it systemd-system.conf?

It doesn't appear so. Accoring to the systmed-system.conf manpage

<pre>
NAME
       systemd-system.conf, system.conf.d, systemd-user.conf, user.conf.d - System and session
       service manager configuration files

SYNOPSIS
       /etc/systemd/system.conf, /etc/systemd/system.conf.d/*.conf,
       /run/systemd/system.conf.d/*.conf, /usr/lib/systemd/system.conf.d/*.conf

       ~/.config/systemd/user.conf, /etc/systemd/user.conf, /etc/systemd/user.conf.d/*.conf,
       /run/systemd/user.conf.d/*.conf, /usr/lib/systemd/user.conf.d/*.conf
</pre>


## What I want to figure out about cgroups

How do cgroups really work? i.e. what view of the system resources is the
elasticsearch user seeing?

How can we adapt the limits so that the elasticsearch user has enough?

What's the difference between setting limits using `systemd.exec` and cgroups
using `systemd.resource-control`?


## plan of action

1. Create a simple program that spawns some arbitrary number of threads. Run it
   on the system outside of the container to get a benchmark of how it runs. See
   what happens when we reach the max number of threads.
2. Containerize this application and create a user in the container that will run it
3. Give control of running the container to the container engine on the test node
4. Try to understand how the control groups are affecting the limits


## Test program

In his paper on Communicating Sequential Processes[^1] Tony Hoare mentions a
really fun paper by Douglas McIlroy that sketches out a concurrent version
of the Sieve of Eratosthenes[^2]. The paper contains versions of the program
in c, shell and haskell. 

For the various tests described below, I wrote a short bash script that
basically just runs that program and polled the number of forked processes.
Here is a pipeline to run the coro-sieve program and keep track of the number of
child processes it spawns. Basically we are running the `coro-sieve` program
for *N* seconds, directing the output to nowhere, then monitoring the number of
child processes that get spawned in the context of the program. This can be
the entrypoint of the container that we'll use for testing.

```sh
#!/bin/bash

timeout ${1}s coro-sieve > /dev/null &
while :
do 
    unset nproc
    nproc=$(pstree -p $! | wc -l)
    sleep 1
    if [ $nproc == "0" ]; then
        exit 0
    else
        echo -e $(date) -- INFO -- coro-sieve forked $nproc processes
    fi
done
```

The output of the pipeline *should* provide us with an overview of how many
child processes have been spawned. We can see that the number of processes is in
the thousands after just a couple of seconds. This is the unoptimized version of
the program from McIlroy's paper and running it confirms his suggestion that it
is a hog of resources.

For our purposes though, it's ideal - we now have a program that we can use to test
the behaviour of our system when it reaches the limit of allowable processes per
user.

## Testing 


### Directly on the system

Okay, so here is the output of the first run of the program directly on the vm.

The first run was probably started too early, the system had just booted and was
busy with a bunch of startup processes. Things get out of hand pretty quickly.
The program only ran for around 1 min and 15 seconds and managed to fork around
456 processes before we hit the limits for our rocky user. Note that the limits
here probably have to do with the fact that our cpus are all pretty busy
incrementing away.

```sh
[rocky@adam cgroups]$ # ./entrypoint.sh 600
...
Mon Nov 18 03:34:00 PM UTC 2024 -- INFO -- coro-sieve forked 456 processes
./entrypoint.sh: fork: retry: Resource temporarily unavailable
./entrypoint.sh: fork: retry: Resource temporarily unavailable
...
```

The second time around it ran for a little bit longer (around 3 mins) and
managed to spawn 2149 children.

```sh
[rocky@adam cgroups]$ echo $(date) coro-sieve start
Mon Nov 18 03:38:45 PM UTC 2024 coro-sieve start
[rocky@adam cgroups]$ # ./entrypoint.sh 600
 Mon Nov 18 03:41:51 PM UTC 2024 coro-sieve end with 2149 children
./entrypoint.sh: fork: retry: Resource temporarily unavailable
./entrypoint.sh: fork: retry: Resource temporarily unavailable
...
```

#### System limits

System wide limits are set up by passing config values to the `pam_limits`
module. The limits of number of processes allowed for the user rocky can be
set as follows:


```sh
[rocky@adam cgroups]$ cat > /etc/security/limits.d/rocky.conf <<EOF
@rocky hard nproc 2048
@rocky soft nproc 2048
EOF
```

Another run of our test program, shows that it runs into trouble more 
or less where we expect.

```sh
 Tue Nov 19 08:51:59 AM UTC 2024 coro-sieve end with 2038 children
./entrypoint.sh: fork: retry: Resource temporarily unavailable
./entrypoint.sh: fork: retry: Resource temporarily unavailable
...
```

A side note on systmed here - reading the systemd docs would lead you to believe
that you should always go through systemd config files in order to set up
limits. In particular, if you are approaching the problem from the perspective
as a service user who is hitting resource limits, you end up getting directed
towards the systemd documentation in order to adjust the cgroup configuration
stuff. There are more details on configuring cgroups for a specific service
below, but here briefly is how you would set up cgroup limits for a user using
systemd. Specifically, `systemd.resource-control` will look for config files in
the form of "Slices".

```sh
cat > /etc/systemd/system/user-1000.slice.d/limits.conf <<EOF
[Slice]
TasksMax=1024
```

```sh
[rocky@adam ~]$ systemctl show user-1000.slice |grep TasksMax
TasksMax=1024
[rocky@adam ~]$ ~/cgroups/entrypoint.sh 300
 Tue Nov 19 12:30:30 PM UTC 2024 coro-sieve end with 945 children
/home/rocky/cgroups/entrypoint.sh: fork: retry: Resource temporarily unavailable
/home/rocky/cgroups/entrypoint.sh: fork: retry: Resource temporarily unavailable
^C/home/rocky/cgroups/entrypoint.sh: fork: Interrupted system call
```

The important question though is how will the program behave if the limits set
by configuring `pam_limits` conflict with those configured through
`systemd.resource-control`?

```sh
[rocky@adam ~]$ ulimit -u
512
[rocky@adam ~]$ cat /etc/security/limits.d/user-1000.conf
@rocky hard nproc 512
@rocky soft nproc 512
[rocky@adam ~]$ cat /etc/systemd/system/user-1000.slice.d/limits.conf
[Slice]
TasksMax=1024
```

We run into trouble when the program hits the limts set by
`/etc/security/limits.d`, i.e. the config values supplied to `pam_limits`.

```sh
 Tue Nov 19 02:48:57 PM UTC 2024 coro-sieve end with 489 children
./entrypoint.sh: fork: retry: Resource temporarily unavailable
./entrypoint.sh: fork: retry: Resource temporarily unavailable
```

What's the moral of the story? The limits set by configuring `pam_limits`
will override those set using `systemd.resource-control`!

### Containerized version

Okay, so here we want to figure out how to use cgroups to control how our
container is running. We'll do this by setting limits for the service running
the container. In the case of this example, that is the docker service which 
runs the container runtime (containerd).

We want to run the docker-compose script as a systemd service. This has become
fairly common practice. RedHat seem to have even gone a bit further and adding a
feature to export running containers to a systemd service file that can be
installed in one of the expected locations. But many other creative people have
found ways to do similiar things with docker-compose and systemd[^3].

So where do the ulimits get set for the service? Initially, I thought that 
adding a Tas

```sh
[root@adam ~]# systemctl start docker-compose@coro-sieve
[root@adam ~]# docker logs coro-sieve-coro-sieve-1
Wed Nov 20 11:38:58 UTC 2024 -- INFO -- coro-sieve forked 3 children
/app/entrypoint.sh: fork: retry: Resource temporarily unavailable
/app/entrypoint.sh: fork: retry: Resource temporarily unavailable
...
```


```yaml
services:
  coro-sieve:
    build: .
    image: coro-sieve:test
    deploy:
      resources:
        limits:
          pids: 256
    stdin_open: true
    tty: true
    command: ${TIMEOUT:-120}
```


```sh
[root@adam coro-sieve]# grep LimitNPROC /usr/lib/systemd/system/docker.service
LimitNPROC=infinity
```

```sh
[root@adam coro-sieve]# sed -i 's/^\(LimitNPROC=\)\(infinity\)/\1128/g' /usr/lib/systemd/system/docker.service
[root@adam coro-sieve]# grep LimitNPROC /usr/lib/systemd/system/docker.service
LimitNPROC=128
```

```sh
[root@adam coro-sieve]# sed -i 's/\(pids: \)\(256\)/\11024/g' /etc/docker/compose/coro-sieve/docker-compose.yml
[root@adam coro-sieve]# grep pids: /etc/docker/compose/coro-sieve/docker-compose.yml
          pids: 1024
```

I had assumed that the limits set in the `docker.service` unit file would be the
limiting factor, but it looks like the cgroups setting in the docker-compose
file are the ones that take precedence. This is interesting, because it means
that _both_ the `ulimit -u` limits and the `docker.service` limits can be
bypassed.

```sh
[root@adam ~]# docker logs coro-sieve-coro-sieve-1
Wed Nov 20 11:48:03 UTC 2024 -- INFO -- coro-sieve forked 1 children
Wed Nov 20 11:48:04 UTC 2024 -- INFO -- coro-sieve forked 185 children
...
Wed Nov 20 11:49:05 UTC 2024 -- INFO -- coro-sieve forked 1013 children
Wed Nov 20 11:49:07 UTC 2024 -- INFO -- coro-sieve forked 1018 children
/app/entrypoint.sh: fork: retry: Resource temporarily unavailable
/app/entrypoint.sh: fork: retry: Resource temporarily unavailable
...
```

[^1]: <https://dl.acm.org/doi/10.1145/359576.359585>
[^2]: <https://www.cs.dartmouth.edu/~doug/sieve/sieve.pdf>
[^3]: <https://gist.github.com/mosquito/b23e1c1e5723a7fd9e6568e5cf91180f>
[^4]: <https://lists.proxmox.com/pipermail/pve-user/2023-August/017214.html>
