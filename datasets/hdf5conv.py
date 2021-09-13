''' 
Script for converting .hdf5 files into the types expected. 
Assumes input files are formatted as in https://github.com/erikbern/ann-benchmarks
That is, each file consists of the training data 'train', test data 'test' and neighbor indices 'neighbors'.

This script will generate 3 files named args.name_[type].csv

Example:

python hdf5conv.py --dataset sift.hdf5 --name sift --newdim 50

will generate files sift_train.csv sift_test.csv sift_neighbors.csv
where the new dimensionality of the data is 50. 
'''
import argparse
import math 
import multiprocessing.pool
import os
import random
import shutil
import sys
import traceback
import h5py
import numpy as np 

def get_dataset(which):
    hdf5_fn = get_dataset_fn(which)
    hdf5_f = h5py.File(hdf5_fn, 'r')
    return hdf5_f

def get_dataset_fn(dataset):
    return dataset

def reduce_dimensionality(data, compression_matrix):
    new_data = []
    for i, vec in enumerate(data):
        new_vec = np.dot(compression_matrix, vec)
        new_data.append(new_vec)
    return np.array(new_data) 

def main():
    argparser = argparse.ArgumentParser(sys.argv[0])
    argparser.add_argument("--dataset", type=str, help=".hdf5 data for testing")
    argparser.add_argument("--name", type=str, help="base name for output files")
    argparser.add_argument("--maxsize", type=int, default=-1, help="max number of items to process")
    argparser.add_argument("--scalefactor", type=float, default=1, help="scale factor for quantization")
    argparser.add_argument("--newdim", type=int, default=0, help="final dim of vectors")
    args = argparser.parse_args()

    
    dataset = get_dataset(args.dataset)
    dimension = len(dataset['train'][0]) 
    point_type = dataset.attrs.get('point_type', 'float')

    print("keys " + str(dataset.keys()))
    print(dimension)
    print(dataset['test'])
    print(dataset['train'])
    print(dataset['neighbors'])

    train_arr = np.array(dataset['train'])
    test_arr = np.array(dataset['test'])
    neighbors_arr = np.array(dataset['neighbors'])

    maxsize = args.maxsize
    if args.maxsize == -1:
        maxsize = len(train_arr)

    train_arr = train_arr[:maxsize]

    if args.newdim > 0:
        print("Reducing data dimensionality")
        curr_dim = len(train_arr[0])
        compression_matrix = np.random.normal(0, 1, (args.newdim, curr_dim))
        train_arr = reduce_dimensionality(train_arr, compression_matrix)
        test_arr = reduce_dimensionality(test_arr, compression_matrix)

    print("Quantizing data")
    train_arr *= args.scalefactor
    test_arr *= args.scalefactor
    
    print("Saving CSVs")

    np.savetxt(args.name + '_test.csv', test_arr, delimiter=',', fmt='%lf')
    np.savetxt(args.name + '_train.csv', train_arr, delimiter=',', fmt='%lf')
    np.savetxt(args.name + '_neighbors.csv', neighbors_arr, delimiter=',', fmt='%lf')


if __name__ == '__main__':
    main()