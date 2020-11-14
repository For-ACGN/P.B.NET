// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// void runtime·asmstdcallwithlasterror(void *c);
TEXT runtime·asmstdcallwithlasterror(SB),NOSPLIT,$0
	MOVL	fn+0(FP), BX

	// SetLastError(0).
	// MOVL	$0, 0x34(FS)

	// Copy args to the stack.
	MOVL	SP, BP
	MOVL	libcall_n(BX), CX	// words
	MOVL	CX, AX
	SALL	$2, AX
	SUBL	AX, SP			// room for args
	MOVL	SP, DI
	MOVL	libcall_args(BX), SI
	CLD
	REP; MOVSL

	// Call stdcall or cdecl function.
	// DI SI BP BX are preserved, SP is not
	CALL	libcall_fn(BX)
	MOVL	BP, SP

	// Return result.
	MOVL	fn+0(FP), BX
	MOVL	AX, libcall_r1(BX)
	MOVL	DX, libcall_r2(BX)

	// GetLastError().
	MOVL	0x34(FS), AX
	MOVL	AX, libcall_err(BX)

	RET
