Short demo of how a linux container is built
============================================

This shows what's going on inside a container:
1. create an archive with required binaries
2. import that archive using docker (this builds an image)
3. run the imported image

The above steps can be accomplised by running:

```
make docker_run
```

## Why the libraries? 

For the moment I'm just plonking the binaries of the shared libraries into version 
control. A more elegant and portable way to do this would to be to write a script
that generates the content of `mydate/`. Something like:

```
cp $(which bash) mydate/usr/bin/
```

Find the required libraries:

```
ldd $(which bash)
```

Copy those libraries 

```
cp /lib64/libc.so.6 mydate/lib64/libc.so.6
...
```
