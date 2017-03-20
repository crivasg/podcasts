# podcasts

A command line podcast client written in golang. 

## Usage

Usage of `podcasts`:<br><br>
  `-days=1`: Number of days back to download an episode<br>
  `-output=/tmp/output.sh`: Path of the output file<br>
  `-add=http://feed.thisamericanlife.org/talpodcast`: Add feed url to the list of podcasts<br>

Example:

```./podcasts -output=/tmp/podcast.sh -days=3```

The list of podcasts is stored in the home directory of the current user. The file name is `~/.podcasts/feeds.txt`
