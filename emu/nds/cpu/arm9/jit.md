# Jit Instructions

## Done
Swp
Mul
Clz
QAlu
ALU
HALF

-- problem
SDT

## Todo
BLOCK
PSR
COPRO


## Not now
BLX
PLD
EXCEPTION (SWI UNDEF BKPT)
B
BX

Note: at final branch of block, it does not complete branching inst - just stops at it 
This may be optimized at some point

# Jit Compiling

Both the arm7 and arm9 cpus use jits (just in time compilers) to optimize performance.
Blocks of instructions are jitted when reaching frequency thresholds, and are evicted from the cache
using a LRU method (Least Recently Used). This method allows for a 2-5x speed increase with minimal memory usage.

Gojit is used as a base jit implimentation; however, many changes had to be made to get it to work with the modern
Go ABI (internal application binary interface). Gojit was last updated a decade ago - modern versions of golang use
registers and the stack for function parameter and return argument passing - previous version used only the stack. 
