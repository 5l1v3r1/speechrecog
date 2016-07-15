package ctc

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/autofunc/functest"
	"github.com/unixpickle/num-analysis/linalg"
)

const (
	testSymbolCount = 5
	testPrecision   = 1e-5

	benchLabelLen    = 50
	benchSeqLen      = 500
	benchSymbolCount = 30
)

var gradTestInputs = []*autofunc.Variable{
	&autofunc.Variable{Vector: []float64{-1.58522, -1.38379, -0.92827, -1.90226}},
	&autofunc.Variable{Vector: []float64{-2.87357, -2.75353, -1.11873, -0.59220}},
	&autofunc.Variable{Vector: []float64{-1.23140, -1.08975, -1.89920, -1.50451}},
	&autofunc.Variable{Vector: []float64{-1.44935, -1.51638, -1.59394, -1.07105}},
	&autofunc.Variable{Vector: []float64{-2.15367, -1.80056, -2.75221, -0.42320}},
}

var gradTestLabels = []int{2, 0, 1}

type logLikelihoodTestFunc struct{}

func (_ logLikelihoodTestFunc) Apply(in autofunc.Result) autofunc.Result {
	resVec := make([]autofunc.Result, len(gradTestInputs))
	for i, x := range gradTestInputs {
		resVec[i] = x
	}
	return LogLikelihood(resVec, gradTestLabels)
}

func (_ logLikelihoodTestFunc) ApplyR(rv autofunc.RVector, in autofunc.RResult) autofunc.RResult {
	resVec := make([]autofunc.RResult, len(gradTestInputs))
	for i, x := range gradTestInputs {
		resVec[i] = autofunc.NewRVariable(x, rv)
	}
	return LogLikelihoodR(resVec, gradTestLabels)
}

func TestLogLikelihoodOutputs(t *testing.T) {
	for i := 0; i < 10; i++ {
		labelLen := 5 + rand.Intn(5)
		seqLen := labelLen + rand.Intn(5)
		label := make([]int, labelLen)
		for i := range label {
			label[i] = rand.Intn(testSymbolCount)
		}
		seq, resSeq, rresSeq := createTestSequence(seqLen, testSymbolCount)
		expected := exactLikelihood(seq, label, -1)
		actual := math.Exp(LogLikelihood(resSeq, label).Output()[0])
		rActual := math.Exp(LogLikelihoodR(rresSeq, label).Output()[0])
		if math.Abs(actual-expected)/math.Abs(expected) > testPrecision {
			t.Errorf("LogLikelihood gave log(%e) but expected log(%e)",
				actual, expected)
		}
		if math.Abs(rActual-expected)/math.Abs(expected) > testPrecision {
			t.Errorf("LogLikelihoodR gave log(%e) but expected log(%e)",
				rActual, expected)
		}
	}
}

func TestLoglikelihoodGradients(t *testing.T) {
	test := functest.FuncTest{
		F:     logLikelihoodTestFunc{},
		Vars:  gradTestInputs,
		Input: gradTestInputs[0],
	}
	test.Run(t)
}

func TestLoglikelihoodRGradients(t *testing.T) {
	gradTestRVector := autofunc.RVector{}

	for _, in := range gradTestInputs {
		rVec := make(linalg.Vector, len(in.Vector))
		for i := range rVec {
			rVec[i] = rand.NormFloat64()
		}
		gradTestRVector[in] = rVec
	}

	test := functest.RFuncTest{
		F:     logLikelihoodTestFunc{},
		Vars:  gradTestInputs,
		Input: gradTestInputs[0],
		RV:    gradTestRVector,
	}
	test.Run(t)
}

func BenchmarkLogLikelihood(b *testing.B) {
	label := make([]int, benchLabelLen)
	for i := range label {
		label[i] = rand.Intn(testSymbolCount)
	}
	_, resSeq, _ := createTestSequence(benchSeqLen, benchSymbolCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogLikelihood(resSeq, label)
	}
}

func createTestSequence(seqLen, symCount int) (seq []linalg.Vector,
	res []autofunc.Result, rres []autofunc.RResult) {
	res = make([]autofunc.Result, seqLen)
	rres = make([]autofunc.RResult, seqLen)
	seq = make([]linalg.Vector, seqLen)
	for i := range seq {
		seq[i] = make(linalg.Vector, symCount+1)
		var probSum float64
		for j := range seq[i] {
			seq[i][j] = rand.Float64()
			probSum += seq[i][j]
		}
		for j := range seq[i] {
			seq[i][j] /= probSum
		}
		logVec := make(linalg.Vector, len(seq[i]))
		res[i] = &autofunc.Variable{
			Vector: logVec,
		}
		for j := range logVec {
			logVec[j] = math.Log(seq[i][j])
		}
		rres[i] = &autofunc.RVariable{
			Variable:   res[i].(*autofunc.Variable),
			ROutputVec: make(linalg.Vector, len(logVec)),
		}
	}
	return
}

func exactLikelihood(seq []linalg.Vector, label []int, lastSymbol int) float64 {
	if len(seq) == 0 {
		if len(label) == 0 {
			return 1
		} else {
			return 0
		}
	}

	next := seq[0]
	blank := len(next) - 1

	var res float64
	res += next[blank] * exactLikelihood(seq[1:], label, -1)
	if lastSymbol >= 0 {
		res += next[lastSymbol] * exactLikelihood(seq[1:], label, lastSymbol)
	}
	if len(label) > 0 {
		res += next[label[0]] * exactLikelihood(seq[1:], label[1:], label[0])
	}
	return res
}
