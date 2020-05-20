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
        props = json.loads(jsON["properties"].replace("\'", "\""))
        jsON['properties'] = props
        specStr = jsON["species"]
        if specStr is not None:
            specStr = specStr.replace("\', \'", "\", \"").replace("[\'", "[\"").replace("\']", "\"]").replace("\', \"", "\", \"").replace("\", \'", "\", \"")
            spec = json.loads(specStr)
            jsON['species'] = spec
        else:
            jsON['species'] = None
        
        jsonOut = sourcedir + srcType + "/" + country + "/" + subdiv + "/" + fname
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
    print("\tsourcemgr.py [mode] <sourcedir>")
    print("\nWhere <sourcdir> is the root folder of the .json sources.")
    print("If <sourcdier> is ommited, it will default to './land-sources'")
    print("\nMODES")
    print("\t-i\tConvert from csv to json")
    print("\t-o\tConvert from json to csv")
    print("\nNote: this script will automaticaly overwrite all files")
else:
    mode = sys.argv[1]
    if len(sys.argv) == 2:
        source_dir = "./land-sources"
    else:
        source_dir = sys.argv[2]
    source_dir = os.path.abspath(source_dir) + "/"
    if (mode == "-i"):
        print('Compiling sources from dir: ' + source_dir)
        compileJSONs(source_dir)
        jsonToCSV("combined-sources.json")
        os.remove("combined-sources.json")
    elif (mode == "-o"):
        csvToJSONs(source_dir, "combined-sources.csv")

