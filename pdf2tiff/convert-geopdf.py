#!/usr/bin/env python3
#based on http://lists.osgeo.org/pipermail/gdal-dev/2013-January/035269.html
from osgeo import gdal, osr
import os
import sys
from optparse import OptionParser
import tempfile
from shapely import wkt
from shapely.geometry import mapping
from fiona import collection, crs
import logging
import shutil

def convert_geopdf(filename, dpi, layers_to_remove, target_srs, resample_alg):
    #read metadata
    ds = gdal.Open(filename)
    if not ds:
        raise Exception("failed to open file " + filename)

    cutline_dir, cutline_path = get_cutline_shp(ds)
    logging.debug("cutline path: " + cutline_path)
    if dpi is None:
        dpi = ds.GetMetadataItem("DPI")
        logging.debug("dpi: " + str(dpi))
        if dpi is None:
            raise Exception("DPI not specified, and not in metadata")

    ds = None

    layers_to_remove = ['"' + layer + '"' for layer in layers_to_remove]
    layers_string = ",".join(layers_to_remove)
    logging.info("removing layers: " + layers_string)
    srs_option = ''
    if target_srs is not None:
        srs_option = '-t_srs ' + target_srs

    command = ('gdalwarp ' +
          ' -co "TILED=YES" -co "TFW=YES" %s %s.tif ' % (filename, filename) +
          srs_option +
        #   ' -crop_to_cutline -cutline ' + cutline_path +
          ' -r ' + resample_alg +
          ' -overwrite ' + 
          ' -dstnodata 255' +
          ' --config GDAL_PDF_LAYERS_OFF ' + layers_string +
          ' --config GDAL_PDF_DPI ' + dpi)
    logging.debug(command)

    os.system(command)
    shutil.rmtree(cutline_dir)


def get_cutline_shp(ds):
    neatline_wkt = ds.GetMetadataItem("NEATLINE")

    ds_srs = osr.SpatialReference()
    ds_srs.ImportFromWkt(ds.GetProjection())
    ds_srs_proj = ds_srs.ExportToProj4()
    ds_crs = crs.from_string(ds_srs_proj)
    neatline = wkt.loads(neatline_wkt)
    schema = { 'geometry': 'Polygon', 'properties': {} }
    cutline_dir = tempfile.mkdtemp()
    cutline_path = os.path.join(cutline_dir, "cutline.shp")
    with collection(cutline_path, "w", driver="ESRI Shapefile", schema=schema, crs=ds_crs) as output:
        output.write({
                'geometry': mapping(neatline),
                'properties' : {}
        })

    return cutline_dir, cutline_path


if __name__=='__main__':    
    usage = "usage: %prog infile.pdf infile2.pdf..."
    parser = OptionParser(usage=usage,
        description="Convert a usgs geoPDF map to tiff")
    parser.add_option("-d", "--debug", action="store_true", dest="debug",
                      help="Turn on debug logging")
    parser.add_option("-q", "--quiet", action="store_true", dest="quiet",
                      help="turn off all logging")
    parser.add_option("-D", "--dpi", action="store", type="string", dest="dpi",
        default=None)
    parser.add_option("-t", "--t_srs", action="store", type="string", 
        dest="target_srs", default=None)
    parser.add_option("-r", "--remove", action="store", type="string", 
        dest="remove_layers",
        default="Map_Collar,Map_Frame.Projection_and_Grids")
    parser.add_option("-R", "--resample", action="store", type="string",
        dest="resample", default="near")
    (options, args) = parser.parse_args()

    logging.basicConfig(level=logging.DEBUG if options.debug else
    (logging.ERROR if options.quiet else logging.INFO))

    layers_to_remove = options.remove_layers.split(',')

    for filename in args:
        convert_geopdf(filename, options.dpi, layers_to_remove,
            options.target_srs, options.resample)
