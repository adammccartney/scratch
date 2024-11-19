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

Here are the default values on a fresh Rocky 9.4 install. The quickest way to
get the user limits on a linux system is with the ulimit command. The number of
processes available to a user is retrieved using the `-u` flag.

```
[adam@scratchy0 ~]$ ulimit -u
15380
```

A look at `/etc/security/limits.conf` shows the file where these limits have
traditionally been set and managed. As is seemingly the trend and tendency these
days, systemd also provides an interface to viewing and setting various limits
for the user. Searching for the default limit of number of processes allowed returns 
both soft and hard limits.

```
[adam@scratchy0 ~]$ systemctl show |grep DefaultLimitNPROC                                                                             

DefaultLimitNPROC=15380
DefaultLimitNPROCSoft=15380
```

I was kind of surprised to see that this default limit was set at roughly 4
times larger than the value that is shown when queried from within a container.

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


## What I want to figure out

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
in c, shell and haskell. Here is the original c version with some additional comments
from me:

```c
  #include <stdio.h>
  #include <unistd.h>

  void source() {
      int n;
      for(n = 2; ; n++) {
          // write an int to stdout (fd1) each time
          write(1, &n, sizeof(n));
      }
  }


  // this filter is created for each prime found,
  // it's job is to filter any multiples of that prime from a passing stream
  void cull(int p) {
      int n;
      for(;;) {
          // read an int from stdint (fd0) each time
          read(0, &n, sizeof(n));
          if (n % p != 0) { // p is not factor of n
              // write n to stdout (fd1)
              write(1, &n, sizeof(n));
          }
      }
  }

  /* connect stdint (k=0) or stdout (k=1) to pipe pd */
  void redirect(int k, int pd[2]) {
      // duplicate the file descriptor k
      dup2(pd[k], k);
      close(pd[0]);
      close(pd[1]);
  }

  void sink() {
      int pd[2];
      int p; /* a prime */     
      for (;;) {
          // read a prime from stdin (fd0)
          read(0, &p, sizeof(p));
          printf("%d\n", p);
          fflush(stdout);
          pipe(pd);
          if(fork()) {
              /* redirect stdin of this process to input of pipe pd */
              redirect(0, pd);
              continue;
          } else {
              /* redirect the stdout to the output of pipe pd */
              redirect(1, pd);
              cull(p);
          }
      }
  }

  int main() {      
      int pd[2];  /* pipe descriptors */
      pipe(pd);
      if (fork()) { /* parent process */
          redirect(0, pd);
          sink();
      } else {      /* child process */
          redirect(1, pd);
          source();
      }
  }
```

Here is a pipeline to run the coro-sieve program and keep track of the number of
child processes it spawns. Basically we are running the `coro-sieve` program
for *N* seconds, directing the output to nowhere, then monitoring the number of
child processes that get spawned in the context of the program. This can be 
the entrypoint of the container that we'll use for testing.

```bash
#!/bin/bash

timeout ${1}s coro-sieve > /dev/null &
while :
do 
    num_children=$(pstree -p $! | wc -l)
    sleep 1
    if [ $num_children == "0" ]; then
        exit 0
    else
        echo $num_children
    fi
done
```

The output of the pipeline *should* provide us with an overview of how many
child processes have been spawned. We can see that the number of processes is in
the thousands after just a couple of seconds. This is the unoptimized version of
the program from McIlroy's paper and running it confirms his suggestion that it
is a hog of resources.


> 3
> 773
> 1459
> 1883
> 2203
> 2524
> 2852
> 3240
> 3977
> 4558
> 5265
> 5960
> 6451
> 6819
> Terminated

For our purposes though, it's ideal - we now have a program that we can use to test
the behaviour of our system when it reaches the limit of allowable processes per
user.

## Testing 

In order to get somewhere close to the production environment, I'm testing on a
fresh rocky 9.4 instance that has docker installed. The container runtime is
different for our kubernetes nodes - rke2 doesn't use docker apparently.

> [root@adam ~]# cat /etc/rocky-release
> Rocky Linux release 9.4 (Blue Onyx)

The vm that we're working with has 8 processors and 8G of memory.

> [rocky@adam cgroups]$ nproc
> 8
> [rocky@adam cgroups]$ lsmem
> RANGE                                 SIZE  STATE REMOVABLE BLOCK
> 0x0000000000000000-0x000000003fffffff   1G online       yes   0-7
> 0x0000000100000000-0x00000002bfffffff   7G online       yes 32-87
> 
> Memory block size:       128M
> Total online memory:       8G
> Total offline memory:      0B


Interestingly the default limits are set quite a bit lower that those gleaned
from the VM at the beginning of this document. Both vms were spun up on
different proxmox instances, so this might be something that gets set at the
cloud provider level.

> [rocky@adam cgroups]$ ulimit -u
> 3296
> [root@adam ~]# systemctl show |grep DefaultLimitNPROC
> DefaultLimitNPROC=3296
> DefaultLimitNPROCSoft=3296

Okay, so basically we want to see what happens to the VM when we try to run
`coro-sieve` as a user, which spins up more than the allowable maximum.

### Directly on the system

Okay, so here is the output of the first run of the program directly on the vm.

The first run was probably started too early, the system had just booted and was
busy with a bunch of startup processes. Things get out of hand pretty quickly.
The program only ran for around 1 min and 15 seconds and managed to fork around
456 processes before we hit the limits for our rocky user. Note that the limits
here probably have to do with the fact that our cpus are all pretty busy
incrementing away.

> [rocky@adam cgroups]$ # ./entrypoint.sh 600
> [rocky@adam cgroups]$ echo $(date) coro-sieve start
> Mon Nov 18 03:32:46 PM UTC 2024 coro-sieve start
>  Mon Nov 18 03:34:00 PM UTC 2024 coro-sieve end with 456 children
> ./entrypoint.sh: fork: retry: Resource temporarily unavailable
> ./entrypoint.sh: fork: retry: Resource temporarily unavailable
> ...

The second time around it ran for a little bit longer (around 3 mins) and
managed to spawn 2149 children.

> [rocky@adam cgroups]$ echo $(date) coro-sieve start
> Mon Nov 18 03:38:45 PM UTC 2024 coro-sieve start
> [rocky@adam cgroups]$ # ./entrypoint.sh 600
>  Mon Nov 18 03:41:51 PM UTC 2024 coro-sieve end with 2149 children
> ./entrypoint.sh: fork: retry: Resource temporarily unavailable
> ./entrypoint.sh: fork: retry: Resource temporarily unavailable
> ...


### Containerized version



[^1]: <https://dl.acm.org/doi/10.1145/359576.359585>
[^2]: <https://www.cs.dartmouth.edu/~doug/sieve/sieve.pdf>
