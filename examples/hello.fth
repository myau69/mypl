\ Output MMIO address. const does not allocate memory.
const IO_OUT 65284

\ ptr stores current address inside msg.
word ptr 0

\ msg is stored as 4-byte chars plus zero terminator.
cstr msg "Hello, World!\n"

: main
  \ ptr = &msg
  &msg &ptr !

print_loop:
  \ char = *ptr; if char == 0 -> done
  &ptr @ @ dup jz done

  \ write char to output port
  IO_OUT !

  \ ptr = ptr + 4 because one char is one 4-byte machine word
  &ptr @ 4 +
  &ptr !
  jmp print_loop

done:
  \ stop simulator
  halt
;