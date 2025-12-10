#include "funcdata.h"
#include "textflag.h"

TEXT ·jitcall(SB),NOSPLIT|NOFRAME,$0
        LEAQ argframe+0(FP), DI
        MOVQ 8(DX), AX
        JMP AX
