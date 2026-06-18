\ Output MMIO address.
const IO_OUT 65284

\ First vector: ASCII A B C D.
word a0 65
word a1 66
word a2 67
word a3 68

\ Second vector: add 1 to every lane.
word b0 1
word b1 1
word b2 1
word b3 1

reserve out 4

: main
  \ V0 = [A, B, C, D]
  vload V0 &a0
  \ V1 = [1, 1, 1, 1]
  vload V1 &b0
  \ V2 = V0 + V1 = [B, C, D, E]
  vadd V2 V0 V1
  \ out = V2
  vstore V2 &out

  \ Print four output words as characters.
  &out @ IO_OUT !
  &out 4 + @ IO_OUT !
  &out 8 + @ IO_OUT !
  &out 12 + @ IO_OUT !
  halt
;