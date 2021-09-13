import sys
import argparse
import numpy as np
from scipy.stats import norm
import json
from plot_config import * # plot configuration file

width = 4.5 # default_width
height = 3.5 # default_height


def plot_distribution_and_radii(x, num_bins, num_tables):

    fig, ax = plt.subplots(figsize=(width, height))
    ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
    ax.set_axisbelow(True)

    # the histogram of the data
    n, bins, patches = ax.hist(x, num_bins, density=True)

    mu = np.mean(x)
    sigma = np.std(x)

    # add a 'best fit' line
    y = ((1 / (np.sqrt(2 * np.pi) * sigma)) *
        np.exp(-0.5 * (1 / sigma * (bins - mu))**2))
    
    for r in range(num_tables):
        z = norm.ppf((float(r)+0.5)/float(num_tables))
        ax.axvline(x = (mu + z * sigma), color='black', lw=0.75, linestyle='dashed')

    
    ax.plot(bins, y, '--', lw=2)

    return ax

def parse_data(a):
    words = a.split(" ")
    arr = []
    for w in words:
        w = w.replace("[", "").replace("]", "")
        z = float(w.strip())
        if z != 0:
            arr.append(z)
    return np.array(arr)


if __name__ == '__main__':
    argparser = argparse.ArgumentParser(sys.argv[0])
    argparser.add_argument("--file", type=str, default='')
    argparser.add_argument("--name", type=str, default='')
    argparser.add_argument("--numbins", type=int, default=50)
    argparser.add_argument("--numtables", type=int, default=10)
    argparser.add_argument("--maxdist", type=float, default=-1)

    args = argparser.parse_args()

    # read experiment file (expected json)
    with open(args.file, 'r') as myfile:
        data=myfile.read()

    x = parse_data(data)

    ax = plot_distribution_and_radii(x, args.numbins, args.numtables)
    ax.set_xlabel('Distance')
    ax.set_ylabel('Probability density')
    ax.set_title(args.name.upper() + " dataset")
    if args.maxdist != -1:
        ax.set_xlim([0,args.maxdist])
    ax.figure.tight_layout()
    ax.figure.savefig(args.name + "_radii.pdf")