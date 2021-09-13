from scipy.stats import norm
import numpy as np

datasets = ["mnist", "sift", "deep1b", "gist"]
datasetSize = [60000, 1000000, 10000000, 1000000]
rawDimension = [784, 128, 96, 960]

# Here we use the values in the original dataset's dimension
# Elsewhere you will see the values from sampling post dimensionality reduction
means = [1065.57, 150.46, 0.5488, 0.6415] # average of 10000 brute force nearest neighbors
stddevs = [307.9, 58.63, 0.1463, 0.24064] # stddev of above

dMins = [209.55, 7.48, 0.01686, 0.0170] # Smallest seen in sample of 10000 sorting of brute force distances
dMaxes = [4303.3, 721.34, 1.74, 8.632] # Largest seen in sample of 10000 sorting of brute force distances

# estimation from https://dl.acm.org/doi/pdf/10.1145/2783258.2783405
# Figure 3 for mnist, sift (deep, gist) TODO
intrinsicDimension = [10, 15, 20, 25]

# Average number of zeros actually seen in the 10000 queries
empiricL = [15, 15, 15, 15]

# We use an Rmax based on the assumed normal distribution
# This means we chop off the upper tail so that
# Each table should be used by 1/L of the points to maximize efficiency
def RMax(datasetNumber, numTables):
    numTables = float(numTables)
    quantile = norm.ppf((numTables - 0.5) / numTables)
    return means[datasetNumber] + quantile * stddevs[datasetNumber]

def IdealLeakage(dimension, rmax, dmax, n):
    return dimension * np.log2(dmax / rmax) + np.log2(n)

def ExtraLeakage(dimension, rmax, dmin, l):
    return dimension * np.log2(rmax / dmin) + np.log2(l)

def LeakageRatio(dimension, rmax, dmax, dmin, n, l):
    return ExtraLeakage(dimension, rmax, dmin, l) / IdealLeakage(dimension, rmax, dmax, n)


if __name__ == "__main__":
    numTables = 30
    for idx, dataset in enumerate(datasets):
        ratio = LeakageRatio(rawDimension[idx], 
            RMax(idx, numTables),
            dMaxes[idx],
            dMins[idx],
            datasetSize[idx],
            numTables - 1
        )
        empiricRatio = LeakageRatio(intrinsicDimension[idx], 
            RMax(idx, numTables),
            dMaxes[idx],
            dMins[idx],
            datasetSize[idx],
            empiricL[idx]
        )
        print("dataset %s: raw leakage: %f empiricLeakage: %f rmax: %f" % (dataset, ratio, empiricRatio, RMax(idx, numTables)))
