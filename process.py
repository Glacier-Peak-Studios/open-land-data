import sys
from os import path
from pathlib import Path
import json

def swapCoords(coords):
  temp = coords[0]
  coords[0] = coords[1]
  coords[1] = temp
  # return [coords[1], coords[0]]

args = sys.argv[1:]
print(args)
try:
  args.index('-f') >= 0
  f_read = args[args.index('-f') + 1]
  filename = Path(f_read).stem
  filedir = path.dirname(f_read)
except:
  filename = "deer"
  filedir = "./generated/US/AR"
  f_read = filedir + '/' +filename + '.geojson'
  print("No file specified (-f), defaulting to: " + f_read)

try:
  args.index('-o') >= 0
  outdir = args[args.index('-o') + 1]
except:
  outdir = filedir

try:
  args.index('-l') >= 0
  newfname = filename + "-swapped"
except:
  newfname = filename



print("Opening file: " + f_read)
if (path.exists(f_read)):
  with open(f_read) as f:
    data = json.load(f)
else:
  sys.exit("File does not exist: " + f_read)


print("Swapping coordinates")

for feature in data['features']:
  if (len(feature['geometry']['coordinates']) == 2):
            swapCoords(feature['geometry']['coordinates'])
  else:
    for coordinates in feature['geometry']['coordinates']:
        for coordinateHolder in coordinates:
          if (len(coordinateHolder) == 2):
            swapCoords(coordinateHolder)
          else:
            for coordinate in coordinateHolder:
              swapCoords(coordinate)

print("Coordinates swapped.")

with open(outdir + '/' + newfname + '.geojson', 'w') as json_file:
  json.dump(data, json_file)
  print("New file is: " + json_file.name)

