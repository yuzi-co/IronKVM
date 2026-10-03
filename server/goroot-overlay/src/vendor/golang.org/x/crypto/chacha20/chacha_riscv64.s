// NanoKVM: ChaCha20 for riscv64. See chacha_riscv64.go.

//go:build gc && !purego

#include "textflag.h"

// One block at a time in the integer registers. The sixteen state words stay
// in X5..X20 for the whole block, so the twenty rounds touch no memory. The
// four quarter rounds of a column (or diagonal) round are independent and run
// side by side, one step of each in turn.
//
// ADDW and th.srriw read only the low 32 bits of their inputs, so the bits
// that XOR leaves above bit 31 never reach a result, and nothing has to clear
// them. Only C906 and C910 cores run this; see useXTheadBb.
//
// Register use:
//	X5..X20	state words 0..15
//	X21	scratch
//	X22	dst
//	X23	src
//	X24	blocks left
//	X25	block counter
//	X26	double rounds left
//	X28	key
//	X29	nonce
//	X30	counter pointer

// th.srriw rd, rs1, imm (XTheadBb) rotates the low word right by imm. The Go
// assembler does not know it, so it is encoded here, which is why the
// register is given by number. Rotating left by n is rotating right by 32-n.
#define THROTL(n, r) WORD $(0x1400100b | ((32-(n))<<20) | ((r)<<15) | ((r)<<7))

// a += b; d ^= a, for four quarter rounds.
#define STEP4(a1, b1, a2, b2, a3, b3, a4, b4, d1, d2, d3, d4) \
	ADDW	b1, a1, a1; \
	ADDW	b2, a2, a2; \
	ADDW	b3, a3, a3; \
	ADDW	b4, a4, a4; \
	XOR	a1, d1, d1; \
	XOR	a2, d2, d2; \
	XOR	a3, d3, d3; \
	XOR	a4, d4, d4

// nbK and ndK are the register numbers of bK and dK, for THROTL.
#define THQR4(a1, b1, c1, d1, a2, b2, c2, d2, a3, b3, c3, d3, a4, b4, c4, d4, nb1, nd1, nb2, nd2, nb3, nd3, nb4, nd4) \
	STEP4(a1, b1, a2, b2, a3, b3, a4, b4, d1, d2, d3, d4); \
	THROTL(16, nd1); THROTL(16, nd2); THROTL(16, nd3); THROTL(16, nd4); \
	STEP4(c1, d1, c2, d2, c3, d3, c4, d4, b1, b2, b3, b4); \
	THROTL(12, nb1); THROTL(12, nb2); THROTL(12, nb3); THROTL(12, nb4); \
	STEP4(a1, b1, a2, b2, a3, b3, a4, b4, d1, d2, d3, d4); \
	THROTL(8, nd1); THROTL(8, nd2); THROTL(8, nd3); THROTL(8, nd4); \
	STEP4(c1, d1, c2, d2, c3, d3, c4, d4, b1, b2, b3, b4); \
	THROTL(7, nb1); THROTL(7, nb2); THROTL(7, nb3); THROTL(7, nb4)

#define THDOUBLEROUND \
	THQR4(X5, X9, X13, X17, X6, X10, X14, X18, X7, X11, X15, X19, X8, X12, X16, X20, 9, 17, 10, 18, 11, 19, 12, 20); \
	THQR4(X5, X10, X15, X20, X6, X11, X16, X17, X7, X12, X13, X18, X8, X9, X14, X19, 10, 20, 11, 17, 12, 18, 9, 19)

#define C0 0x61707865
#define C1 0x3320646e
#define C2 0x79622d32
#define C3 0x6b206574

#define LOADSTATE \
	MOV	$C0, X5; \
	MOV	$C1, X6; \
	MOV	$C2, X7; \
	MOV	$C3, X8; \
	MOVWU	0(X28), X9; \
	MOVWU	4(X28), X10; \
	MOVWU	8(X28), X11; \
	MOVWU	12(X28), X12; \
	MOVWU	16(X28), X13; \
	MOVWU	20(X28), X14; \
	MOVWU	24(X28), X15; \
	MOVWU	28(X28), X16; \
	MOV	X25, X17; \
	MOVWU	0(X29), X18; \
	MOVWU	4(X29), X19; \
	MOVWU	8(X29), X20; \
	MOV	$10, X26

#define ADDK(k, s) MOV $(k), X21; ADDW X21, s, s
#define ADDM(off, p, s) MOVWU off(p), X21; ADDW X21, s, s

#define FEEDFORWARD \
	ADDK(C0, X5); \
	ADDK(C1, X6); \
	ADDK(C2, X7); \
	ADDK(C3, X8); \
	ADDM(0, X28, X9); \
	ADDM(4, X28, X10); \
	ADDM(8, X28, X11); \
	ADDM(12, X28, X12); \
	ADDM(16, X28, X13); \
	ADDM(20, X28, X14); \
	ADDM(24, X28, X15); \
	ADDM(28, X28, X16); \
	ADDW	X25, X17, X17; \
	ADDM(0, X29, X18); \
	ADDM(4, X29, X19); \
	ADDM(8, X29, X20)

// XW xors one key stream word into the input, little-endian like the target.
// The loads and stores may be misaligned: the TLS record layer seals in place
// five bytes into its buffer. The C906 does misaligned accesses in hardware.
#define XW(off, s) MOVWU off(X23), X21; XOR s, X21, X21; MOVW X21, off(X22)

#define XORBLOCK \
	XW(0, X5); \
	XW(4, X6); \
	XW(8, X7); \
	XW(12, X8); \
	XW(16, X9); \
	XW(20, X10); \
	XW(24, X11); \
	XW(28, X12); \
	XW(32, X13); \
	XW(36, X14); \
	XW(40, X15); \
	XW(44, X16); \
	XW(48, X17); \
	XW(52, X18); \
	XW(56, X19); \
	XW(60, X20); \
	ADD	$64, X23; \
	ADD	$64, X22; \
	ADDW	$1, X25; \
	ADD	$-1, X24

// The caller passes a whole number of blocks (a multiple of bufSize) and has
// already checked that the counter does not wrap.
#define PROLOGUE \
	MOV	dst_base+0(FP), X22; \
	MOV	src_base+24(FP), X23; \
	MOV	src_len+32(FP), X24; \
	MOV	key+48(FP), X28; \
	MOV	nonce+56(FP), X29; \
	MOV	counter+64(FP), X30; \
	MOVWU	(X30), X25; \
	SRL	$6, X24

// func xorKeyStreamTH(dst, src []byte, key *[8]uint32, nonce *[3]uint32, counter *uint32)
TEXT ·xorKeyStreamTH(SB), NOSPLIT, $0-72
	PROLOGUE
	BEQZ	X24, thdone

thblock:
	LOADSTATE

thround:
	THDOUBLEROUND
	ADD	$-1, X26
	BNEZ	X26, thround

	FEEDFORWARD
	XORBLOCK
	BNEZ	X24, thblock
	MOVW	X25, (X30)

thdone:
	RET
