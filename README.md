## Usages

##### "I want to run a single workload and look at some basic stats for it."
1.  Make a [command file](#workloads) and save it or pick one from the `samples` folder.
2.  `cd` into your bssim folder and run `./bssim -wl <path-to-workload>`
3.  Check your stdout for some basic stats.

##### "I want to see graphs for the workload I just ran."
1.  `cd bssim/data`
2.  Edit `config.ini` to your liking.
3.  `python grapher.py`
4.  By default, graphs are displayed and saved to`data/graphs.pdf`

##### "I want to run a workload with a bunch of different latencies and bandwidths and see graphs for it."
1.  Edit `bssim/data/config.ini` to your liking.  The `latencies` and `bandwidths` will be what the workload is run with and what the graphs will show.
2.  Run `./bssim/scripts/latbw.sh <path-to-workload>`.

## Commands
Commands work similarly to [dhtHell](https://github.com/whyrusleeping/dhtHell).  
Commands are of the syntax `node# command arg`, where `node#` can be a single node number or a range in the form `[#-#]`.

Possible commands as of now are:  

* `put` - adds file where arg is the file path  
* `get` - gets file where arg is the file path  
* `putb` - adds block where arg is the contents of the block  
* `getb` - gets block where arg is the contents of the block  
* `leave` - causes nodes to leave network where arg is the number of seconds until the node leaves

There's also a few special commands:
* `create_dummy_files <# of files> <file size>`  - creates a specified number of files in the samples directory with names dummy(n) and then deletes them when the script finishes (see `samples/lotsofiles` for an example).  
* `node#->node <latency> <bandwidth>` - assigns a latency and bandwidth to the links between the nodes in node# and node (floats, ms and megabits per second).

## Workloads<a name="workloads"></a>
The first line of a workload file should contain comma separated key value pairs.  
Example: `node_count:20, query_delay:1`  
The fields you can configure are currently:  

* `node_count` - Number of nodes (integer greater than 1, defaults to 10).  
* `visibility_delay` - Value visibility delay. The time (ms) taken for a value to be visible in the network (integer, defaults to 0).  
* `query_delay` - Routing query delay, the time (ms) taken to receive a response from a routing query (integer, defaults to 0).  
* `block_size` - Block size in bytes (integer, defaults to `splitter.DefaultBlockSize`).
* `deadline` - Number of seconds it takes for a GetBlocks request to time out (float, defaults to 60).
* `bandwidth` - Specifies default bandwidth in megabits/sec for all links.  Can be changed using the `->` command (float, defaults to 100).
* `latency` - Specifies default latency in milliseconds for all links.  Can be changed using the `->` command (float, defaults to  0).

## Prometheus
Prometheus metrics are pushed to [localhost:8080/metrics](localhost:8080/metrics).
Current metrics collected are:
* `file_times_ms`
* `block_times_ms`
* `dup_blocks_count`