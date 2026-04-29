# EGM2008 File Reading Utility
1. Download the EGM2008 16-bit PGM file from NASA or somewhere.
1. Use the library in your code (you will need to pass the file location or use an environment variable
1. Use the command line utility to create one of the following:
* get the geoidal height for a single point
* get a list of heights from a grid specified by four corners and 2 spacings
* get a list of heights from a a list of positions specified in a CSV file

There is help in the CLI for command line usage. The package is fairly basic, and only exposes
- creating a new EGM2008Reader, which accesses the file
- extracting heights from points, lists of points or grids of points
- closing the file. This should probably be done automatically, but isn't.