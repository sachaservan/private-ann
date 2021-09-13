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
#define SIZE 64
#define FIELDSIZE 2147483647
#define FIELDBITS 31

#define FIELDMASK (((uint64_t) 1 << FIELDBITS) - 1)

typedef __int128 int128_t;
typedef unsigned __int128 uint128_t;

//PRG cipher context

extern EVP_CIPHER_CTX* GetContext(uint8_t*);
extern void destroyContext(EVP_CIPHER_CTX*);

//DPF functions

extern void genDPF(EVP_CIPHER_CTX *ctx, uint64_t index, unsigned char* k0, unsigned char *k1);
extern uint64_t evalDPF(EVP_CIPHER_CTX *ctx, bool b, unsigned char* k, uint64_t x);

#endif
