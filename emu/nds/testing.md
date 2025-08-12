# Testing


### Atem2069/armwrestler-fixed

test crashes from ldm currently

👍 Arm Alu
👍 Arm Ldr/Str
👍 Arm Lsm/Stm
👍 Thumb Alu
👍 Thumb Ldr/Str
👍 Thumb Lsm/Stm

❌ Arm v5TE
    👍 CLZ
    ❌ LDRD
    ❌ MRC
    👍 QADD
    👍 SMLABB
    👍 SMLABT
    👍 SMLATB
    👍 SMLATT

### RockPolish/rockwrestler

👍 Armv4
    👍 Condition Codes

❌ Armv5
    👍 CLZ
    👍 QADD, QSUB
    👍 QDADD, QDSUB
    👍 SMULxy
    👍 SMLAxy
    👍 SMULWy
    👍 SMLAWy
    👍 SMLALxy
    👍 BLX
    👍 PC SPEC
    ❌ LDM / STM

❌ Ds math
    ❌ Sqrt 32
    ❌ Sqrt 64
    ❌ Div 32/32
    ❌ Div 64/32
    👍 Div 64/64
