/*
 * This program is for exploring a few ideas found in the srouces of dennis
 * ritchie's v6, specifically in the source code for find the find
 * implementation was attributed to Dick Haight (what a name)
 * https://doc.cat-v.org/unix/find-history sources:
 * https://www.tuhs.org/Archive/Distributions/Research/Dennis_v6/
 * */
#include <signal.h>
#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

/* use execvp semantics */
int
run_cmd(char** argv)
{
    pid_t pid;
    int status;
    if (signal(SIGCHLD, SIG_IGN) == SIG_ERR) {
        perror("signal");
        exit(EXIT_FAILURE);
    }
    pid = fork();
    switch (pid) {
    case -1:
        perror("fork");
        exit(EXIT_FAILURE);
    case 0:
        execvp(argv[0], argv);
        fflush(stdout);
        _exit(EXIT_SUCCESS);
    default:
        wait(&status);
        exit(EXIT_SUCCESS);
    }
}


/* hold state for parsing arguments */
struct arg_cursor {
    int pos;
    int count;
    char** vec;
};

/* get next arg from commandline */
char* nxtarg(struct arg_cursor* cur)
{
    if(cur->pos >= cur->count) return NULL;
    return(cur->vec[cur->pos++]);
}


/*
 * handling stuff _after_ exec is the tricky part
 * */
int 
main(int argc, char* argv[])
{
    struct arg_cursor cur = {
        .pos = 1,
        .count = argc,
        .vec = argv
    };

    int exec_start_idx = 0, exec_end_idx = 0;
    if (argc < 2) {
        printf("Insufficient args\n");
        exit(9);
    }
    printf("%s %s\n", nxtarg(&cur), nxtarg(&cur));

    char* a;
    char* args[cur.count];

    if (strcmp(nxtarg(&cur), "-exec")==0) {
        exec_start_idx = cur.pos;
        exec_end_idx = exec_start_idx;
        int i = 0;
        args[i] = malloc(strlen(cur.vec[cur.pos]) + 1);
        strcpy(args[i], cur.vec[cur.pos]);
        i++;
        while ((a = nxtarg(&cur)) != NULL && (strcmp(a, ";") != 0)) {
            args[i] = malloc(strlen(cur.vec[cur.pos]) + 1);
            strcpy(args[i], cur.vec[cur.pos]);
            exec_end_idx++;
            i++;
        }
        args[i - 1] = NULL;
        run_cmd(args);
    }
    int j = 0;
    while (args[j] != NULL) {
        free(args[j++]);
    }
}
