import os 
import argparse
import sys


def concat(directory, outputfile):
    filenames = []
    for filename in os.listdir(directory):
        if filename.endswith(".json"): 
            filenames.append(os.path.join(directory, filename))
            continue

    print(filenames)


    fno = 0
    with open(outputfile, 'w') as outfile:
        outfile.write("[")
        for fname in filenames:
            fno += 1
            with open(fname) as infile:
                outfile.write(infile.read())
            if fno  < len(filenames):
                outfile.write(",")
        outfile.write("]")

if __name__ == '__main__':
    argparser = argparse.ArgumentParser(sys.argv[0])
    argparser.add_argument("--dir", type=str, default='')
    argparser.add_argument("--out", type=str, default='')
    args = argparser.parse_args()

    concat(args.dir, args.out)
