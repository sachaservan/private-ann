#!/bin/bash

# Runs the client, kills the servers, waits 10s and repeats the process, 100 times in total.  
#
# example usage: 
#     bash clicycle.sh
#

# remove results directory to avoid conflicts
rm -rf ../results

for run in {1..100}
   do
     bash client.sh; 
     sleep 5;
   done

# concatenate all temp files and delete them
rm -rf ../results/tmp
mkdir -p ../results/tmp
mv ../results/experiment* ../results/tmp/
python3 ./concat.py --dir ../results/tmp --out ../results/results.json
rm -rf ../results/tmp

# plot the results 
python3 ./plot.py --file ../results/results.json