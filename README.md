## Usage
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
* `node#->node <latency> <bandwidth>` - assigns a latency and bandwidth to the links between the nodes in node# and node.

## Config
The first line of a command file should contain comma separated key value pairs.  
Example: `node_count:20, query_delay:1`  
The fields you can configure are currently:  

* `node_count` - Number of nodes (integer greater than 1, defaults to 10).  
* `visibility_delay` - Value visibility delay. The time (ms) taken for a value to be visible in the network (integer, defaults to 0).  
* `query_delay` - Routing query delay, the time (ms) taken to receive a response from a routing query (integer, defaults to 0).  
* `block_size` - Block size in bytes (integer, defaults to `splitter.DefaultBlockSize`).
* `deadline` - Number of seconds it takes for a GetBlocks request to time out (float, defaults to 60).
* `bandwidth` - Specifies default bandwidth in megabits/sec for all links.  Can be changed using the `->` command (float, defaults to 100).
* `latency` - Specifies default latency in milliseconds for all links.  Can be changed using the `->` command (float, defaults to  0).