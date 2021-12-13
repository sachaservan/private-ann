// From SabaEskandarian/OblivKeyValCode/
// From https://github.com/ucbrise/dory/blob/master/src/c/dpf.c

// This is the 2-party FSS evaluation function for point functions,
// based on "Function Secret Sharing: Improvements and Extensions." by Boyle, Elette, Niv Gilboa, and Yuval Ishai.  
// Proceedings of the 2016 ACM SIGSAC Conference on Computer and Communications Security. 
// ACM, 2016.

#include "../include/dpf.h"
#include <openssl/rand.h>

static inline uint128_t convert(uint128_t raw) {
	uint128_t r = raw & FIELDMASK;
	return r < FIELDSIZE ? r : r - FIELDSIZE;
}

static inline uint128_t reverse_lsb(uint128_t input) {
	return input ^ 1;
}

static inline uint128_t lsb(uint128_t input) {
    return input & 1;
}

static inline uint128_t set_lsb_zero(uint128_t input) {
    int lsb = (input & 1);
	if (lsb == 1){
		return reverse_lsb(input);
	}else{
		return input;
	}
}

static inline int getbit(uint128_t x, int size, int b) {
	return ((x) >> (size - b)) & 1;
}

static inline uint128_t negate(uint128_t x) {
	return x != 0 ? ((uint128_t)FIELDSIZE) - x : 0;
}

static inline uint128_t modAfterAdd(uint128_t r) {
	return r < FIELDSIZE ? r : r - ((uint128_t)FIELDSIZE);
}

static inline uint128_t dpf_reverse_lsb(uint128_t input){
    uint128_t xor = 1;
	return input ^ xor;
}

static inline uint128_t dpf_set_lsb_zero(uint128_t input){
    int lsb = input & 1;

	if(lsb == 1){
		return dpf_reverse_lsb(input);
	}else{
		return input;
	}
}

static void printBytes(void* p, int num) {
	unsigned char* c = (unsigned char*) p;
	for (int i = 0; i < num; i++) {
		printf("%02x", c[i]);
	}
	printf("\n");
}

EVP_CIPHER_CTX* getDPFContext(uint8_t* key) {
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

uint128_t getRandomBlock(){  
    uint128_t randBlock;    
	RAND_bytes((uint8_t*)&randBlock, 16);
	return randBlock;
}


// this is the PRG used for the DPF
void dpfPRG(EVP_CIPHER_CTX *ctx, uint128_t input, uint128_t* output1, uint128_t* output2, int* bit1, int* bit2){
	input = set_lsb_zero(input);

	uint128_t stashin[2];
	stashin[0] = input;
	stashin[1] = reverse_lsb(input);

    int len = 0;
	uint128_t stash[2];
	
    if (1 != EVP_EncryptUpdate(ctx, (uint8_t*)&stash[0], &len, (uint8_t*)&stashin[0], 32))
        printf("errors occured in encrypt\n");
  
	stash[0] = stash[0] ^ input;
	stash[1] = stash[1] ^ input;
	stash[1] = reverse_lsb(stash[1]);

	*bit1 = lsb(stash[0]);
	*bit2 = lsb(stash[1]);

	*output1 = dpf_set_lsb_zero(stash[0]);
	*output2 = dpf_set_lsb_zero(stash[1]);
}


void genDPF(EVP_CIPHER_CTX *ctx, int size, uint64_t index, unsigned char* k0, unsigned char *k1) {
	uint128_t seeds0[size+1];
	uint128_t seeds1[size+1];
	int bits0[size+1];
 	int bits1[size+1];

	uint128_t sCW[size];
	int tCW0[size];
	int tCW1[size];

    seeds0[0] = getRandomBlock();
	seeds1[0] = getRandomBlock();
	bits0[0] = 0;
	bits1[0] = 1;

    uint128_t s0[2], s1[2]; // 0=L,1=R
    int t0[2], t1[2];

	for(int i = 1; i <= size; i++){
		dpfPRG(ctx, seeds0[i-1], &s0[LEFT], &s0[RIGHT], &t0[LEFT], &t0[RIGHT]);
		dpfPRG(ctx, seeds1[i-1], &s1[LEFT], &s1[RIGHT], &t1[LEFT], &t1[RIGHT]);

		int keep, lose;
		int indexBit = getbit(index, size, i);
		if (indexBit == 0){
			keep = LEFT;
			lose = RIGHT;
		} else{
			keep = RIGHT;
			lose = LEFT;
		}

		sCW[i-1] = s0[lose] ^ s1[lose];

		tCW0[i-1] = t0[LEFT] ^ t1[LEFT] ^ indexBit ^ 1;
		tCW1[i-1] = t0[RIGHT] ^ t1[RIGHT] ^ indexBit;

		if (bits0[i-1] == 1){
			seeds0[i] = s0[keep] ^ sCW[i-1];
			if (keep == 0)
				bits0[i] = t0[keep] ^ tCW0[i-1];
			else 
				bits0[i] = t0[keep] ^ tCW1[i-1];
		} else{
			seeds0[i] = s0[keep];
			bits0[i] = t0[keep];
		}

		if (bits1[i-1] == 1){
			seeds1[i] = s1[keep] ^ sCW[i-1];
			if (keep == 0)
				bits1[i] = t1[keep] ^ tCW0[i-1];
			else
				bits1[i] = t1[keep] ^ tCW1[i-1];
		} else{
			seeds1[i] = s1[keep];
			bits1[i] = t1[keep];
		}
	}

  	uint128_t sFinal0 = convert(seeds0[size]);
	uint128_t sFinal1 = convert(seeds1[size]);
	uint128_t lastCW = modAfterAdd(1 + negate(sFinal0) + sFinal1);

	if (bits1[size] == 1) {
        lastCW = negate(lastCW);
    }
	
	k0[0] = 0;
	memcpy(&k0[1], seeds0, 16);
	k0[CWSIZE-1] = bits0[0];
	for(int i = 1; i <= size; i++){
		memcpy(&k0[CWSIZE * i], &sCW[i-1], 16);
		k0[CWSIZE * i + CWSIZE-2] = tCW0[i-1];
		k0[CWSIZE * i + CWSIZE-1] = tCW1[i-1];
	}
	memcpy(&k0[INDEX_LASTCW], &lastCW, 16);
	memcpy(k1, k0, INDEX_LASTCW + 16);
	memcpy(&k1[1], seeds1, 16); // only value that is different from k0
	k1[0] = 1; 
	k1[17] = bits1[0];
}

void batchEvalDPF(
	EVP_CIPHER_CTX *ctx, 
	int size, 
	bool b, 
	unsigned char* k, 
	uint64_t *in, 
	uint64_t inl, 
	uint8_t* out) {

	// parse the key 
	uint128_t seeds[size+1];
	int bits[size+1];
	uint128_t sCW[size+1];
	int tCW0[size];
	int tCW1[size];

	memcpy(&seeds[0], &k[1], 16);
	bits[0] = b;

	for(int i = 1; i <= size; i++){
		memcpy(&sCW[i-1], &k[18 * i], 16);
		tCW0[i-1] = k[18 * i + 16];
		tCW1[i-1] = k[18 * i + 17];
	}

	// [optimization]: because we're evaluating a whole batch of 
	// inputs we can cache the first X layers of the tree to avoid
	// evaluating the PRG again 
	int numCacheLayers = 12;
	int numCached = (1 << numCacheLayers);
	uint128_t *cachedSeeds = malloc(numCached * sizeof(uint128_t)); 
	int *cachedBits = malloc(numCached * sizeof(int)); 
	fullDomainDPF(ctx, numCacheLayers, b, k, cachedSeeds, cachedBits);
		
	// outter loop: iterate over all evaluation points 
	for (int l = 0; l < inl; l++) { 

		uint64_t idx = (in[l] >> (size - numCacheLayers)) & (numCached - 1);
		seeds[numCacheLayers] = cachedSeeds[idx];
		bits[numCacheLayers] = cachedBits[idx];

		uint128_t sL, sR;
		int tL, tR;
		for (int i = numCacheLayers+1; i <= size; i++){
			dpfPRG(ctx, seeds[i - 1], &sL, &sR, &tL, &tR);

			if (bits[i-1] == 1){
				sL = sL ^ sCW[i-1];
				sR = sR ^ sCW[i-1];
				tL = tL ^ tCW0[i-1];
				tR = tR ^ tCW1[i-1];
			}

			uint128_t xbit = getbit(in[l], size, i);
			
			//if (xbit == 0): seeds[i] = sL else seeds[i] = sR
			seeds[i] = (1-xbit) * sL + xbit * sR;
			bits[i] = (1-xbit) * tL + xbit * tR;		
		}
		
		uint128_t res = convert(seeds[size]);

		if (bits[size] == 1) {
			//correction word
			uint128_t lastCW;
			memcpy(&lastCW, &k[INDEX_LASTCW], 16);
			res = modAfterAdd(res + lastCW);
		}

		if (b == true) {
			// negate
			res = negate(res);
		}

		// copy block to byte output
		memcpy(&out[l*sizeof(uint128_t)], &res, sizeof(uint128_t));
	}

	free(cachedSeeds);
	free(cachedBits);
}


/* Need to allow specifying start and end for dataShare */
void fullDomainDPF(EVP_CIPHER_CTX *ctx, int size, bool b, unsigned char* k, uint128_t *outSeeds, int *outBits){

    //dataShare is of size dataSize
    int numLeaves = 1 << size;
	int maxLayer = size;

    int currLevel = 0;
    int levelIndex = 0;
    int numIndexesInLevel = 2;

    int treeSize = 2 * numLeaves - 1;

	uint128_t *seeds = malloc(sizeof(uint128_t)*treeSize); // treesize too big to allocate on stack
	int *bits = malloc(sizeof(int)*treeSize);
	uint128_t sCW[maxLayer+1];
	int tCW0[maxLayer+1];
	int tCW1[maxLayer+1];

	memcpy(seeds, &k[1], 16);
	bits[0] = b;

	for (int i = 1; i <= maxLayer; i++){
		memcpy(&sCW[i-1], &k[18 * i], 16);
		tCW0[i-1] = k[CWSIZE * i + CWSIZE-2];
		tCW1[i-1] = k[CWSIZE * i + CWSIZE-1];
	}

	uint128_t sL, sR;
	int tL, tR;
	for (int i = 1; i < treeSize; i+=2){
        int parentIndex = 0;
        if (i > 1) {
            parentIndex = i - levelIndex - ((numIndexesInLevel - levelIndex) / 2);
        }
		
        dpfPRG(ctx, seeds[parentIndex], &sL, &sR, &tL, &tR);

		if (bits[parentIndex] == 1){
			sL = sL ^ sCW[currLevel];
			sR = sR ^ sCW[currLevel];
			tL = tL ^ tCW0[currLevel];
			tR = tR ^ tCW1[currLevel];
		}

        int lIndex =  i;
        int rIndex =  i + 1;
        seeds[lIndex] = sL;
        bits[lIndex] = tL;
        seeds[rIndex] = sR;
        bits[rIndex] = tR;

        levelIndex += 2;
        if (levelIndex == numIndexesInLevel) {
            currLevel++;
            numIndexesInLevel *= 2;
            levelIndex = 0;
        }
    }

	for (int i = 0; i < numLeaves; i++) {
        int index = treeSize - numLeaves + i;
		outSeeds[i] = seeds[index];
		outBits[i] = bits[index];
    }

	free(bits);
	free(seeds);
}

