// NanoKVM: Poly1305 for riscv64, a register-for-register translation of the
// toolchain's sum_loong64.s (Copyright 2025 The Go Authors, BSD licence, see
// the LICENSE file of the Go distribution). See server/goroot-overlay/overlay.sh.
//
// The loong64 code maps one to one: MULV and MULHVU are MUL and MULHU, and
// "SGTU a, b, d" (d = a > b) is "SLTU a, b, d" (d = b < a) here. The two
// 8-byte message loads may be misaligned; the C906 does that in hardware.
//
// Register map from sum_loong64.s:
//	R4 X5	state		R13 X14		R24 X19
//	R5 X6	msg		R14 X15		R25 X20
//	R6 X7	len		R15 X16		R26 X21
//	R7 X8	16		R16 X17		R27 X22
//	R8 X9	h0		R17 X18		R28 X23
//	R9 X10	h1
//	R10 X11	h2
//	R11 X12	r0
//	R12 X13	r1

//go:build gc && !purego

#include "textflag.h"

// func update(state *macState, msg []byte)
TEXT ·update(SB), NOSPLIT, $0-32
	MOV	state+0(FP), X5
	MOV	msg_base+8(FP), X6
	MOV	msg_len+16(FP), X7

	MOV	$0x10, X8

	MOV	(X5), X9	// h0
	MOV	8(X5), X10	// h1
	MOV	16(X5), X11	// h2
	MOV	24(X5), X12	// r0
	MOV	32(X5), X13	// r1

	BLT	X7, X8, bytes_between_0_and_15

loop:
	MOV	(X6), X15	// msg[0:8]
	MOV	8(X6), X17	// msg[8:16]
	ADD	X15, X9, X9	// h0 (x1 + y1 = z1', if z1' < x1 then z1' overflow)
	ADD	X17, X10, X22
	SLTU	X15, X9, X19	// h0.carry
	SLTU	X10, X22, X23
	ADD	X22, X19, X10	// h1
	SLTU	X22, X10, X19
	OR	X19, X23, X19	// h1.carry
	ADD	$0x01, X19, X19
	ADD	X11, X19, X11	// h2

	ADD	$16, X6, X6	// msg = msg[16:]

multiply:
	MUL	X9, X12, X15	// h0r0.lo
	MULHU	X9, X12, X16	// h0r0.hi
	MUL	X10, X12, X14	// h1r0.lo
	MULHU	X10, X12, X17	// h1r0.hi
	ADD	X14, X16, X16
	SLTU	X14, X16, X19
	ADD	X19, X17, X17
	MUL	X11, X12, X20
	ADD	X17, X20, X20
	MUL	X9, X13, X14	// h0r1.lo
	MULHU	X9, X13, X17	// h0r1.hi
	ADD	X14, X16, X16
	SLTU	X14, X16, X19
	ADD	X19, X17, X17
	MOV	X17, X9
	MUL	X11, X13, X21	// h2r1
	MUL	X10, X13, X14	// h1r1.lo
	MULHU	X10, X13, X17	// h1r1.hi
	ADD	X14, X20, X20
	ADD	X17, X21, X22
	SLTU	X14, X20, X19
	ADD	X22, X19, X21
	ADD	X9, X20, X20
	SLTU	X9, X20, X19
	ADD	X19, X21, X21
	AND	$3, X20, X11
	AND	$-4, X20, X18
	ADD	X18, X15, X9
	ADD	X21, X16, X22
	SLTU	X18, X9, X19
	SLTU	X21, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X19, X11, X11
	SLL	$62, X21, X22
	SRL	$2, X20, X23
	SRL	$2, X21, X21
	OR	X22, X23, X20
	ADD	X20, X9, X9
	ADD	X21, X10, X22
	SLTU	X20, X9, X19
	SLTU	X21, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X19, X11, X11

	SUB	$16, X7, X7
	BGE	X7, X8, loop

bytes_between_0_and_15:
	BEQZ	X7, done
	MOV	$1, X15
	MOV	$0, X16
	ADD	X7, X6, X6

flush_buffer:
	MOVBU	-1(X6), X20
	SRL	$56, X15, X19
	SLL	$8, X16, X23
	SLL	$8, X15, X15
	OR	X19, X23, X16
	XOR	X20, X15, X15
	SUB	$1, X7, X7
	SUB	$1, X6, X6
	BNEZ	X7, flush_buffer

	ADD	X15, X9, X9
	SLTU	X15, X9, X19
	ADD	X16, X10, X22
	SLTU	X16, X22, X23
	ADD	X22, X19, X10
	SLTU	X22, X10, X19
	OR	X19, X23, X19
	ADD	X11, X19, X11

	MOV	$16, X7
	JMP	multiply

done:
	MOV	X9, (X5)
	MOV	X10, 8(X5)
	MOV	X11, 16(X5)
	RET
