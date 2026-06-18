const IO_IN_DATA 65280
const IO_OUT     65284

word ptr 0
word done 0
word name_len 0
reserve name_buf 16

cstr prompt "What is your name?\n"
cstr hello "Hello, "
cstr suffix "!\n"

: emit_cstr
  &ptr !

emit_loop:
  &ptr @ @ dup jz emit_done
  IO_OUT !
  &ptr @ 4 + &ptr !
  jmp emit_loop

emit_done:
  drop
  ret
;

: on_input
  IO_IN_DATA @ dup
  10 - jz input_done

  &name_buf &name_len @ 4 * + !
  &name_len @ 1 + &name_len !
  iret

input_done:
  drop
  0 &name_buf &name_len @ 4 * + !
  1 &done !
  iret
;

: main
  &prompt emit_cstr
  ei

wait_loop:
  &done @ jz wait_loop

  &hello emit_cstr
  &name_buf emit_cstr
  &suffix emit_cstr
  halt
;
