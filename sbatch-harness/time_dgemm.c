#include "stdio.h"
#include "stdlib.h"
#include "sys/time.h"
#include "time.h"
#include <stdbool.h>
#include <unistd.h>

#include <cblas.h>

int openblas_get_num_threads(void);
int openblas_get_num_procs(void);
char* openblas_get_config(void);

extern void dgemm_(char*, char*, int*, int*,int*, double*, double*, int*, double*, int*, double*, double*, int*);

int main(int argc, char** argv)
{

  int ob_nthrds = openblas_get_num_threads();
  int ob_nprcs = openblas_get_num_procs();
  char* ob_cnfg =  openblas_get_config();

  printf("openblas nmbr of threads:%d\n", ob_nthrds);
  printf("openblas nmbr of procs:%d\n", ob_nprcs);
  printf("openblas config:%s\n", ob_cnfg);


  const char *filename = NULL;
  int option;

  while ((option = getopt(argc, argv, "f:")) != -1) {
      switch (option) {
          case 'f':
              filename = optarg;
              break;
          default:
              fprintf(stderr,
                      "Usage: %s -f FILE M N K\n",
                      argv[0]);
              return EXIT_FAILURE;
      }
  }
  printf("argc=%d, optind=%d, remaining=%d, filename=%s\n",
       argc, optind, argc - optind,
       filename ? filename : "(null)");

  if (filename == NULL || argc - optind != 3) {
      fprintf(stderr,
               "Usage: %s -f FILE M N K\n",
               argv[0]);
      return EXIT_FAILURE;
  }

  char **args = argv + optind;

  int m = atoi(args[0]);
  int n = atoi(args[1]);
  int k = atoi(args[2]);
  int sizeofa = m * k;
  int sizeofb = k * n;
  int sizeofc = m * n;
  char ta = 'N';
  char tb = 'N';
  double alpha = 1.2;
  double beta = 0.001;

  struct timeval start,finish;
  double duration;

  double* A = (double*)malloc(sizeof(double) * sizeofa);
  double* B = (double*)malloc(sizeof(double) * sizeofb);
  double* C = (double*)malloc(sizeof(double) * sizeofc);

  srand((unsigned)time(NULL));

  int i;
  for (i=0; i<sizeofa; i++)
    A[i] = i%3+1;//(rand()%100)/10.0;

  for (i=0; i<sizeofb; i++)
    B[i] = i%3+1;//(rand()%100)/10.0;

  for (i=0; i<sizeofc; i++)
    C[i] = i%3+1;//(rand()%100)/10.0;
  //#if 0
  printf("m=%d,n=%d,k=%d,alpha=%lf,beta=%lf,sizeofc=%d\n",m,n,k,alpha,beta,sizeofc);
  gettimeofday(&start, NULL);
  dgemm_(&ta, &tb, &m, &n, &k, &alpha, A, &m, B, &k, &beta, C, &m);
  gettimeofday(&finish, NULL);

  duration = ((double)(finish.tv_sec-start.tv_sec)*1000000 + (double)(finish.tv_usec-start.tv_usec)) / 1000000;
  double gflops = 2.0 * m *n*k;
  gflops = gflops/duration*1.0e-6;

  FILE *fp;
  fp = fopen(filename, "a");
  fprintf(fp, "%dx%dx%d\t%lf s\t%lf MFLOPS\n", m, n, k, duration, gflops);
  fclose(fp);

  free(A);
  free(B);
  free(C);
  return 0;
}
