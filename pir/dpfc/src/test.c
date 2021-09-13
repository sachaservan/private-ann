
#include <openssl/rand.h>
#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <stdint.h>
#include <stdlib.h>
#include "../include/dpf.h"

int main(int argc, char** argv) {
    uint64_t secretIndex = 5;
    if (argc > 1) {
        secretIndex = atoi(argv[1]);
    }
    uint8_t* key = malloc(16);
    RAND_bytes(key, 16);
    EVP_CIPHER_CTX* ctx = GetContext(key);
    unsigned char k0[18 * SIZE + 18 + 8];
    unsigned char k1[18 * SIZE + 18 + 8];
    genDPF(ctx, secretIndex, k0, k1);

    // evaluate at secretIndex
    uint64_t share0, share1;

    share0 = evalDPF(ctx, false, k0, secretIndex);
    share1 = evalDPF(ctx, true, k1, secretIndex);

    if (((share0 + share1) % FIELDSIZE) != 1) {
        printf("Fail expected 1, got %lu\n", share0 + share1);
    }

    
    for (int i = 0; i < 1000000; i++) {
        int anotherIndex = rand();
        if (anotherIndex == secretIndex) {
            continue;
        }
        share0 = evalDPF(ctx, false, k0, anotherIndex);
        share1 = evalDPF(ctx, true, k1, anotherIndex);
        if (((share0 + share1) % FIELDSIZE) != 0) {
            printf("Fail 0, got %lu\n", share0 + share1);
        }
    }
}