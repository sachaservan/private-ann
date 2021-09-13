package main

// For a sample of the training data,
// Calculate the mean and standard deviation of the (true) distances to nearest neighbors
// Use this to choose the distribution of LSH radii

// Set --dimension=24 to precompute dimensionality reduction
// This includes the variance from the dimensionality reduction in the standard deviation
// And should be the radii actually used for the LSH radii with the 24-dimensional Leech lattice hash
// (Rather than attempting to compute the effect later, we can just empirically see the result)

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/montanaflynn/stats"
	"github.com/sachaservan/private-ann/ann"
	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/vec"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {

	// command-line arguments to the server
	var args struct {
		Dataset   string `default:"../../../datasets/mnist"`
		Samples   int    `default:"10000"`
		Dimension int    `default:"0"`
	}

	arg.MustParse(&args)
	dataset := args.Dataset
	numSamples := args.Samples
	transformDim := args.Dimension
	// only use training data
	data, _, _, err := ann.ReadDataset(dataset)
	if err != nil {
		panic(err)
	}

	path := strings.Split(args.Dataset, "/")
	name := fmt.Sprintf("%v%vx%v", path[len(path)-1], transformDim, numSamples)

	done := make(chan error)
	numThreads := runtime.NumCPU()

	inputDim := data[0].Size()
	if transformDim > 0 && transformDim <= inputDim {
		r := hash.NewHashCommon(inputDim, transformDim, 0, true)
		scaleFactor := math.Sqrt(float64(inputDim) / float64(transformDim))
		spans := hash.Spans(len(data), numThreads)
		for t := 0; t < numThreads; t++ {
			go func(t int) {
				for row := spans[t][0]; row < spans[t][1]; row++ {
					data[row] = r.Project(data[row]).Scale(scaleFactor)
				}
				done <- nil
			}(t)
		}
		for i := 0; i < numThreads; i++ {
			<-done
		}
	} else {
		fmt.Printf("Not transforming or rotating\n")
	}

	if numSamples >= len(data) {
		fmt.Printf("Computing all pairwise distances\n")
		Exact(data, name)
	} else {

		results := make(map[int]float64)
		neighbors := make(map[int]int)
		farResults := make(map[int]float64)
		mu := sync.Mutex{}
		for t := 0; t < numThreads; t++ {
			go func() {
				mu.Lock()
				for len(results) < numSamples {
					index := rand.Intn(len(data))
					if results[index] != 0 {
						// try again
						continue
					} else {
						results[index] = -1
					}
					mu.Unlock()

					bestDist := math.MaxFloat64
					worstDist := float64(0)
					bestNeighbor := -1
					for j := 0; j < len(data); j++ {
						d := 0.0
						for k := 0; k < len(data[index].Coords); k++ {
							di := data[index].Coords[k] - data[j].Coords[k]
							d += di * di
						}
						if d < bestDist && d != 0 {
							bestDist = d
							bestNeighbor = j
						}
						if d > worstDist {
							worstDist = d
						}
					}

					mu.Lock()
					results[index] = bestDist
					neighbors[index] = bestNeighbor
					farResults[index] = worstDist
				}
				mu.Unlock()
				done <- nil
			}()
		}
		for i := 0; i < numThreads; i++ {
			<-done
		}

		dists := make([]float64, len(results))
		farthestDists := make([]float64, len(results))
		closestDist := math.MaxFloat64
		farthestDist := float64(0)
		farthestClosestDist := float64(0)
		i := 0
		for _, v := range results {
			dists[i] = math.Sqrt(v)
			if dists[i] < closestDist {
				closestDist = dists[i]
			}
			if dists[i] > farthestClosestDist {
				farthestClosestDist = dists[i]
			}
			i++
		}
		i = 0
		for _, v := range farResults {
			farthestDists[i] = math.Sqrt(v)
			if farthestDists[i] > farthestDist {
				farthestDist = farthestDists[i]
			}
			i++
		}
		fmt.Printf("Closest distance (Dmin, Rmin): %v\n", closestDist)
		fmt.Printf("Farthest distance (Dmax): %v\n", farthestDist)
		fmt.Printf("Farthest closest distance (Rmax): %v\n", farthestClosestDist)
		output(dists, neighbors, name)
		output(farthestDists, make(map[int]int), name+"farthest")

		f, err := os.Create(name + "Data.txt")
		if err != nil {
			panic(err)
		}
		w := bufio.NewWriter(f)
		w.WriteString(fmt.Sprintf("%v\n", dists))
		w.Flush()
	}
}

type Neighbors struct {
	closest          []int
	farthest         []int
	closestDistance  []float64
	farthestDistance []float64
}

// more than runtime.NumCPU
var locks = [128]sync.Mutex{}

func Exact(data []*vec.Vec, name string) {
	jobs := make(chan int, len(data))
	for i := 0; i < len(data); i++ {
		jobs <- i
	}
	close(jobs)

	neighbors := NewNeighborStruct(len(data))
	done := make(chan error)

	numThreads := runtime.NumCPU()
	for t := 0; t < numThreads; t++ {
		go func() {
			for i := range jobs {
				localBest := math.MaxFloat64
				for j := i + 1; j < len(data); j++ {
					// compute the squared distance between i and j
					d := 0.0
					for k := 0; k < len(data[i].Coords); k++ {
						di := data[i].Coords[k] - data[j].Coords[k]
						d += di * di
					}
					if d < localBest {
						neighbors.CheckAndUpdate(i, j, d)
						localBest = d
					}
					neighbors.CheckAndUpdate(j, i, d)
				}
			}
			done <- nil
		}()
	}
	for i := 0; i < numThreads; i++ {
		<-done
	}
	closestDist := math.MaxFloat64
	farthestDist := float64(0)
	farthestClosestDist := float64(0)
	for i, v := range neighbors.closestDistance {
		neighbors.closestDistance[i] = math.Sqrt(v)
		if neighbors.closestDistance[i] < closestDist {
			closestDist = neighbors.closestDistance[i]
		}
		if neighbors.closestDistance[i] > farthestClosestDist {
			farthestClosestDist = neighbors.closestDistance[i]
		}
		neighbors.farthestDistance[i] = math.Sqrt(neighbors.farthestDistance[i])
		if neighbors.farthestDistance[i] > farthestDist {
			farthestDist = neighbors.farthestDistance[i]
		}
	}
	fmt.Printf("Closest distance (Dmin, Rmin): %v\n", closestDist)
	fmt.Printf("Farthest distance (Dmax): %v\n", farthestDist)
	fmt.Printf("Fathest closest distance (Rmax): %v\n", farthestClosestDist)
	n := make(map[int]int)
	for i, v := range neighbors.closest {
		n[i] = v
	}
	output(neighbors.closestDistance, n, name+"closest")
	n = make(map[int]int)
	for i, v := range neighbors.farthest {
		n[i] = v
	}
	output(neighbors.farthestDistance, n, name+"farthest")
}

func NewNeighborStruct(size int) *Neighbors {
	n := &Neighbors{
		closest:          make([]int, size),
		farthest:         make([]int, size),
		closestDistance:  make([]float64, size),
		farthestDistance: make([]float64, size),
	}
	for i := range n.closest {
		n.closest[i] = -1
		n.farthest[i] = -1
		n.closestDistance[i] = math.MaxFloat64
	}
	return n
}

func (n *Neighbors) CheckAndUpdate(index1 int, index2 int, dist float64) {
	lockNum := index1 % 32
	locks[lockNum].Lock()
	defer locks[lockNum].Unlock()
	if dist < n.closestDistance[index1] {
		n.closestDistance[index1] = dist
		n.closest[index1] = index2
	}
	if dist > n.farthestDistance[index1] {
		n.farthestDistance[index1] = dist
		n.farthest[index1] = index2
	}
}

func output(distances []float64, neighbors map[int]int, name string) {

	mean, _ := stats.Mean(distances)
	stddev, _ := stats.StandardDeviationSample(distances)
	fmt.Printf("mean: %v, stddev: %v\n", mean, stddev)
	histPlot(distances, "distances"+name)

	f, err := os.Create("neighbors" + name + ".txt")
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("mean: %v, stddev: %v\n", mean, stddev))
	for k, v := range neighbors {
		w.WriteString(fmt.Sprintf("%v:%v\n", k, v))
	}
	w.WriteString(fmt.Sprintf("%v\n", distances))
	w.Flush()
}

func histPlot(values []float64, title string) {
	p := plot.New()
	p.Title.Text = title
	hist, err := plotter.NewHist(plotter.Values(values), 20)
	if err != nil {
		panic(err)
	}
	p.Add(hist)

	if err := p.Save(3*vg.Inch, 3*vg.Inch, title+".png"); err != nil {
		panic(err)
	}
}
