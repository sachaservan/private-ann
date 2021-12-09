
#include <openssl/rand.h>
#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <stdint.h>
#include <stdlib.h>
#include <time.h>
#include "../include/dpf.h"

#define EVALSIZE 1 << 20
#define EVALDOMAIN 64
#define FULLEVALDOMAIN 20
#define MAXRANDINDEX 1ULL << FULLEVALDOMAIN

uint64_t randIndex() {
    return ((uint64_t)rand()) % (MAXRANDINDEX);
}

void testDPF() {
    srand(time(NULL));
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