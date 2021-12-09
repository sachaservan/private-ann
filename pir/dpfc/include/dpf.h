// From SabaEskandarian/OlivKeyValCode
// from https://github.com/ucbrise/dory/blob/master/src/c/dpf.h
#ifndef _DPF
#define _DPF

#include <stdio.h>
#include <string.h>
#include <stdint.h>

#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <string.h>

#include <stdbool.h>

#define INDEX_LASTCW 18*size + 18
#define CWSIZE 18

#define FIELDSIZE 2147483647
#define FIELDBITS 31
#define FIELDMASK (((uint128_t) 1 << FIELDBITS) - 1)

#define LEFT 0
#define RIGHT 1

typedef struct Hash hash;
typedef __int128 int128_t;
typedef unsigned __int128 uint128_t;

// PRG cipher context
extern EVP_CIPHER_CTX* getDPFContext(uint8_t*);
extern void destroyContext(EVP_CIPHER_CTX*);

// DPF functions
extern void genDPF(EVP_CIPHER_CTX *ctx, int size, uint64_t index, unsigned char* k0, unsigned char *k1);
extern void batchEvalDPF(EVP_CIPHER_CTX *ctx, int size, bool b, unsigned char* k, uint64_t *in, uint64_t inl, uint8_t* out);
extern void fullDomainDPF(EVP_CIPHER_CTX *ctx, int size, bool b, unsigned char* k, uint128_t *outSeeds, int *outBits);

#endif
