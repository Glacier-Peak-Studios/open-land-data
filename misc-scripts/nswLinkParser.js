"use strict";

const fs = require("fs");


// Linke examples:
// https://portal.spatial.nsw.gov.au/download/geopdf/25k/8823-2S+NADGEE.pdf
// https://portal.spatial.nsw.gov.au/download/geopdf/100k/7139+FORT+GREY.pdf

let base_url = "https://portal.spatial.nsw.gov.au/download/geopdf/";

let rawQuery = fs.readFileSync('nsw-query-8.json');
let query = JSON.parse(rawQuery);

let features = query.features;

for (let feature of features) {
	let attrs = feature.attributes;
	let tileid = attrs.tileid;
	let tilename = attrs.tilename;
	let scale = attrs.scale;
	let titledScale = (scale / 1000) + "k";
	let filename = (tileid + " " + tilename).replace(/\s+/g, '+');
	let url = base_url + titledScale + "/" + filename + ".pdf";
	console.log(url);
}

	

