import os

# Iterate over all datasets and parameters

tables = [1,5,10,15,20,25,30,50]
probes = [0] #Reuse tables for different levels of probing, see go code
trials = 3
datasets = ["mnist", "sift", "deep1b", "gist"]
means = [887.77, 129.30, 0.4094, 0.6415]
stddevs = [244.92, 43.46, 0.097486, 0.24064]


for trial in range(trials):
    for numTables in tables:
        for numProbes in probes:
            for idx, datasetName in enumerate(datasets):  
                command = "./accuracy"
                command += " --dataset=../../../datasets/" + datasetName
                command += " --tables=" + str(numTables)
                command += " --probes=" + str(numProbes)
                command += " --projectionwidthmean=" + str(means[idx])
                command += " --projectionwidthstddev=" + str(stddevs[idx])
                command += " --mode=test"
                command += " --sequencetype=normal2"
                print(command)
                ret = os.system(command)
                if ret != 0:
                    print("Error non zero exit code: " + str(ret))
                    quit()

