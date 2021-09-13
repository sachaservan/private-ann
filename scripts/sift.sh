#!/bin/bash

# Runs the server on each config for evaluating performance 
#
# example usage: 
#     bash mnist.sh --sid 0 --procs 1 
#
for numtables in 1 5 10 20 30 40 50
   do
      for numprobes in 1 5 10 50 100 500
         do 
            # "$@" contains all parameters that are passed to the script (and do not change between experiments)
            bash server.sh --dataset ../datasets/sift --cachedir ../cache --numtables ${numtables} --numprobes ${numprobes} --bucketcap 1 --maxval 1000 --pwm 129.30 --pws 43.46 --procs ${numtables} "$@"
         done 
   done
