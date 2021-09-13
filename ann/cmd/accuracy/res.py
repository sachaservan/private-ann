import os

directories = [f for f in os.listdir(".") if os.path.isdir(f)]

output = open("res.csv", "w")
columns = ["Dataset", "Tables", "Probes", "Lattice", "ApproximationFactor", "SequenceType", "Mode", "ProjectionWidthMean", "ProjectionWidthStddev"]
output.write(", ".join(columns))
output.write(", 95thCentile")
output.write(", trial1, trial2, trial3\n")

def ReadDic(s):
    pairs = s.split(" ")
    d = {}
    for p in pairs:
        p = p.replace("{", "").replace("}", "")
        parts = p.split(":")
        d[parts[0].strip()] = parts[1].strip()
    return d

def ReadArr(a):
    words = a.split(" ")
    arr = []
    for w in words:
        w = w.replace("[", "").replace("]", "")
        z = float(w.strip())
        if z != 0:
            arr.append(z)
    return arr

for d in directories:
    try:
        f = open(d + "/results.txt", "r")
        hits = []
        cur = {}
        distances = []
        r = 0
        for line in f.readlines():
            if line.startswith("{"):
                d = ReadDic(line)
                if not cur:
                    cur = d
                elif d != cur:
                    print("Multiple formats in same result file")
            if line.startswith("Hits"):
                parts = line.split(":")
                hits.append(parts[1].strip())
            if line.startswith("["):
                a = ReadArr(line)
                distances.extend([x for x in a if x > 0])
                r = r + len(a)

        distances.sort()
        centile = 0
        if len(distances) > 0.95 * r:
            centile = distances[int(0.95*r - 1)]
        output.write(", ".join([cur[x] for x in columns]))
        while len(hits) < 3:
            hits.append("")
        output.write(", " + str(centile) + ", " + ", ".join(hits) + "\n")
    except IOError:
        print("Ignoring " + d)
