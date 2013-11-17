// +build !gccgo

// func bitlen(x Word) (n int)
TEXT ·bitlen(SB),4,$0
    MOVW    x+0(FP), R0
    CLZ     R0, R0
    MOVW    $32, R1
    SUB.S   R0, R1
    MOVW    R1, n+4(FP)
    RET
