package ann

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sachaservan/vec"
)

const trainDatasetSuffix = "_train.csv"
const testDatasetSuffix = "_test.csv"
const neighborsDatasetSuffix = "_neighbors.csv"

func NewDatastream(fileName string) ([]*vec.Vec, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	data := make([]*vec.Vec, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		if len(line) > 1 {
			tokens := strings.Split(line, ",")
			valuesFloat := make([]float64, len(tokens))
			for j, v := range tokens {
				// note in particular the glove dataset uses float and not int values
				f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
				if err != nil {
					panic(err)
				}
				valuesFloat[j] = f
			}
			data = append(data, vec.NewVec(valuesFloat))
		}
		if err == io.EOF {
			break
		}
	}
	file.Close()
	return data, nil
}

func ReadDataset(datasetName string) ([]*vec.Vec, []*vec.Vec, [][]int, error) {
	trainDataset := datasetName + trainDatasetSuffix
	testDataset := datasetName + testDatasetSuffix
	neighborsDataset := datasetName + neighborsDatasetSuffix

	trainData, err := NewDatastream(trainDataset)
	if err != nil {
		return nil, nil, nil, err
	}

	testData, err := NewDatastream(testDataset)
	if err != nil {
		return nil, nil, nil, err
	}

	neighborIndices, err := NewDatastream(neighborsDataset)
	if err != nil {
		return nil, nil, nil, err
	}

	neighborIdxs := make([][]int, len(neighborIndices))
	// somewhat of a hack: convert vec.Vec
	for i, indexVector := range neighborIndices {
		indexes := make([]int, indexVector.Size())
		for j := range indexes {
			indexes[j] = int(indexVector.Coord(j))
		}
		neighborIdxs[i] = indexes
	}

	return trainData, testData, neighborIdxs, nil
}
