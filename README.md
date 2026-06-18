# mypl

# Лабораторная работа №4: язык, транслятор, модель процессора

ФИО: Коршун Артём Сергеевич 
Группа: P3219

Вариант:

```text
forth | stack | neum | hw | tick | binary | trap | mem | cstr | prob1 | vector
```

Проект реализует полный цикл:

```text
source .fth -> forthc -> binary image FSN4 -> forthsim -> output + trace
```

То есть здесь есть свой Forth-подобный язык, транслятор в бинарный машинный код, формат `.bin`, модель стекового процессора, tick-level трассировка, trap-ввод и vector extension.

## Содержание

- [Язык программирования](#язык-программирования)
- [Организация памяти](#организация-памяти)
- [Система команд](#система-команд)
- [Транслятор](#транслятор)
- [Модель процессора](#модель-процессора)
- [Trap и ввод](#trap-и-ввод)
- [Vector extension](#vector-extension)
- [Тестирование](#тестирование)

## Язык программирования

Язык стековый, в Forth/RPN стиле: сначала operands, потом operation.

```forth
2 3 +
```

Смысл:

```text
push 2
push 3
add
```

### Синтаксис

```text
program      ::= { declaration | definition }

declaration  ::= "const" name literal
               | "word" name literal
               | "reserve" name literal
               | "cstr" name string

definition   ::= ":" name { item } ";"
item         ::= literal | builtin | name | label | branch | call | execution-token | vector-op

label        ::= name ":"
branch       ::= ("jmp" | "jz" | "jnz" | "jltz" | "jgtz" | "jgez" | "jlez") name
call         ::= "call" name
execution-token ::= "'" name
```

### Семантика

| Конструкция | Смысл |
|---|---|
| `: main ... ;` | procedure definition |
| `label:` | label текущего code address |
| `jmp label` | безусловный переход |
| `jz label` | снять condition со stack, перейти если `0` |
| `call word` / `ret` | обычный вызов procedure |
| `' word` / `execute` | execution token: address слова как значение |
| `word x 0` | 4-byte variable в data segment |
| `reserve buf 8` | выделить 8 machine words |
| `cstr msg "Hi\n"` | null-terminated string, один rune = один int32 word |
| `@` / `!` | load/store по memory address |

Пример:

```forth
const IO_OUT 65284

: main
  72 IO_OUT !
  73 IO_OUT !
  halt
;
```

Эта программа выводит `HI`: `IO_OUT` - memory-mapped output address.

## Организация памяти

Архитектура варианта - `neum`: code, data и MMIO находятся в едином адресном пространстве.

```text
0x0010        InputHandlerSlotAddr
0x0200        CodeBase
0x4000        DataBase
0xFF00        IOInDataAddr
0xFF04        IOOutDataAddr
0xFF08        IOInReadyAddr
```

Compiled program хранится как `Image`:

```go
type Image struct {
	Version          uint16
	MemorySize       uint32
	CodeBase         uint32
	DataBase         uint32
	EntryPoint       uint32
	InputHandlerAddr uint32
	Code             []byte
	Data             []byte
}
```

`Image` - это не config, а результат компиляции: куда загрузить code/data, с какого address стартовать и где находится `on_input`.

Бинарный файл `FSN4`:

```text
header + code bytes + data bytes
```

Header содержит magic `FSN4`, version, memory layout, entry point, input handler address и размеры сегментов. Числа кодируются little-endian.

## Система команд

### Stack/ALU

| Команда | Stack effect | Назначение |
|---|---|---|
| `push imm` | `-- imm` | положить immediate |
| `dup` | `a -- a a` | дублировать top |
| `drop` | `a --` | снять top |
| `swap` | `a b -- b a` | поменять местами |
| `over` | `a b -- a b a` | скопировать второй элемент |
| `+ - * / mod` | `a b -- r` | arithmetic |
| `eq lt gt` | `a b -- 0/1` | comparison |
| `and or xor shl shr inv` | bit ops | bit operations |

### Memory/control/trap

| Команда | Назначение |
|---|---|
| `@` | read int32 from memory |
| `!` | write int32 to memory |
| `jmp/jz/jnz/jltz/jgtz/jgez/jlez` | branches |
| `call/ret` | procedure call/return |
| `execute` | indirect call по address со stack |
| `ei/di/iret` | interrupt control |
| `halt` | останов simulator-а |

Instruction sizes:

```text
1 byte  opcode only
5 bytes opcode + int32 immediate/address
6 bytes vector load/store: opcode + register + int32 address
4 bytes vector arithmetic: opcode + dst + a + b
```

## Транслятор

CLI:

```bash
go run ./cmd/forthc -src examples/hello.fth -out build/hello.bin -list build/hello.lst
```

Pipeline:

```text
read source
tokenize
compile declarations/procedures/labels
resolve fixups
build Image
write FSN4 binary
write listing
```

Главные файлы:

```text
internal/compiler/tokenizer.go
internal/compiler/compiler.go
internal/binaryfmt/format.go
cmd/forthc/main.go
```

Listing показывает address, hex bytes and mnemonic. Это удобно для защиты: видно, какой machine code породил compiler.

## Модель процессора

CLI:

```bash
go run ./cmd/forthsim -bin build/hello.bin -config configs/hello.json -trace build/hello.trace.log
```

CPU state:

```go
mem []byte
pc  uint32
ds  []int32
rs  []uint32
vr  [4][4]int32
```

Смысл:

| Поле | Назначение |
|---|---|
| `mem` | единая memory |
| `pc` | address текущей instruction |
| `ds` | data stack |
| `rs` | return stack |
| `vr` | vector registers `V0..V3` |

Control unit hardwired: исполнение сделано через `switch op` в `internal/sim/cpu.go`, без микрокода.

Execution loop:

```text
service interrupt if needed
fetch/decode instruction
execute instruction
advance ticks
append trace entry
stop on halt
```

Trace пишет ticks, pc, instruction text, data stack, return stack, IRQ state and event.

## Trap и ввод

Config задает input events:

```json
{
  "max_ticks": 200000,
  "events": [
    { "tick": 5, "token": "H" },
    { "tick": 12, "token": "i" }
  ]
}
```

Flow:

```text
tick event -> inputQueue -> pendingIRQ -> on_input -> iret -> old pc
```

`on_input` - обычная procedure по коду, но специальная по имени: compiler записывает ее address в `Image.InputHandlerAddr`, а simulator использует этот address для входа в interrupt handler.

## Vector

Векторное расширение добавляет 4 vector registers по 4 lane

Команды:

```forth
vload  V0 &a
vload  V1 &b
vadd   V2 V0 V1
vstore V2 &out
```

Смысл: одна vector instruction обрабатывает 4 `int32` values lane-wise. Vector arithmetic не использует data stack.

## Тестирование

Быстрый старт:

```bash
go test ./...
make run-hello
make run-hello-username
make run-cat
make run-prob1
make run-vector
```

Generated artifacts:

```text
build/*.bin        binary images
build/*.lst        readable listings
build/*.trace.log  tick-level traces
```

Golden/e2e cases:

```text
golden/*.yml       source + config + expected output/listing hash/trace fragments
tests/golden_test.go
```

`build/` - generated artifacts.
