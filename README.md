## Config
The first line of a command file should contain comma separated key value pairs.  
Example: `n:20, q:1`
The fields you can configure are currently:

n - number of nodes (integer greater than 1, defaults to 10)
vv - value visibility delay (from routing/mock/interface.go), the time (ms) taken for a value to be visible in the network (integer, defaults to 0)
q - routing query delay, the time (ms) taken to receive a response from a routing query (integer, defaults to 0)
md - message delay, the time (ms) taken for a bitswap message to be delivered (testnet/virtual.go?), (integer, defaults to 0)