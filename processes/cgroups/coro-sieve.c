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
