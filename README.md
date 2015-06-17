## Usage
Commands work similarly to [dhtHell](https://github.com/whyrusleeping/dhtHell).  
Commands are of the syntax `node# command arg`, where `node#` can be a single node number or a range in the form `[#-#]`.

Possible commands as of now are:  

* `put` - adds file where arg is the file path  
* `get` - gets file where arg is the file path  
* `putb` - adds block where arg is the contents of the block  
* `getb` - gets block where arg is the contents of the block  

## Config
The first line of a command file should contain comma separated key value pairs.  
Example: `node_count:20, query_delay:1`  
The fields you can configure are currently:  

* `node_count` - Number of nodes (integer greater than 1, defaults to 10).  
* `visibility_delay` - Value visibility delay. The time (ms) taken for a value to be visible in the network (integer, defaults to 0).  
* `query_delay` - Routing query delay, the time (ms) taken to receive a response from a routing query (integer, defaults to 0).  
* `block_size` - Block size in bytes (integer, defaults to `splitter.DefaultBlockSize`).