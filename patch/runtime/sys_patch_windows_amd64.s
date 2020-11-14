// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// maxargs should be divisible by 2, as Windows stack
// must be kept 16-byte aligned on syscall entry.
#define maxargs 16

// void runtime·asmstdcallwithlasterror(void *c);
TEXT runtime·asmstdcallwithlasterror(SB),NOSPLIT|NOFRAME,$0
	// asmcgocall will put first argument into CX.
	PUSHQ	CX			// save for later
	MOVQ	libcall_fn(CX), AX
	MOVQ	libcall_args(CX), SI
	MOVQ	libcall_n(CX), CX

	// SetLastError(0).
	// MOVQ	0x30(GS), DI
	// MOVL	$0, 0x68(DI)

	SUBQ	$(maxargs*8), SP	// room for args

	// Fast version, do not store args on the stack.
	CMPL	CX, $4
	JLE	loadregs

	// Check we have enough room for args.
	CMPL	CX, $maxargs
	JLE	2(PC)
	INT	$3			// not enough room -> crash

	// Copy args to the stack.
	MOVQ	SP, DI
	CLD
	REP; MOVSQ
	MOVQ	SP, SI

loadregs:
	// Load first 4 args into correspondent registers.
	MOVQ	0(SI), CX
	MOVQ	8(SI), DX
	MOVQ	16(SI), R8
	MOVQ	24(SI), R9
	// Floating point arguments are passed in the XMM
	// registers. Set them here in case any of the arguments
	// are floating point values. For details see
	//	https://msdn.microsoft.com/en-us/library/zthk2dkh.aspx
	MOVQ	CX, X0
	MOVQ	DX, X1
	MOVQ	R8, X2
	MOVQ	R9, X3

	// Call stdcall function.
	CALL	AX

	ADDQ	$(maxargs*8), SP

	// Return result.
	POPQ	CX
	MOVQ	AX, libcall_r1(CX)
	// Floating point return values are returned in XMM0. Setting r2 to this
	// value in case this call returned a floating point value. For details,
	// see https://docs.microsoft.com/en-us/cpp/build/x64-calling-convention
	MOVQ    X0, libcall_r2(CX)

	// GetLastError().
	MOVQ	0x30(GS), DI
	MOVL	0x68(DI), AX
	MOVQ	AX, libcall_err(CX)

	RET
