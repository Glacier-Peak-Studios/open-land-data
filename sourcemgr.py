#! /usr/bin/env python

import sys
import os
from os import path
from pathlib import Path
import json
import csv
import pandas as pd
import glob


def csvToJSONs(sourcedir: str, csvPath: str):
    # jsonOut = csvPath.replace(Path(csvPath).suffix, ".json")
    csv_file = pd.DataFrame(pd.read_csv(
        csvPath, sep=",", header=0, index_col=False))
    outJS = csv_file.to_json(None, orient="records", date_format="epoch",
                             double_precision=10, force_ascii=True, date_unit="ms", default_handler=None)
    jsDat = json.loads(outJS)
    writeJSONs(sourcedir, jsDat)


def writeJSONs(sourcedir: str, jsonArr):
    for jsON in jsonArr:
        srcType = jsON.pop("type")
        country = jsON.pop("country")
        subdiv = jsON.pop("subdivision")
        fname = jsON.pop("filename")
        propStr = jsON["properties"]
        if propStr is not None:
            props = json.loads(propStr.replace("\'", "\""))
            jsON['properties'] = props
        else:
            jsON['properties'] = None
        specStr = jsON["species"]
        if specStr is not None:
            specStr = specStr.replace("\', \'", "\", \"").replace("[\'", "[\"").replace(
                "\']", "\"]").replace("\', \"", "\", \"").replace("\", \'", "\", \"")
            spec = json.loads(specStr)
            jsON['species'] = spec
        else:
            jsON['species'] = None
        dirPath = sourcedir + srcType + "/" + country + "/" + subdiv
        os.makedirs(dirPath, exist_ok=True)
        jsonOut = dirPath + "/" + fname
        with open(jsonOut, 'w') as jsonFile:
            jsonFile.write(json.dumps(jsON, indent=4))
        print("json saved to: " + path.abspath(jsonOut))


def jsonToCSV(jsonPath):
    csvOut = jsonPath.replace(Path(jsonPath).suffix, ".csv")
    with open(jsonPath) as data_file:
        data = json.load(data_file)
    df = pd.json_normalize(data, max_level=0)
    df.to_csv(csvOut, index=False)
    print("csv saved to: " + path.abspath(csvOut))


def compileJSONs(sourcedir: str):
    jsonList = []
    for filename in glob.iglob(source_dir + '**/*.json', recursive=True):
        with open(filename) as data_file:
            data = json.load(data_file)
            data["filename"] = path.basename(filename)
            subdiv = path.dirname(filename)
            data["subdivision"] = path.basename(subdiv)
            country = path.dirname(subdiv)
            data["country"] = path.basename(country)
            datType = path.dirname(country)
            data["type"] = path.basename(datType)
            jsonList.append(data)
    with open("combined-sources.json", 'w') as jsonFile:
        jsonFile.write(json.dumps(jsonList, indent=4))
        # print("json saved to: " + path.abspath(jsonOut))


if len(sys.argv) < 2 or len(sys.argv) > 3:
    print("usage:")
    print("\tsourcemgr.py [mode] <source>")
    print("\nWhere <source> is the root folder of the .json sources or the .csv file.")
    print("If <source> is ommited, it will default to './land-sources' for -c and './combined-sources.csv' for -j")
    print("\nMODES")
    print("\t-j\tConvert from csv to json")
    print("\t-c\tConvert from json to csv")
    print("\nNote: this script will automaticaly overwrite all files")
else:
    mode = sys.argv[1]
    source_dir = "./land-sources"
    source_dir = os.path.abspath(source_dir) + "/"
    if (mode == "-c"):
        if len(sys.argv) > 2:
            source_dir = sys.argv[2]
        print('Compiling sources from dir: ' + source_dir)
        compileJSONs(source_dir)
        jsonToCSV("combined-sources.json")
        os.remove("combined-sources.json")
    elif (mode == "-j"):
        sourceFile = "combined-sources.csv"
        if len(sys.argv) > 2:
            sourceFile = sys.argv[2]
        print('Parsing sources from: ' + sourceFile)
        csvToJSONs(source_dir, sourceFile)
