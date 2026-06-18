\ Euler-style prob1: largest palindrome product for two-digit numbers (10..99)
\ Output: decimal number + \n

\ Output MMIO address.
const IO_OUT 65284

\ Variables for loops, current product and palindrome check.
word i 0
word j 0
word max 0
word prod 0
word n 0
word tmp 0
word rev 0
word digit 0
\ Decimal output buffer.
reserve digits 8
word dptr 0



: is_palindrome
  \ input: n on stack
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
  \ input: value on stack
  \ Special case for zero.
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
  \ Start from highest two-digit number.
  99 &i !

outer_loop:
  &i @ 9 gt jz done
  99 &j !

inner_loop:
  &j @ 9 gt jz outer_next

  \ prod = i * j
  &i @ &j @ * &prod !
  \ Skip if prod <= max.
  &prod @ &max @ gt jz skip_update

  \ Update max only if prod is palindrome.
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