package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/spf13/cobra"
)

func readJSONFile(path string) (interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

type DiffResult struct {
	Key      string      `json:"key"`
	Type     string      `json:"type"`
	Filename string      `json:"filename,omitempty"`
	Value1   interface{} `json:"value1,omitempty"`
	Value2   interface{} `json:"value2,omitempty"`
}

type DiffOutput struct {
	IsDiff bool         `json:"isdiff"`
	Diff   []DiffResult `json:"diff"`
}

func diffJSON(a, b interface{}, file1, file2, prefix string) []DiffResult {
	return diffJSONWithOptions(a, b, file1, file2, prefix, false)
}

func diffJSONWithOptions(a, b interface{}, file1, file2, prefix string, ignoreArrayOrder bool) []DiffResult {
	var diffs []DiffResult

	// Handle different types
	switch aVal := a.(type) {
	case map[string]interface{}:
		if bMap, ok := b.(map[string]interface{}); ok {
			diffs = append(diffs, diffObjects(aVal, bMap, file1, file2, prefix, ignoreArrayOrder)...)
		} else {
			diffs = append(diffs, DiffResult{Key: prefix + "root", Type: "type_mismatch", Filename: file1, Value1: a, Value2: b})
		}
	case []interface{}:
		if bArray, ok := b.([]interface{}); ok {
			if ignoreArrayOrder {
				diffs = append(diffs, diffArraysIgnoreOrder(aVal, bArray, file1, file2, prefix, ignoreArrayOrder)...)
			} else {
				diffs = append(diffs, diffArrays(aVal, bArray, file1, file2, prefix, ignoreArrayOrder)...)
			}
		} else {
			diffs = append(diffs, DiffResult{Key: prefix + "root", Type: "type_mismatch", Filename: file1, Value1: a, Value2: b})
		}
	default:
		// For primitive types
		if !reflect.DeepEqual(a, b) {
			diffs = append(diffs, DiffResult{Key: prefix + "root", Type: "changed", Filename: file1, Value1: a, Value2: b})
		}
	}

	return diffs
}

func diffObjects(a, b map[string]interface{}, file1, file2, prefix string, ignoreArrayOrder bool) []DiffResult {
	var diffs []DiffResult
	for k, vA := range a {
		if vB, ok := b[k]; ok {
			if !reflect.DeepEqual(vA, vB) {
				// Check if both are simple values (not objects or arrays)
				if !isComplexType(vA) && !isComplexType(vB) {
					diffs = append(diffs, DiffResult{Key: prefix + k, Type: "changed", Filename: file1, Value1: vA, Value2: vB})
				} else {
					diffs = append(diffs, diffJSONWithOptions(vA, vB, file1, file2, prefix+k+".", ignoreArrayOrder)...)
				}
			}
		} else {
			diffs = append(diffs, DiffResult{Key: prefix + k, Type: "only_in_first", Filename: file1, Value1: vA})
		}
	}
	for k, vB := range b {
		if _, ok := a[k]; !ok {
			diffs = append(diffs, DiffResult{Key: prefix + k, Type: "only_in_second", Filename: file2, Value2: vB})
		}
	}
	return diffs
}

func isComplexType(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

func diffArrays(a, b []interface{}, file1, file2, prefix string, ignoreArrayOrder bool) []DiffResult {
	var diffs []DiffResult
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	for i := 0; i < maxLen; i++ {
		indexKey := fmt.Sprintf("%s[%d]", prefix, i)
		if i < len(a) && i < len(b) {
			// Both arrays have this index
			if !reflect.DeepEqual(a[i], b[i]) {
				diffs = append(diffs, diffJSONWithOptions(a[i], b[i], file1, file2, indexKey+".", ignoreArrayOrder)...)
			}
		} else if i < len(a) {
			// Only first array has this index
			diffs = append(diffs, DiffResult{Key: indexKey, Type: "only_in_first", Filename: file1, Value1: a[i]})
		} else {
			// Only second array has this index
			diffs = append(diffs, DiffResult{Key: indexKey, Type: "only_in_second", Filename: file2, Value2: b[i]})
		}
	}

	return diffs
}

func diffArraysIgnoreOrder(a, b []interface{}, file1, file2, prefix string, ignoreArrayOrder bool) []DiffResult {
	var diffs []DiffResult

	// Create copies to avoid modifying originals
	aCopy := make([]interface{}, len(a))
	bCopy := make([]interface{}, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	// Track which elements have been matched
	matchedA := make([]bool, len(aCopy))
	matchedB := make([]bool, len(bCopy))

	// Find matches
	for i, aItem := range aCopy {
		for j, bItem := range bCopy {
			if !matchedB[j] && reflect.DeepEqual(aItem, bItem) {
				matchedA[i] = true
				matchedB[j] = true
				break
			}
		}
	}

	// Report unmatched elements from first array
	for i, aItem := range aCopy {
		if !matchedA[i] {
			indexKey := fmt.Sprintf("%s[%d]", prefix, i)
			diffs = append(diffs, DiffResult{Key: indexKey, Type: "only_in_first", Filename: file1, Value1: aItem})
		}
	}

	// Report unmatched elements from second array
	for j, bItem := range bCopy {
		if !matchedB[j] {
			indexKey := fmt.Sprintf("%s[%d]", prefix, j)
			diffs = append(diffs, DiffResult{Key: indexKey, Type: "only_in_second", Filename: file2, Value2: bItem})
		}
	}

	return diffs
}

func main() {
	var file1, file2 string
	var ignoreArrayOrder bool

	var checkDiffCmd = &cobra.Command{
		Use:   "checkdiff",
		Short: "Check diff between two JSON files",
		Run: func(cmd *cobra.Command, args []string) {
			if file1 == "" || file2 == "" {
				fmt.Println("Both --f1 and --f2 flags are required.")
				os.Exit(1)
			}
			json1, err := readJSONFile(file1)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", file1, err)
				os.Exit(1)
			}
			json2, err := readJSONFile(file2)
			if err != nil {
				fmt.Printf("Error reading %s: %v\n", file2, err)
				os.Exit(1)
			}
			diffs := diffJSONWithOptions(json1, json2, file1, file2, "", ignoreArrayOrder)
			var output DiffOutput
			if len(diffs) > 0 {
				output = DiffOutput{
					IsDiff: true,
					Diff:   diffs,
				}
			} else {
				output = DiffOutput{
					IsDiff: false,
					Diff:   make([]DiffResult, 0),
				}
			}
			jsonOut, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling diff result: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOut))
		},
	}

	checkDiffCmd.Flags().StringVar(&file1, "f1", "", "First JSON file (required)")
	checkDiffCmd.Flags().StringVar(&file2, "f2", "", "Second JSON file (required)")
	checkDiffCmd.Flags().BoolVar(&ignoreArrayOrder, "ignore-order", false, "Ignore array element order (treat arrays as sets)")
	checkDiffCmd.MarkFlagRequired("f1")
	checkDiffCmd.MarkFlagRequired("f2")

	var rootCmd = &cobra.Command{Use: "diffjson"}
	rootCmd.AddCommand(checkDiffCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
