\ Euler problem 4: largest palindrome product for 3-digit numbers (100..999)
\ NOTE: this is intentionally compute-heavy in tick-accurate model.

const IO_OUT 65284

word i 0
word j 0
word max 0
word prod 0
word n 0
word tmp 0
word rev 0
word digit 0
reserve digits 8
word dptr 0

: is_palindrome
  dup &n !
  &tmp !
  0 &rev !

pal_loop:
  &tmp @ jz pal_done
  &tmp @ 10 mod &digit !
  &rev @ 10 * &digit @ + &rev !
  &tmp @ 10 / &tmp !
  jmp pal_loop

pal_done:
  &rev @ &n @ eq
  ret
;

: print_u32
  dup jnz print_nonzero
  drop
  48 IO_OUT !
  10 IO_OUT !
  ret

print_nonzero:
  &tmp !
  &digits &dptr !

extract_loop:
  &tmp @ jz emit_start
  &tmp @ 10 mod 48 +
  &dptr @ !
  &dptr @ 4 + &dptr !
  &tmp @ 10 / &tmp !
  jmp extract_loop

emit_start:
emit_loop:
  &dptr @ 4 - &dptr !
  &dptr @ @ IO_OUT !
  &dptr @ &digits gt jnz emit_loop

  10 IO_OUT !
  ret
;

: main
  999 &i !

outer_loop:
  &i @ 99 gt jz done
  999 &j !

inner_loop:
  &j @ 99 gt jz outer_next

  &i @ &j @ * &prod !
  &prod @ &max @ gt jz skip_update

  &prod @ is_palindrome
  jz skip_update

  &prod @ &max !

skip_update:
  &j @ 1 - &j !
  jmp inner_loop

outer_next:
  &i @ 1 - &i !
  jmp outer_loop

done:
  &max @ print_u32
  halt
;