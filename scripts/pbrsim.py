import sys
import time
import math 
import argparse
import numpy as np
from plot_config import * # plot configuration file

width=4
height=4

def pbr_success(x, m):
   frac = m*(1 - (1  - 1/m)**x)/x
   return frac

# randomly assign items in y into n_buckets
def random_hash(y, n_buckets):
    counts = np.zeros(n_buckets)
    y_buckets = np.random.choice(np.arange(n_buckets), size=len(y))
    for i in range(len(y)):
        counts[y_buckets[i]] += y[i]
    return counts, y_buckets


probs = []
var = []
bucket_factors = [1, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5, 7, 7.5, 8, 8.5, 9]
for num_buckets_factor in bucket_factors:
    num_items = 100
    num_trials = 100
    avg_success = 0 

    all_trials = []
    for j in range(num_trials):
        y = np.ones(num_items)
        counts, buckets = random_hash(y, int(num_buckets_factor*len(y)))
        num_empty = int(num_buckets_factor*len(y)) - np.count_nonzero(counts) 
        
        if num_buckets_factor == 1:
            success = 1 - num_empty / len(y)
        else:
            success = (int(num_buckets_factor*len(y)) - len(y)) / num_empty 
       
        avg_success += success
        all_trials.append(success)

    expected_retrieved = (avg_success / num_trials)
    std_retrieved = np.std(np.array(all_trials))

    probs.append(expected_retrieved)
    var.append(2*std_retrieved/math.sqrt(num_trials))

probs = np.array(probs)
var = np.array(var)
print(probs)

ax = plt.figure().gca()
ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
fig = matplotlib.pyplot.gcf()
fig.set_size_inches(width, height)

ax.plot(
    bucket_factors, 
    probs, 
    color=colors[3],
    lw=linewidth,
)

plt.fill_between(
        bucket_factors,
        probs - var, 
        probs + var, 
        color=colors[3],
        alpha=error_opacity,
    )

ax.set_xlabel(r'Partition overhead ($\frac{m}{\ell}$)')
ax.set_title('Fraction retreived')
ax.set_xticks([1, 3, 5, 7, 9])
ax.set_ylim([0.0, 1.0])
ax.figure.tight_layout()
ax.figure.savefig('pbc.pdf', bbox_inches='tight')

############################################
# PBR overhead plot
############################################
ax = plt.figure().gca()
ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
fig = matplotlib.pyplot.gcf()
fig.set_size_inches(width, height)

x = np.array([x + 1 for x in range(0, 19)])
x_ticks = np.array([1, 5, 10, 15, 20])

ybc = ((3/2)**np.log2(x)) #  (3/2)^log(x)
ypbc = np.ones(len(x)) * 3 # 3x
ypbr1 = np.ones(len(x)) * 1.0/pbr_success(x, x)
ypbr2 = np.ones(len(x)) * 1.0/pbr_success(x, 2*x)

ax.plot(
    x, 
    x, 
    label="naive",
    color='grey',
    linestyle='dotted',
    lw=linewidth,
)

ax.plot(
    x, 
    ybc, 
    label="BC",
    color=colors[0],
    lw=linewidth,
)

ax.plot(
    x,
    ypbc, 
    label="PBC",
    color=colors[2],
    lw=linewidth,
)


ax.plot(
    x, 
    ypbr1, 
    label=r'PBR ($m=\ell$)',
    color=colors[3],
    lw=linewidth,
)


ax.plot(
    x, 
    ypbr2, 
    label=r'PBR ($m=2\ell$)',
    color=colors[4],
    lw=linewidth,
)


ax.annotate(text='Naive', xy=(0.7, 2.25), xycoords='data', rotation=80)
ax.annotate(text='PBC', xy=(10, 3.05), xycoords='data')
ax.annotate(text='SBC', xy=(6, 3.2), xycoords='data', rotation=62)
ax.annotate(text=r'PBR ($m = $2$\ell$)', xy=(11, 1.3), xycoords='data')
ax.annotate(text=r'PBR ($m = \ell$)', xy=(11, 1.6), xycoords='data')

ax.set_xlabel(r'$\ell$')
ax.set_ylabel('Processing Factor')
ax.set_ylim([1, 4])
ax.figure.tight_layout()
ax.set_xticks(x_ticks)
ax.set_yticks([1, 2, 3, 4])
ax.figure.savefig('pbc_comparisons_work.pdf', bbox_inches='tight')


############################################
# PBR overhead plot
############################################
ax = plt.figure().gca()
ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
fig = matplotlib.pyplot.gcf()
fig.set_size_inches(width, height)

x = np.array([x + 1 for x in range(0, 19)])
x_ticks = np.array([1, 5, 10, 15, 20])

#   comm overhead  +  bucket overhead
#  log( (3)^log(x) / (3/2)^log(x) ) + (3/2)^log(x)  = total comm
ybc = ((3)**np.log2(x))/x
ypbc = np.ones(len(x)) * 1.5*3 # 1.5k buckets, 3 values per bucket due to replication
ypbr1 = np.ones(len(x)) * 1.0/pbr_success(x, x)
ypbr2 = np.ones(len(x)) * 2.0/pbr_success(x, 2*x)

ax.plot(
    x, 
    ybc, 
    label="BC",
    color=colors[0],
    lw=linewidth,
)

ax.plot(
    x,
    ypbc, 
    label="PBC",
    color=colors[2],
    lw=linewidth,
)

ax.plot(
    x, 
    ypbr1, 
    label=r'PBR ($m=\ell$)',
    color=colors[3],
    lw=linewidth,
)

ax.plot(
    x, 
    ypbr2, 
    label=r'PBR ($m=2\ell$)',
    color=colors[4],
    lw=linewidth,
)

ax.annotate(text='PBC', xy=(6, 4.6), xycoords='data')
ax.annotate(text='SBC', xy=(6, 3.2), xycoords='data', rotation=50)
ax.annotate(text=r'PBR ($m = $2$\ell$)', xy=(11, 2.58), xycoords='data')
ax.annotate(text=r'PBR ($m = \ell$)', xy=(11, 1.64), xycoords='data')

ax.set_xlabel(r'$\ell$')
ax.set_ylim([1, 5])
ax.set_ylabel('Communication Factor')
ax.set_xticks(x_ticks)
ax.figure.tight_layout()
ax.figure.savefig('pbc_comparisons_comm.pdf', bbox_inches='tight')
