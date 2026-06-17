\ Input and output MMIO addresses.
const IO_IN_DATA 65280
const IO_OUT     65284

\ Main loop runs while running != 0.
word running 1

\ on_input особый: compiler записывает его address как InputHandlerAddr.
: on_input
  \ Read one queued input token.
  IO_IN_DATA @ dup
  \ Echo it to output.
  IO_OUT !

  \ Newline (ASCII 10) stops the program.
  10 - jz stop
  iret

stop:
  \ running = 0
  0 &running !
  \ Return from interrupt to interrupted PC.
  iret
;

: main
wait_loop:
  \ Busy-wait until interrupt handler clears running.
  &running @ jnz wait_loop
  halt
;