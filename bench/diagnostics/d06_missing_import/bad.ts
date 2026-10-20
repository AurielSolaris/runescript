// D06: an import of a file that does not exist.
declare var process: any;
import helpers = require("./helpers");

process.stdout.write(helpers.double(2) + "\n");
