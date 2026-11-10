sbatch harness
==============

test program to dispatch a number of sbatch workers that test out
various openblas installations provided by EESSI

## usage

(note that slurm partitions are hardcoded, first modify main.go if you want to
run on another site)

```
make runner
./runner
```
