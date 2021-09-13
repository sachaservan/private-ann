// From SabaEskandarian/OblivKeyValCode/
// From https://github.com/ucbrise/dory/blob/master/src/c/dpf.c

// This is the 2-party FSS evaluation function for point functions.
// This is based on the following paper:
// Boyle, Elette, Niv Gilboa, and Yuval Ishai.
// "Function Secret Sharing: Improvements and Extensions." 
// Proceedings of the 2016 ACM SIGSAC Conference on Computer and Communications Security. 
// ACM, 2016.

#include "../include/dpf.h"
#include <openssl/rand.h>
#include <stdint.h>
#include <time.h>
#include <math.h>

EVP_CIPHER_CTX* GetContext(uint8_t* key) {
    EVP_CIPHER_CTX* randCtx;
    if(!(randCtx = EVP_CIPHER_CTX_new()))
        printf("errors occured in creating context\n");
    if(1 != EVP_EncryptInit_ex(randCtx, EVP_aes_128_ecb(), NULL, key, NULL))
        printf("errors occured in randomness init\n");
    EVP_CIPHER_CTX_set_padding(randCtx, 0);
    return randCtx;
}

void destroyContext(EVP_CIPHER_CTX * ctx) {
 	EVP_CIPHER_CTX_free(ctx);
}

/*
void printBytes(void* p, int num) {
	unsigned char* c = (unsigned char*) p;
	for (int i = 0; i < num; i++) {
		printf("%02x", c[i]);
	}
	printf("-\n");
}
*/

static inline uint128_t dpf_reverse_lsb(uint128_t input){
    uint128_t xor = 1;
	return input ^ xor;
}

static inline uint128_t dpf_lsb(uint128_t input){
    return input & 1;
}

static inline uint128_t dpf_set_lsb_zero(uint128_t input){
    int lsb = input & 1;

	if(lsb == 1){
		return dpf_reverse_lsb(input);
	}else{
		return input;
	}
}

uint128_t getRandomBlock(){
    static uint8_t* randKey = NULL;//(uint8_t*) malloc(16);
    static EVP_CIPHER_CTX* randCtx;
    static uint128_t counter = 0;

    int len = 0;
    uint128_t output = 0;
    if(!randKey){
        randKey = (uint8_t*) malloc(16);
        if(!(randCtx = EVP_CIPHER_CTX_new()))
            printf("errors occured in creating context\n");
        if(!RAND_bytes(randKey, 16)){
            printf("failed to seed randomness\n");
        }
        if(1 != EVP_EncryptInit_ex(randCtx, EVP_aes_128_ecb(), NULL, randKey, NULL))
            printf("errors occured in randomness init\n");
        EVP_CIPHER_CTX_set_padding(randCtx, 0);
    }

    if(1 != EVP_EncryptUpdate(randCtx, (uint8_t*)&output, &len, (uint8_t*)&counter, 16))
        printf("errors occured in generating randomness\n");
    counter++;
    return output;
}

//this is the PRG used for the DPF
void dpfPRG(EVP_CIPHER_CTX *ctx, uint128_t input, uint128_t* output1, uint128_t* output2, int* bit1, int* bit2){

	input = dpf_set_lsb_zero(input);

    int len = 0;
	uint128_t stashin[2];
	stashin[0] = input;
	stashin[1] = dpf_reverse_lsb(input);
	uint128_t stash[2];

    EVP_CIPHER_CTX_set_padding(ctx, 0);
    if(1 != EVP_EncryptUpdate(ctx, (uint8_t*)stash, &len, (uint8_t*)stashin, 32))
        printf("errors occured in encrypt\n");
    //no need to do this since we're working with exact multiples of the block size
    //if(1 != EVP_EncryptFinal_ex(ctx, stash + len, &len))
    //    printf("errors occured in final\n");

	stash[0] = stash[0] ^ input;
	stash[1] = stash[1] ^ input;
	stash[1] = dpf_reverse_lsb(stash[1]);

	*bit1 = dpf_lsb(stash[0]);
	*bit2 = dpf_lsb(stash[1]);

	*output1 = dpf_set_lsb_zero(stash[0]);
	*output2 = dpf_set_lsb_zero(stash[1]);
}

static inline int getbit(uint64_t x, int b){
	return ((x) >> (SIZE - b)) & 1;
}

static inline uint64_t convert(uint128_t raw) {
	uint64_t r = (uint64_t)(raw) & FIELDMASK;
	return r < FIELDSIZE ? r : r - FIELDSIZE;
}

static inline uint64_t negate(uint64_t x) {
	return x != 0 ? FIELDSIZE - x : 0;
}

static inline uint64_t modAfterAdd(uint64_t r) {
	return r < FIELDSIZE ? r : r - FIELDSIZE;
}


void genDPF(EVP_CIPHER_CTX *ctx, uint64_t index, unsigned char* k0, unsigned char *k1){
    uint128_t s[SIZE + 1][2];
	int t[SIZE + 1 ][2];
	uint128_t sCW[SIZE];
	int tCW[SIZE][2];

    s[0][0] = getRandomBlock();
	s[0][1] = getRandomBlock();
	t[0][0] = 0;
	t[0][1] = 1;

    uint128_t s0[2], s1[2]; // 0=L,1=R
    int t0[2], t1[2];
	#define LEFT 0
	#define RIGHT 1
	for(int i = 1; i <= SIZE; i++){
        dpfPRG(ctx, s[i-1][0], &s0[LEFT], &s0[RIGHT], &t0[LEFT], &t0[RIGHT]);
		dpfPRG(ctx, s[i-1][1], &s1[LEFT], &s1[RIGHT], &t1[LEFT], &t1[RIGHT]);

        int keep, lose;
		int indexBit = getbit(index, i);
        if(indexBit == 0){
			keep = LEFT;
			lose = RIGHT;
		}else{
			keep = RIGHT;
			lose = LEFT;
		}


        sCW[i-1] = s0[lose] ^ s1[lose];

		tCW[i-1][LEFT] = t0[LEFT] ^ t1[LEFT] ^ indexBit ^ 1;
		tCW[i-1][RIGHT] = t0[RIGHT] ^ t1[RIGHT] ^ indexBit;

		if(t[i-1][0] == 1){
			s[i][0] = s0[keep] ^ sCW[i-1];
			t[i][0] = t0[keep] ^ tCW[i-1][keep];
		}else{
			s[i][0] = s0[keep];
			t[i][0] = t0[keep];
		}

		if(t[i-1][1] == 1){
			s[i][1] = s1[keep] ^ sCW[i-1];
			t[i][1] = t1[keep] ^ tCW[i-1][keep];
		}else{
			s[i][1] = s1[keep];
			t[i][1] = t1[keep];
		}

    }

	// printBytes(&s[SIZE][0], 16);
	// printBytes(&s[SIZE][1], 16);
    uint64_t sFinal0 = convert(s[SIZE][0]);
    uint64_t sFinal1 = convert(s[SIZE][1]);
    uint64_t lastCW = modAfterAdd(1 + negate(sFinal0) + sFinal1);
	
    if (t[SIZE][1] % 2 != 0) {
        lastCW = negate(lastCW);
    }

	// printf("0: %lu 1: %lu lastCw: %lu\n", sFinal0, sFinal1, lastCW);

	k0[0] = SIZE;
	memcpy(&k0[1], &s[0][0], 16);
	k0[17] = t[0][0];
	for(int i = 1; i <= SIZE; i++){
		memcpy(&k0[18 * i], &sCW[i-1], 16);
		k0[18 * i + 16] = tCW[i-1][0];
		k0[18 * i + 17] = tCW[i-1][1];
	}
    memcpy(&k0[18 * SIZE + 18], &lastCW, 8);

	memcpy(k1, k0, 18 * SIZE + 18 + 8);
	memcpy(&k1[1], &s[0][1], 16);
	k1[17] = t[0][1];
}

uint64_t evalDPF(EVP_CIPHER_CTX *ctx, bool b, unsigned char* k, uint64_t x) {
	uint128_t s[SIZE + 1];
	int t[SIZE + 1];
	uint128_t sCW[SIZE];
	int tCW[SIZE][2];

	memcpy(&s[0], &k[1], 16);
	t[0] = b;

	for(int i = 1; i <= SIZE; i++){
		memcpy(&sCW[i-1], &k[18 * i], 16);
		tCW[i-1][0] = k[18 * i + 16];
		tCW[i-1][1] = k[18 * i + 17];
	}

	uint128_t sL, sR;
	int tL, tR;
	for(int i = 1; i <= SIZE; i++){
		dpfPRG(ctx, s[i - 1], &sL, &sR, &tL, &tR);

		if(t[i-1] == 1){
			sL = sL ^ sCW[i-1];
			sR = sR ^ sCW[i-1];
			tL = tL ^ tCW[i-1][0];
			tR = tR ^ tCW[i-1][1];
		}

		int xbit = getbit(x, i);
		
		if(xbit == 0){
			s[i] = sL;
			t[i] = tL;

		}else{
			s[i] = sR;
			t[i] = tR;
		}
	}
	// printBytes(&s[SIZE], 16);
    uint64_t res = convert(s[SIZE]);

    if(t[SIZE] %2 != 0) {
        //correction word
		uint64_t lastCW;
		memcpy(&lastCW, &k[18 * SIZE + 18], 8);

		// printf("last CW %d: %lu\n", b, lastCW);
        res = modAfterAdd(res + lastCW);
    }

	// printf("Res %d: %lu\n", b, res);

    if (b % 2 == 1) {
        // negate
        res = negate(res);
    }

    return res;
}
