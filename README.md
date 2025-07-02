# diffjson

A powerful Go CLI tool to compare two JSON files and output their differences in JSON format. Supports both JSON objects and JSON arrays with detailed change tracking.

## Features

- ✅ **JSON Objects**: Compare key-value pairs in JSON objects
- ✅ **JSON Arrays**: Compare arrays element by element with index tracking
- ✅ **Nested Structures**: Recursively compare nested objects and arrays
- ✅ **Type Validation**: Detect type mismatches between files
- ✅ **Detailed Output**: Comprehensive difference reporting in JSON format


### Run directly
```sh
go run main.go checkdiff --f1 "./file1.json" --f2 "./file2.json"
```

## Usage

### Compare two JSON files

```sh
./diff-tool.exe checkdiff --f1 "./testfile/test1.json" --f2 "./testfile/test2.json"
```

### Compare with array order ignored

```sh
./diff-tool.exe checkdiff --f1 "./file1.json" --f2 "./file2.json" --ignore-order
```

**Parameters:**
- `--f1` : Path to the first JSON file (required)
- `--f2` : Path to the second JSON file (required)  
- `--ignore-order` : Ignore array element order (treat arrays as sets) (optional)

## Examples

### JSON Objects Comparison

**Input files:**
```json
// test1_old.json
{
  "name": "Alice",
  "age": 30,
  "city": "New York"
}

// test2_old.json
{
  "name": "Alice", 
  "age": 31,
  "country": "USA"
}
```

**Output:**
```json
{
  "isdiff": true,
  "diff": [
    {
      "key": "age",
      "type": "changed",
      "filename": "./test1_old.json",
      "value1": 30,
      "value2": 31
    },
    {
      "key": "city",
      "type": "only_in_first",
      "filename": "./test1_old.json",
      "value1": "New York"
    },
    {
      "key": "country",
      "type": "only_in_second",
      "filename": "./test2_old.json",
      "value2": "USA"
    }
  ]
}
```

### JSON Arrays Comparison

**Input files:**
```json
// test_simple1.json
[
  {"name": "John", "age": 30},
  {"name": "Jane", "age": 25}
]

// test_simple2.json
[
  {"name": "John", "age": 31},
  {"name": "Bob", "age": 25}
]
```

**Output:**
```json
{
  "isdiff": true,
  "diff": [
    {
      "key": "[0].age",
      "type": "changed",
      "filename": "./test_simple1.json",
      "value1": 30,
      "value2": 31
    },
    {
      "key": "[1].name",
      "type": "changed",
      "filename": "./test_simple1.json",
      "value1": "Jane",
      "value2": "Bob"
    }
  ]
}
```

### Array Order Comparison

**Without `--ignore-order` (default behavior):**
```json
// test_seq1.json
[
  {"aa": "1", "bb": "1"},
  {"cc": "2", "dd": "2"}
]

// test_seq2.json  
[
  {"cc": "2", "dd": "2"},
  {"aa": "1", "bb": "1"}
]
```

**Output:**
```json
{
  "isdiff": true,
  "diff": [
    {
      "key": "[0].aa",
      "type": "only_in_first",
      "filename": "./test_seq1.json",
      "value1": "1"
    },
    // ... more differences showing position-based comparison
  ]
}
```

**With `--ignore-order` flag:**
```sh
./diff-tool.exe checkdiff --f1 "./test_seq1.json" --f2 "./test_seq2.json" --ignore-order
```

**Output:**
```json
{
  "isdiff": false,
  "diff": []
}
```

The `--ignore-order` flag treats arrays as sets, matching elements regardless of their position in the array.

### No Differences

When files are identical:
```json
{
  "isdiff": false,
  "diff": []
}
```

## Difference Types

- **`changed`**: Field exists in both files but with different values
- **`only_in_first`**: Field exists only in the first file
- **`only_in_second`**: Field exists only in the second file
- **`type_mismatch`**: Field exists in both files but with different JSON types

## Array Indexing

For JSON arrays, differences are reported using bracket notation:
- `[0]`, `[1]`, `[2]` - Array element indices
- `[0].fieldname` - Field within array element
- `[0].nested[1].field` - Deeply nested structures

## Command Help

```sh
./diff-tool.exe checkdiff --help
```

View all available options and usage information.

