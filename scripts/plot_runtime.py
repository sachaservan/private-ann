import sys
import argparse
import numpy as np
import json
from plot_config import * # plot configuration file

width = 4.5 # default_width
height = 3.5 # default_height

def plot_runtime(x, y, z, group_labels, group_size, nolegend=False):
    ######################## PLOT CODE ########################
    ax = plt.figure().gca()
    ax.yaxis.grid(color=gridcolor, linestyle=linestyle)
    fig = matplotlib.pyplot.gcf()
    fig.set_size_inches(width, height)

    xticks = np.sort(np.unique(x))

    secondary_labels = ['Parallelized', '1 CPU']
    line_styles = ['solid', 'dashed']

    group_number = 0
    for i in range(0, len(y), group_size): 
        sort = np.argsort(x[i:i+group_size])

        # multi-core plot
        ax.plot(
            x[i:i+group_size][sort], 
            y[:,0][i:i+group_size][sort], 
            marker=markers[0],
            color=colors[group_number],
            lw=linewidth,
            ls=line_styles[1],
        )

        plt.fill_between(
            x[i:i+group_size][sort],
            y[:,0][i:i+group_size][sort] - y[:,1][i:i+group_size][sort], 
            y[:,0][i:i+group_size][sort] + y[:,1][i:i+group_size][sort], 
            color=colors[group_number],
            alpha=error_opacity,
        )

        # set the "ghost" labels
        ax.plot(np.NaN, np.NaN,     
            label=group_labels[group_number],
            marker=markers[0],
            color=colors[group_number],
            lw=linewidth,
            ls=line_styles[0],
        )

        # compute parallelized time  
        server_time = z[:,0][i:i+group_size][sort][0] # server time on one table
        num_tables = x[i:i+group_size][sort]
        server_time_single = server_time * num_tables
        other_latency = y[:,0][i:i+group_size][sort] - server_time
        px = x[i:i+group_size][sort]
        py = other_latency + server_time_single

        # single core plot
        ax.plot(
            px, 
            py, 
            marker=markers[0],
            color=colors[group_number],
            lw=linewidth,
            ls=line_styles[0],
        )
        
        group_number += 1

    if not nolegend:
        ax2 = ax.twinx() # twin ghost axis to overlay
        ax2.plot(np.NaN, np.NaN,     
            label=secondary_labels[0],
            ls=line_styles[1],
            color='black',
        )
        ax2.plot(np.NaN, np.NaN,     
            label=secondary_labels[1],
            ls=line_styles[0],
            color='black',
        )
        ax2.get_yaxis().set_visible(False)

        ax.legend(title='Probes',loc='upper left',  edgecolor='white', framealpha=1, fancybox=False)
        ax2.legend(loc='upper center',  edgecolor='white', framealpha=1, fancybox=False)

    ax.set_xticks(xticks)
    return ax



if __name__ == '__main__':
    argparser = argparse.ArgumentParser(sys.argv[0])
    argparser.add_argument("--file", type=str, default='')
    argparser.add_argument("--cap", type=int, default=-1)
    argparser.add_argument("--nolegend", type=bool, nargs='?', const=True, default=False)

    args = argparser.parse_args()

    # read experiment file (expected json)
    with open(args.file, 'r') as myfile:
        data=myfile.read()

    # parse the experiment file as json
    results = json.loads(data)

    num_tables = []
    num_probes = []
    server_time_ms = []
    client_latency_ms = []
    bandwidth_up_bytes = []
    bandwidth_down_bytes = []
    bandwidth_total_bytes = []

    num_results = 0
    num_trials = len(results[0]["query_up_bandwidth_bytes"])
    dataset = results[0]["dataset_name"]

    # first we extract the relevent bits 
    for i in range(len(results)):

        num_probes.append(results[i]["num_probes"])
        num_tables.append(results[i]["num_tables"])
        
        # extract and compute bandwidth statistics 
        bandwidth_up = np.array(results[i]["query_up_bandwidth_bytes"])
        bandwidth_down = np.array(results[i]["query_down_bandwidth_bytes"])
        bandwidth_total = bandwidth_up + bandwidth_down
        avg = np.mean(bandwidth_total)
        std = np.std(bandwidth_total)
        bandwidth_total_bytes.append([avg, confidence95(std, num_trials)])

        avg_up = np.mean(bandwidth_up)
        std_up = np.std(bandwidth_up)
        bandwidth_up_bytes.append([avg_up, confidence95(std_up, num_trials)])

        avg_down = np.mean(bandwidth_down)
        std_down = np.std(bandwidth_down)
        bandwidth_down_bytes.append([avg_down, confidence95(std_down, num_trials)])

        avg = np.mean(results[i]["query_client_ms"]) 
        std = np.std(results[i]["query_client_ms"])
        client_latency_ms.append([avg, confidence95(std, num_trials)])

        server_total = np.array(results[i]["dpf_server_ms"]) + np.array(results[i]["masking_server_us"])*MICRO_TO_MILLI
        avg = np.mean(server_total)
        std = np.std(server_total)
        server_time_ms.append([avg, confidence95(std, num_trials)])

        num_results += 1        

    # convert everything to numpy arrays
    num_tables = np.array(num_tables)
    num_probes = np.array(num_probes)
    server_time_ms = np.array(server_time_ms)
    client_latency_ms = np.array(client_latency_ms)
    bandwidth_total_bytes = np.array(bandwidth_total_bytes)
    bandwidth_down_bytes = np.array(bandwidth_down_bytes)
    bandwidth_up_bytes = np.array(bandwidth_up_bytes)

    group_size = len(np.unique(num_tables))

    # make the ANN processing as a function of table size plots 
    sort = np.argsort(num_probes)
    num_tables = num_tables[sort][:args.cap * group_size]
    num_probes = num_probes[sort][:args.cap * group_size]
    server_time_ms = server_time_ms[sort][:args.cap * group_size]
    client_latency_ms = client_latency_ms[sort][:args.cap * group_size]
    bandwidth_total_bytes = bandwidth_total_bytes[sort][:args.cap * group_size]
    bandwidth_down_bytes = bandwidth_down_bytes[sort][:args.cap * group_size]
    bandwidth_up_bytes = bandwidth_up_bytes[sort][:args.cap * group_size]

    # figure out how many different groups we have 
    group_size = len(np.unique(num_tables))
    group_labels = np.array([str(i) for i in np.unique(num_probes)])

    # plot client end-to-end time 
    ax = plot_runtime(num_tables, client_latency_ms*MILLI_TO_SECONDS, server_time_ms*MILLI_TO_SECONDS, group_labels, group_size, args.nolegend)
    ax.set_xlabel('Number of hash tables')
    ax.set_ylabel('Client latency (seconds)')
    #ax.set_yscale("log", base=10)
    ax.set_ylim(0, ax.get_ylim()[1] * 1.25) # make y axis 25% bigger
    ax.set_title(dataset.upper() + " dataset")
    # leg = ax.legend(title='Probes', loc="best", framealpha=1, edgecolor=edgecolor)
    ax.figure.tight_layout()
    ax.figure.savefig(dataset + '_latency_client.pdf', bbox_inches='tight')


    sort = np.argsort(num_tables)
    group_size_probes = len(np.unique(num_probes))
    print("Bandwidth (kB) per #probes (1 table): " + str(np.sort(bandwidth_total_bytes[:,0][sort][0:group_size_probes])*BYTES_TO_KB))
    
