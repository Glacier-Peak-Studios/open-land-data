#! /usr/bin/env python

import sys
from os import path
from pathlib import Path
import json
import csv
import pandas as pd

def csvToJSON(csvPath):
    jsonOut = csvPath.replace(Path(csvPath).suffix, ".json")
    csv_file = pd.DataFrame(pd.read_csv(
        csvPath, sep=",", header=0, index_col=False))
    outJS = csv_file.to_json(None, orient="records", date_format="epoch",
                             double_precision=10, force_ascii=True, date_unit="ms", default_handler=None)
    jsDat = json.loads(outJS)
    with open(jsonOut, 'w') as jsonFile:
        jsonFile.write(json.dumps(jsDat[0], indent=4))
    print("json saved to: " + path.abspath(jsonOut))

def jsonToCSV(jsonPath):
    csvOut = jsonPath.replace(Path(jsonPath).suffix, ".csv")
    with open(jsonPath) as data_file:
        data = json.load(data_file)
    df = pd.json_normalize(data)
    df.to_csv(csvOut, index=False)
    print("csv saved to: " + path.abspath(csvOut))

if len(sys.argv) == 1 or len(sys.argv) > 2:
    print("usage:")
    print("\tsourcemgr.py <file>")
    print("\nWhere <file> is the file to convert.")
    print("The script will automatically detect .csv or .json and convert appropriatly.")
    print("Note: this script will automaticaly overwrite output")
else:
    file = sys.argv[1]
    ext = Path(file).suffix
    if path.exists(file) and path.isfile(file):
        if ext == ".json":
            print("Converting to csv")
            jsonToCSV(sys.argv[1])
        elif ext == ".csv":
            print("Converting to json")
            csvToJSON(sys.argv[1])
        else:
            print("sourcemgr: Filetype not supported: " + ext)
    else:
        print("sourcemgr: File does not exist: " + path.abspath(file))
