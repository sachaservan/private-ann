# Private Similarity Search
## Prototype implementation of lightweight private similarity search

### Reproducing paper experiments 

### Dependencies 
- GMP Library: On Ubuntu run ```sudo apt-get install libgmp3-dev```.  On yum, ```sudo yum install gmp-devel```.
- Go v1.16 or later: On Ubuntu run ```sudo apt-get install golang-go```.  On yum, ```sudo yum install golang```.
- OpenSSL: On Ubuntu run ```sudo apt-get install libssl-dev```.  On yum, ```sudo yum install openssl-devel```.
- Make: On Ubuntu run ```sudo apt install make```.  On yum, ```sudo yum install make```.

For optimal performance, you should compile the C code with clang-12 (approximately 10-20 percent faster than the default in some distributions).
- Clang-12: On Ubuntu run ```sudo apt install clang-12```.
You'll also need llvm if you use clang.  On yum, ```sudo yum install clang```.
- LLVM-AR: On Ubuntu run ```sudo apt install llvm```. On yum, ```sudo yum install llvm```.

### Datasets 
All datasets used in the paper can be obtained from https://github.com/erikbern/ann-benchmarks.

The raw data is in HDF5 format which can be converted to CSV using ```/datasets/dataconv.py``` script. 
The python script generates three files prefixed by the dataset name. 
For example, ```python dataconv.py deep1b.hdf5``` will output *deep1b_train.csv*, *deep1b_test.csv*, and *deep1b_neighbors.csv*. 

The bash script argument requires ```DATASET_PATH``` point to the directory where these three files are located as well as the dataset name predix. 
For example, to run the server on the *deep1b* data, set```DATASET_PATH=/home/user/datasets/deep1b``` (note the lack of suffix in the dataset file name).
The code will automatically locate and use the training data to build the data structure and the test data as "queries" issued by clients. 


### Running the servers 
Both servers must have access to the same datasets so that they can locally compute the necessary data structure. 
Each server is run automatically using the provided scripts.
There is one script per dataset. 
Each script will cycle through all experimental configurations (e.g., number of hash tables, number of probes, etc.).

#### On each server machine
1. Compile the C DPF library which is used by the Go code. 
```
cd ~/go/src/private-ann/pir/dpfc/src
make
```

2. Set the C compiler to the corresponding compiler for cgo compilation.

```
export CC=clang-12
```

3. Download and process the datasets, placing each dataset into ```~/go/src/private-ann/datasets/```.


#### On server machine A
```
cd scripts
bash mnist.sh --sid 0 --procs NUMBER_OF_CORES 
```

#### On server machine B
```
cd scripts
bash mnist.sh --sid 1 --procs NUMBER_OF_CORES 
```

#### Running the client 
After configuring ```client.sh``` with the server IP addresses, run
```
bash client.sh
```
which will query the servers and save the experiment results to ```../results/``` under a random ```.json``` file. 
To continuously query the servers (until all parameter combinations are exhausted), run 
```
bash clicycle.sh
```
which will spin up a new client once the servers have initialized the new experiment configuration.

### Finding dataset parameters

First go to the parameters directory
```
cd ann/cmd/parameters
go build
```
To find the mean and standard deviation of the brute force distances for a dataset
```
./parameters --dataset ../../../datasets/mnist --dimension 24
```
The dimension 24 argument cause dimensionality reduction to 24 dimensions before calculating the distances,
so the radii will account for the variance introduced by the reduction.
This is the form expected by the implementation of the 24 dimensional Leech Lattice LSH.

### Checking hash function accuracy

First go to the accuracy directory
```
cd ann/cmd/accuracy
go build
```
The accuracy script accepts parameters for many aspects of the LSH.  For example
```
./accuracy --dataset=../../../datasets/mnist --tables=10 --probes=30 --projectionwidthmean=887.77 --projectionwidthstddev=244.92 --mode=test --sequencetype=normal2
```
Evaluates the 10000 test queries for accuracy under approximation factor 2 for the MNIST dataset.
The values of width and stddev are those found with the parameter program.
To use training data to modify parameters, first run the parameter program to generate an answer set, move it into the directory, and use --mode=train.
Sequence type provides slightly different options for computing the radii.

The test.py python file contains the parameters used to run the experiments.


## Important Warning
This implementation is intended as a proof-of-concept prototype only! The code was implemented for research purposes and has not been vetted by security experts. As such, no portion of the code should be used in any real-world or production setting!

## License
Copyright © 2021 Sacha Servan-Schreiber

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
