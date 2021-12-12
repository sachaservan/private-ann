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
            bash server.sh --dataset ../datasets/deep1b --cachedir ../cache --numtables ${numtables} --hashrange 40 --numprobes ${numprobes} --bucketcap 1 --maxval 1000 --pwm 0.4094 --pws 0.097486 --procs ${numtables} "$@"
         done 
   done
