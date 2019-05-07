package lua

import (
    "fmt"
    "github.com/Azure/golua/lua/code"
)

// rk returns the i'th stack value or the i'th
// constant if 'i' is a constant index.
func (ls *thread) rk(fn *Func, i int) *Value {
    if code.IsKst(i) {
		k := fn.kst(code.ToKst(i)).(Value)
		return &k
    }
    return &(ls.stack[ls.calls.base+i])
}

func execute(ls *thread) {
    ls.calls.flag |= fresh
    ci := ls.calls
frame:
    var (
        fn = ci.fn.(*Func)
        fp = fn.proto
    )
    traceFn(ls, fn)
    for {
        switch inst := traceVM(fp, ci); inst.Code() {
            // Unary operators.
            //
            // @args A B
            //
            // R(A) := OP RK(B)
            case code.BNOT, code.UNM, code.NOT, code.LEN:
                var (
                    op = Op(inst.Code()-code.UNM)+OpMinus
                    x  = ls.rk(fn, inst.B())
                    v Value
                )
                if v, ci.err = unary(ls, op, x); ci.err != nil {
                    return
                }
                ls.stack[ci.base+inst.A()] = v
            
            // Comparison operators with conditional jump.
            //
            // @args A B C
            //
            // if ((RK(B) OP RK(C)) ~= A) then pc++
            case code.EQ, code.LT, code.LE:
                var (
                    op = Op(inst.Code()-code.EQ)+OpEq
                    x  = ls.rk(fn, inst.B())
                    y  = ls.rk(fn, inst.C())
                    v bool
                )
                if v, ci.err = compare(ls, op, x, y); ci.err != nil {
                    return
                }
                if v != (inst.A() == 1) {
                    ci.pc++
                }

            // Binary operators.
            //
            // @args A B C
            //
            // R(A) := RK(B) OP RK(C) 
            case code.ADD,
                code.SUB,
                code.MUL,
                code.MOD,
                code.POW,
                code.DIV,
                code.IDIV,
                code.BAND,
                code.BOR,
                code.BXOR,
                code.SHL,
                code.SHR:
                var (
                    op = Op(inst.Code()-code.ADD)+OpAdd
                    x  = ls.rk(fn, inst.B())
                    y  = ls.rk(fn, inst.C())
                    v Value
				)
                if v, ci.err = binary(ls, op, x, y); ci.err != nil {
                    return
                }
                ls.stack[ci.base+inst.A()] = v

            // // CONCAT: Concatenate a range of registers.
            // //
            // // @args A B C
            // //
            // // R(A) := R(B).. ... ..R(C)
            // case code.CONCAT:
            //     // fmt.Printf("\t[%d] %v\n", ci.pc, inst)
            //     var (
            //         xs = fn.stack[ci.base+inst.B():ci.base+inst.C()+1]
            //         v Value
            //     )
            //     if v, err = concat(ls, xs); err != nil {
            //         break frame 
            //     }
            //     fn.stack[ci.base+inst.A()] = v

            // FORPREP: Initialization for a numeric for loop.  
            //
            // @args A sBx 
            //
            // R(A) -= R(A+2); pc+=sBx
            // case code.FORPREP:
            //     fmt.Printf("\t[%d] %v\n", ci.pc, inst)
            //     var (
            //         index = fn.stack[ci.base+inst.A()+0]
            //         limit = fn.stack[ci.base+inst.A()+1]
            //         step  = fn.stack[ci.base+inst.A()+2]
            //     )
            //     var (
            //         i1, ok1 = index.(Int)
            //         i2, ok2 = limit.(Int)
            //         i3, ok3 = step.(Int)
            //     )
            //     // TODO: Try converting forlimit to an integer rounding if possible.
            //     if ok1 && ok2 && ok3 {
            //         fn.stack[ci.base+inst.A()+0] = i1 - i3
            //         fn.stack[ci.base+inst.A()+1] = i2
            //         fn.stack[ci.base+inst.A()+2] = i3
            //         ci.pc += inst.SBX()
            //     } else {
            //         var ( f1, f2, f3 Float )

            //         if f1, ok1 = AsFloat(index); !ok1 {
            //             err = fmt.Errorf("'for' init must be a number")
            //             break frame
            //         }
            //         if f1, ok2 = AsFloat(limit); !ok2 {
            //             err = fmt.Errorf("'for' limit must be a number")
            //             break frame
            //         }
            //         if f3, ok3 = AsFloat(step); !ok3 {
            //             err = fmt.Errorf("'for' step must be a number")
            //             break frame
            //         }
            //         fn.stack[ci.base+inst.A()+0] = f1 - f3
            //         fn.stack[ci.base+inst.A()+1] = f2
            //         fn.stack[ci.base+inst.A()+2] = f3
            //         ci.pc += inst.SBX()
            //     }

            // // FORLOOP: Iterate a numeric for loop.
            // //
            // // @args A sBx 
            // //
            // // R(A) += R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }
            // case code.FORLOOP:
            //     // fmt.Printf("\t[%d] %v\n", ci.pc, inst)
            //     var (
            //         index = fn.stack[ci.base+inst.A()+0]
            //         limit = fn.stack[ci.base+inst.A()+1]
            //         step  = fn.stack[ci.base+inst.A()+2]
            //     )
            //     if i1, ok := index.(Int); ok {
            //         var (
            //             i2 = limit.(Int)
            //             i3 = step.(Int)
            //         )
            //         if i1 += i3; (i3 >= 0 && (i1 <= i2)) || (i3 <= 0 && (i1 >= i2)) {
            //             fn.stack[ci.base+inst.A()] = i1
            //             fn.stack[ci.base+inst.A()+3] = i1
            //             ci.pc += inst.SBX()
            //         }
            //     } else {
            //         var (
            //             f1 = index.(Float)
            //             f2 = limit.(Float)
            //             f3 = step.(Float)
            //         )
            //         if f1 += f3; (f3 >= 0 && (f1 <= f2)) || (f3 <= 0 && (f1 >= f2)) {
            //             fn.stack[ci.base+inst.A()] = f1
            //             fn.stack[ci.base+inst.A()+3] = f1
            //             ci.pc += inst.SBX()
            //         }
            //     }

            // TFORCALL: Iterate a generic for loop.
            //
            // R(A) is the iterator function, R(A+1) is the state, R(A+2) is the
            // control variable. At the start, R(A+2) has an initial value.
            //
            // Loop variables reside at locations R(A+3) and up, and their count
            // is specified in operand C. Operand C must be at least 1.
            //
            // Each time tforcall executes, the iterator function referenced by
            // R(A) is called with two arguments, the state R(A+1) and control
            // variable R(A+2). The results are returned in the local loop
            // variables, from R(A+3) up to R(A+2+C).
            //
            // @args A C
            //
            // R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2))
            // case code.TFORCALL:
            //     // tforcall expects the for variables below to be at a fixed
            //     // position in the stack for every iteration, so we need to
            //     // adjust the stack to ensure this to avoid side effects.
            //     var (
            //         ctrl = ls.stack[ci.base+inst.A()+2]
            //         data = ls.stack[ci.base+inst.A()+1]
            //         iter = ls.stack[ci.base+inst.A()]
            //         base = ci.base+inst.A() + 3
            //     )
            //     rvs := ls.call(iter, []Value{data, ctrl}, inst.C())
            //     for i, ret := range rvs {
            //         fn.stack[base+i] = ret
            //     }

            // TFORLOOP: Initialization for a generic for loop.
            //
            // @args A sBx 
            //
            // if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx } 
            case code.TFORLOOP:
                if ctrl := ls.stack[ci.base+inst.A()+1]; ctrl != nil { // continue loop?
                    ls.stack[ci.base+inst.A()] = ctrl // save control variable
                    ci.pc += inst.SBX() // jump back
                }

            // NEWTABLE: Create a new table.
            //
            // Creates a new empty table at register R(A). B and C are the encoded size information
            // for the array part and the hash part of the table, respectively. Appropriate values
            // for B and C are set in order to avoid rehashing when initially populating the table
            // with array values or hash key-value pairs.
            //
            // Operand B and C are both encoded as a "floating point byte" (see lobject.c)
            // which is eeeeexxx in binary, where x is the mantissa and e is the exponent.
            // The actual value is calculated as 1xxx*2^(eeeee-1) if eeeee is greater than
            // 0 (a range of 8 to 15*2^30).
            //
            // If eeeee is 0, the actual value is xxx (a range of 0 to 7.)
            //
            // If an empty table is created, both sizes are zero. If a table is created with a number
            // of objects, the code generator counts the number of array elements and the number of
            // hash elements. Then, each size value is rounded up and encoded in B and C using the
            // floating point byte format.
            //
            // @args A B C
            //
            // R(A) := {} (size = B,C)
            case code.NEWTABLE:
                var (
                    arrN = fb2int(inst.B())
                    kvsN = fb2int(inst.C())
                )
                ls.stack[ci.base+inst.A()] = NewTableSize(arrN, kvsN)

            // GETTABLE: Read a table element into a register (locals).
            //
            // @args A B C
            //
            // R(A) := R(B)[RK(C)]
            case code.GETTABLE:
                var (
                    t = ls.stack[ci.base+inst.B()]
                    k = ls.rk(fn, inst.C())
                    v Value
                )
                if v, ci.err = gettable(ls, t, *k); ci.err != nil {
                    return
                }
                ls.stack[ci.base+inst.A()] = v

            // SETTABLE: Write a register value into a table element (locals).
            //
            // @args A B C
            //
            // R(A)[RK(B)] := RK(C)
            case code.SETTABLE:
                var (
                    t = ls.stack[ci.base+inst.A()]
                    k = ls.rk(fn, inst.B())
                    v = ls.rk(fn, inst.C())
                )
                if ci.err = settable(ls, t, *k, *v); ci.err != nil {
                    return
                }

            // SETLIST: Set a range of array elements for a table.
            //
            // Sets the values for a range of array elements in a table referenced by R(A). Field B is the number
            // of elements to set. Field C encodes the block number of the table to be initialized.
            //
            // The values used to initialize the table are located in registers R(A+1), R(A+2), and so on.
            //
            // The block size is denoted by FPF. FPF is ‘fields per flush’, defined as LFIELDS_PER_FLUSH in the source
            // file lopcodes.h, with a value of 50. For example, for array locations 1 to 20, C will be 1 and B will
            // be 20.
            // 
            // If B is 0, the table is set with a variable number of array elements, from register R(A+1) up to the top
            // of the stack. This happens when the last element in the table constructor is a function call or a vararg
            // operator.
            // 
            // If C is 0, the next instruction is cast as an integer, and used as the C value. This happens only when
            // operand C is unable to encode the block number, i.e. when C > 511, equivalent to an array index greater
            // than 25550.
            //
            // @args A B C
            //
            // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
            // case code.SETLIST:
            //     var (
            //         a = inst.A()
            //         b = inst.B()
            //         c = inst.C()
            //     )
            //     if b == 0 {
            //         b = (ci.sp - a) - 1
            //     }
            //     if c == 0 {
            //         // ASSERT: fn.Instrs[fn.pc+1] == EXTRAARG)
            //         c = fp.Instrs[ci.pc+1].AX()
            //         ci.pc++
            //     }
            //     o := (c - 1) * fieldsPerFlush + b
            //     t := fn.stack[ci.base+a].(*Table)
            //     for b > 0 {
            //         t.Set(Int(o), fn.stack[ci.base+a+b])
            //         o--
            //         b--
            //     }

            // SELF: Prepare an object method for calling.  
            //
            // @args A B C
            //
            // R(A+1) := R(B); R(A) := R(B)[RK(C)] 
            case code.SELF:
                var (
                    self = ls.stack[ci.base+inst.B()]
                    k    = ls.rk(fn, inst.C())
                    v Value   
                )
                if v, ci.err = gettable(ls, self, *k); ci.err != nil {
                    return
                }
                ls.stack[ci.base+inst.A()+1] = self
                ls.stack[ci.base+inst.A()] = v

            // GETTABUP: Read a value from table in
            // up-value into a register (globals).
            //
            // @args A B C
            //   
            // R(A) := UpValue[B][RK(C)]
            case code.GETTABUP:
                var (
                    k = ls.rk(fn, inst.C())
                    t = fn.up[inst.B()].get()
                    v Value
                )
                if v, ci.err = gettable(ls, t, *k); ci.err != nil {
                    return
                }
                ls.stack[ci.base+inst.A()] = v

            // SETTABUP: Write a register value into table in up-value (globals).
            //
            // @args A B C
            //
            // UpValue[A][RK(B)] := RK(C)
            case code.SETTABUP:
                var (
                    k = ls.rk(fn, inst.B())
                    v = ls.rk(fn, inst.C())
                    t = fn.up[inst.A()].get()
                )
                if ci.err = settable(ls, t, *k, *v); ci.err != nil {
                    return
                }

            // GETUPVAL: Read an upvalue into a register.
            //
            // @args A B
            //       
            // R(A) := UpValue[B]
            case code.GETUPVAL:
                ls.stack[ci.base+inst.A()] = fn.up[inst.B()].get()

            // SETUPVAL: Write a register value into an upvalue.
            //
            // @args A B
            //
            // UpValue[B] := R(A)
            case code.SETUPVAL:
                fn.up[inst.B()].set(ls.stack[ci.base+inst.A()])
            
            // TESTSET: Boolean test, with conditional jump and assignment.
            //
            // @args A B C
            //
            // if (R(B) <=> C) then R(A) := R(B) else pc++
            case code.TESTSET:
                if Truth(ls.stack[ci.base+inst.B()]) != (inst.C() == 1) {
                    ci.pc++
                }
                ls.stack[ci.base+inst.A()] = ls.stack[ci.base+inst.B()]

            // TEST: Boolean test, with conditional jump.    
            //
            // @args A C 
            //
            // if not (R(A) <=> C) then pc++  
            case code.TEST:
                if Truth(ls.stack[ci.base+inst.A()]) != (inst.C() == 1) {
                    ci.pc++
                }

            // LOADNIL: Load nil values into a range of registers.
            // 
            // @args A B
            //
            // R(A), R(A+1), ..., R(A+B) := nil
            case code.LOADNIL:
                for i := inst.A(); i <= inst.A() + inst.B(); i++ {
                    ls.stack[ci.base+i] = nil
                }

            // LOADBOOL: Load a boolean into a register.
            //
            // @args A B C
            //
            // R(A) := (Bool)B; if (C) pc++
            case code.LOADBOOL:
                truth := (Bool(inst.B() == 1))
                ls.stack[ci.base+inst.A()] = truth
                if inst.C() != 0 {
                    ci.pc++
                }

            // LOADKX: Load a constant into a register.
            // The next 'instruction' is always EXTRAARG.
            //
            // @args A
            //
            // R(A) := Kst(extra arg)
            case code.LOADKX:
                ls.stack[ci.base+inst.A()] = fn.kst(fp.Instrs[ci.pc+1].AX())
                ci.pc++

            // LOADK: Load a constant into a register.
            //
            // @args A Bx
            //
            // R(A) := Kst(Bx)
            case code.LOADK:
                ls.stack[ci.base+inst.A()] = fn.kst(inst.BX())

            // MOVE: Copy a value between registers.
            //
            // @args A B
            //
            // R(A) := R(B)
            case code.MOVE:
                ls.stack[ci.base+inst.A()] = ls.stack[ci.base+inst.B()]

            // JMP: Unconditional jump.
            //
            // @args A sBx
            //
            // pc+=sBx; if (A) close all upvalues >= R(A-1)
            case code.JMP:
                // if (A) close all upvalues >= R(A-1)
                if inst.A() != 0 {
                    fn.close(ci.base+inst.A()-1)
                }
                ci.pc += inst.SBX()

            // CLOSURE: Create a closure of a function prototype.
            //
            // @args A Bx
            // 
            // R(A) := closure(KPROTO[Bx])
            case code.CLOSURE:
                cls := &Func{proto: fp.Protos[inst.BX()]}
                ls.stack[ci.base+inst.A()] = cls
                cls.open(ls, fn.up...)

            // VARARG: Assign vararg function arguments to registers.
            //
            // VARARG copies B-1 parameters into a number of registers starting from R(A),
            // padding with nils if there aren’t enough values. If B is 0, VARARG copies
            // as many values as it can based on the number of parameters passed.
            //
            // If a fixed number of values is required, B is a value greater than 1.
            // If any number of values is required, B is 0.
            //
            // If B == 0, load all varargs.
            // If B >= 1, load B-1 varargs.
            //
            // @args A B
            //
            // R(A), R(A+1), ..., R(A+B-2) = vararg
            case code.VARARG:
                var (
                    n = (ci.base - ci.fnID) - fp.ParamN - 1
                    a = ci.base + inst.A()
                    b = inst.B() - 1
                    j int
                )
                if n < 0 {
                   n = 0
                }
                if b < 0 {
                    b = n
                    ls.check(n)
                    ls.top = a + n
                }
                for j < b && j < n {
                    ls.stack[a+j] = ls.stack[ci.base-n+j]
                    j++
                }
                for j < b {
                    ls.stack[a+j] = nil
                    j++
                }

            // TAILCALL: Perform a tail call.
            //
            // TAILCALL performs a tail call, which happens when a return statement has a single
            // function call as the expression, e.g. return foo(bar). A tail call results in the
            // function being interpreted within the same call frame as the caller -- the stack
            // is replaced and then a 'goto' executed to start at the entry point in the VM. Only
            // Lua functions can be tailcalled. Tail calls allow infinite recursion without growing
            // the stack.
            //
            // Like OP_CALL, registry R(A) holds the reference to the function object to be called.
            // B encodes the number of parameters in the same way as in OP_CALL.
            //
            // C isn't used by TAILCALL, since all return results are used. In any case, Lua always
            // generates a 0 for C denoting multiple return results. 
            //
            // @args A B C
            //
            // return R(A)(R(A+1), ... ,R(A+B-1))
            case code.TAILCALL:
                var (
                    fnAt = ci.base + inst.A()
                    argc = inst.B()
                )
                if argc != 0 {
                    ls.top = fnAt + argc
                }
                if ls.calls = ls.cont(fnAt, -1); !ls.calls.isLua() {
                    ls.calls.fn.call(ls)
                    continue
                }
                // put called frame (n) in place of caller (o)
                var (
                    newci = ls.calls
                    oldci = ls.calls.prev                    
                    newfn = ls.stack[newci.fnID].(*Func)
                )

                // last stack slot filled by 'precall'
                last := newci.base + newfn.proto.ParamN

                // close all upvalues from previous call
                if len(fp.Protos) > 0 {
                    fn.close(oldci.base)
                }
                // move new frame into old one
                for idx := 0; newci.fnID + idx < last; idx++ {
                    ls.stack[oldci.fnID + idx] = ls.stack[newci.fnID + idx]
                }
                oldci.base = oldci.fnID + (newci.base - newci.fnID) // correct base
                oldci.top = oldci.fnID + (ls.top - newci.fnID) // correct top
                ls.top = oldci.top

                oldci.flag |= tailcall
                oldci.pc = newci.pc
                oldci.fn = newfn

                ci, ls.calls = oldci, oldci
                goto frame

            // CALL: Calls a function.
            //
            // CALL performs a function call, with register R(A) holding the reference to
            // the function object to be called. Parameters to the function are placed in
            // the registers following R(A).
            //
            // If B is 1, the function has no parameters.
            //
            // If B is 2 or more, there are (B-1) parameters, and upon entry entry to the
            // called function, R(A+1) will become the base.
            //
            // If B is 0, then B = 'top', i.e., the function parameters range from R(A+1) to
            // the top of the stack. This form is used when the number of parameters to pass
            // is set by the previous VM instruction, which has to be one of OP_CALL or OP_VARARG.
            //
            // If C is 1, no return results are saved. If C is 2 or more, (C-1) return values are
            // saved.
            //
            // If C == 0, then 'top' is set to last_result+1, so that the next open instruction
            // (i.e. OP_CALL, OP_RETURN, OP_SETLIST) can use 'top'.
            //
            // If C > 1, results returned by the function call are placed in registers ranging
            // from R(A) to R(A+C-1).
            //
            // @args A B C
            //
            // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
            case code.CALL:
                var (
                    fnAt = ci.base + inst.A()
                    retc = inst.C() - 1
                    argc = inst.B()
                )
                if argc != 0 {
                    ls.top = fnAt + argc
                }
                if ls.calls = ls.cont(fnAt, retc); !ls.calls.isLua() {
                    if ls.calls.fn.call(ls); retc >= 0 {
                        ls.top = ci.top
                    }
                } else {
                    ci = ls.calls
                    goto frame
                }

            // RETURN: Returns from function call.
            //
            // Returns to the calling function, with optional return values.
            // First, op.RETURN closes any open upvalues by calling fn.close().
            //
            // If B == 0, the set of values ranges from R(A) to the top of the stack.
            // If B == 1, there are no return values.
            // If B >= 2, there are (B-1) return values, located in consecutive
            // register from R(A) ... R(A+B-1).
            //
            // If B == 0, then the previous instruction (which must be either op.CALL or
            // op.VARARG) would have set state top to indicate how many values to return.
            // The number of values to be returned in this case is R(A) to ci.sp.
            //
            // If B > 0, then the number of values to be returned is simply B-1.
            // 
            // If (B == 0) then return up to 'top'.
            //
            // @args A B
            // 
            // return R(A), ... ,R(A+B-2)
            case code.RETURN:
                if len(fp.Protos) > 0 {
                    fn.close(ci.base)
                }
                var (
                    a = ci.base + inst.A()
                    b = inst.B()
                )
                if b != 0 {
                    b--
                } else {
                    b = ls.top - a
                }
                rx := ls.returns(ci, a, b)
                if ci.flag & fresh != 0 {
                    return
                }
                if ci = ls.calls; rx {
                    ls.top = ci.top
                }
                goto frame

            // EXTRAARG: Extra (larger) argument for previous opcode.
            //
            // @args Ax
            case code.EXTRAARG:
                // This op func should never execute directly.
                panic("unreachable")

            default:
                panic(fmt.Errorf("unhandled instruction: %v", inst)) // Fatal
        }
    }
}

const debugVM = false

func traceVM(fp *code.Proto, ci *call) code.Instr {
    var (
        inst = fp.Instrs[ci.pc]
        line = int32(-1)
    )
    if fp.PcLine != nil {
        line = fp.PcLine[ci.pc]
    }
    if debugVM {
        fmt.Printf("\t[%d] %s:%d: %s\n", ci.pc, fp.Source, line, inst)
    }
    ci.pc++
    return inst
}

func traceFn(ls *thread, fn *Func) *thread {
    if ci := ls.calls; debugVM {
        fmt.Printf("[LUA]: %s\n", ci)
        fmt.Printf("\t@stack = %v (top = %d)\n", ls.stack[ci.base:ci.top], ls.top)
        fmt.Printf("\t@args  = %v\n", ls.stack[ci.base:ci.base+fn.proto.ParamN])
        fmt.Printf("\t@varg  = %v\n\t--\n", ls.stack[ci.fnID+1:ci.fnID+(ci.base-ci.fnID)])
    }
    return ls
}

func (ls *thread) returns(ci *call, first, retsN int) bool {
    // TODO: handle return/line hooks
    ls.calls = ci.prev
    var (
        dst = ci.fnID
        ret = ci.retc
    )
    // fmt.Printf("\t\t@ return %d results to %d (wanted = %d)\n", retsN, dst, ret)
    switch ret {
        case multRet:
            for i := 0; i < retsN; i++ {
                ls.stack[dst+i] = ls.stack[first+i]
            }
            ls.top = dst + retsN
            return true
		case 0:
            break
        case 1:
            if retsN == 0 {
                ls.stack[dst] = nil
            } else {
                // fmt.Printf("\t\t@ <= %v\n", ls.stack[first])
                ls.stack[dst] = ls.stack[first]
            }
        default:
            if ret <= retsN {
                for i := 0; i < ret; i++ {
                    ls.stack[dst+i] = ls.stack[first+i]
                }
            } else {
                var i int
                for i < retsN {
                    ls.stack[dst+i] = ls.stack[first+i]
                    i++
                }
                for i < ret {
                    ls.stack[dst+i] = nil
                    i++
                }
            }
    }
    ls.top = dst + ret
    return false
}