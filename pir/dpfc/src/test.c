
#include <openssl/rand.h>
#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <stdint.h>
#include <stdlib.h>
#include <time.h>
#include "../include/dpf.h"

#define EVALSIZE 1 << 20
#define EVALDOMAIN 20
#define FULLEVALDOMAIN 20
#define MAXRANDINDEX 1ULL << FULLEVALDOMAIN
#define FIELDSIZE 2

uint64_t randIndex() {
    srand(time(NULL));
    return ((uint64_t)rand()) % (MAXRANDINDEX);
}

void testDPF() {
    int size = EVALDOMAIN;
    uint64_t secretIndex = randIndex();
    uint8_t* key = malloc(16);
    RAND_bytes(key, 16);
    EVP_CIPHER_CTX* ctx = getDPFContext(key);
    unsigned char *k0 = malloc(INDEX_LASTCW + 16);
    unsigned char *k1 = malloc(INDEX_LASTCW + 16);
    genDPF(ctx, size, secretIndex, k0, k1);
    // printf("finished genDPF()\n");

    size_t L = EVALSIZE;
    uint64_t *X = malloc(sizeof(uint64_t) * L);
    for (size_t i = 0; i < L; i++) {
        int anotherIndex = randIndex();
        if (anotherIndex == secretIndex) {
            continue;
        }

        X[i] = anotherIndex;
    }

    X[0] = secretIndex;

    //************************************************
    // Test point-by-pont evaluation
    //************************************************

    uint128_t *shares0 = malloc(sizeof(uint128_t) * L);
    uint128_t *shares1 = malloc(sizeof(uint128_t) * L);
    
    clock_t t;
    t = clock();
    batchEvalDPF(ctx, size, false, k0, X, L, (uint8_t*)shares0);
    t = clock() - t;
    double time_taken = ((double)t) / (CLOCKS_PER_SEC / 1000.0); // ms 
    printf("Batch eval time (total) %f ms\n", time_taken);

    batchEvalDPF(ctx, size, true, k1, X, L, (uint8_t*)shares1);

   if (((shares0[0] + shares1[0]) % FIELDSIZE) != 1) {
        printf("FAIL (zero)\n");
        exit(0);
    }
    for (size_t i = 1; i < L; i++) {
        if (((shares0[i] + shares1[i]) % FIELDSIZE) != 0) {
            printf("FAIL (non-zero) at %zu\n", i);
            exit(0);
        }
    }
    free(shares0);
    free(shares1);
    free(k0);
    free(k1);
    free(X);
    printf("DONE\n\n");
    //************************************************
   
    //************************************************
    // Test full domain evaluation
    //************************************************    

    size = FULLEVALDOMAIN; // evaluation will result in 2^size points 
    int outl = 1 << size;
    secretIndex = randIndex();
    k0 = malloc(INDEX_LASTCW + 16);
    k1 = malloc(INDEX_LASTCW + 16);
    genDPF(ctx, size, secretIndex, k0, k1);

    // printf("Full domain = %i\n", outl);

    shares0 = malloc(sizeof(uint128_t) * outl);
    shares1 = malloc(sizeof(uint128_t) * outl);
    
    t = clock();
    fullDomainDPF(ctx, size, false, k0, (uint8_t*)shares0);
    t = clock() - t;
    time_taken = ((double)t) / (CLOCKS_PER_SEC / 1000.0); // ms 
    
    fullDomainDPF(ctx, size, true, k1, (uint8_t*)shares1);

    printf("Full-domain eval time (total) %f ms\n",time_taken);

    if (((shares0[secretIndex] + shares1[secretIndex]) % FIELDSIZE) != 1) {
        printf("FAIL (zero)\n");
        exit(0);
    }

    for (size_t i = 0; i < outl; i++) {
        if (i == secretIndex) 
            continue;

        if (((shares0[i] + shares1[i]) % FIELDSIZE) != 0) {
            printf("FAIL (non-zero)\n");
            exit(0);
        }
    }
   
    destroyContext(ctx);
    free(k0);
    free(k1);
    free(shares0);
    free(shares1);
    printf("DONE\n\n");
}


int main(int argc, char** argv) {

    int testTrials = 10;

    printf("******************************************\n");
    printf("Testing DPF\n");
    for (int i = 0; i < testTrials; i++) testDPF();
    printf("******************************************\n");
    printf("PASS\n");
    printf("******************************************\n\n");
}