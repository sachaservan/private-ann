import sys
import os
import argparse
import numpy as np
import json
from plot_config import * # plot configuration file

width = 3 # default_width
height = 3 # default_height

fig = matplotlib.pyplot.gcf()
fig.set_size_inches(width, height)

directories = [f for f in os.listdir(".") if os.path.isdir(f)]

allResults = []
for d in directories:
    try:
        f = open(d + "/results.txt", "r")
        for file in os.listdir(d):
            try:
                if file.endswith(".json") and file.startswith("results"):
                    f = open(d + "/" + file, 'r') 
                    results = json.loads(f.read())
                    allResults.append(results)
            except IOError as e:
                print("Error reading file " + file + e)
    except IOError:
        print("Ignoring " + d)

tables =[1,5,10,15,20,30,50]
probes = [1,5,10,50,100]

for dataset in ['mnist', 'sift', 'deep1b', "gist"]:
    datasetResults = [r for r in allResults if r['Dataset'] == dataset]
    ax = plt.figure().gca()
    ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
    group_number = 0
    for p in probes:
        lineResults = [r for r in datasetResults if r['Probes'] == p]
        lineAccuracies = []
        lineStdDev = []
        for t in tables:
            tableAccuracies = []
            tableResults = [r for r in lineResults if r['Tables'] == t]
            for x in tableResults:
                if dataset == "gist":
                    tableAccuracies.append(float(x['Hits']) / 1000.0)
                else:
                    tableAccuracies.append(float(x['Hits']) / 10000.0)
            tableAccuracies = np.array(tableAccuracies)
            lineAccuracies.append(np.mean(tableAccuracies))
            lineStdDev.append(np.std(tableAccuracies))

        lineAccuracies = np.array(lineAccuracies)
        lineStdDev = np.array(lineStdDev)
        
        ax.plot(
            tables,
            lineAccuracies, 
            label=str(p),
            marker=markers[0],
            color=colors[group_number],
            lw=linewidth,
            ls="solid")

        lower = lineAccuracies - confidence95(lineStdDev, len(lineStdDev))
        upper = lineAccuracies + confidence95(lineStdDev, len(lineStdDev))

        plt.fill_between(
            tables,
            lower, 
            upper, 
            color=colors[group_number],
            alpha=error_opacity,
        )


        group_number = group_number + 1
    
    ax.legend(title="Probes")
    ax.set(xlabel='Tables', ylabel='Accuracy')
    if dataset == "deep":
        dataset = "deep1b"
    ax.set_title(dataset.upper()+" dataset")
    ax.figure.tight_layout()
    ax.figure.savefig(dataset + '_accuracy.pdf', bbox_inches='tight')
    ax.figure.savefig(dataset + '_accuracy.png', bbox_inches='tight')

