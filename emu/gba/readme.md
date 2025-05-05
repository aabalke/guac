
https://github.com/shonumi/Emu-Docs/tree/master/GameBoy%20Advance

https://medium.com/@michelheily/hello-gba-journey-of-making-an-emulator-part-1-8793000e8606

Need to Add Alignments to:
- STR,STRH,STM,LDM,LDRD,STRD,PUSH,POP (forced align)
- LDR,SWP (rotated read)
- PC/R15 (branch opcodes, or MOV/ALU/LDR with Rd=R15)
