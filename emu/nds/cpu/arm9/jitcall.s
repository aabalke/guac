#include "textflag.h"

TEXT ·jitRead32(SB), NOSPLIT, $0-4
    MOVL addr+0(FP), DI   // ABIInternal: arg already in DI
    CALL ·Read32(SB)
    MOVL AX, ret+8(FP)
    RET
