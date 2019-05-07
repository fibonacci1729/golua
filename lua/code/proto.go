package code

type (
	Const interface {}

	Proto struct {
		UpVars []*UpVar // information about the function's upvalues
		Protos []*Proto // functions defined inside the function
		Consts []Const  // constants used by the function
		Instrs []Instr  // function instructions
		Vararg bool     // true variable number of arguments
		ParamN int      // number of fixed parameters
		StackN int      // number of registers need by this function

		// debug information
		Locals []*Local // local variable information
		PcLine []int32  // pc -> line
		Source string   // source name
		SrcPos int      // line defined
		EndPos int      // last line defined
	}

	UpVar struct {
		Name  string
		Stack bool
		Index int
	}

	Local struct {
		Name string
		Live int32
		Dead int32
	}
)

// FindSetRegister tries to find the last instruction before
// 'lastPC' that modified register 'register'.
//
// Returns the index of the instruction.
func FindSetRegister(proto *Proto, lastPC, reg int) int {
	var (
		set = -1 // keep last instruction that changed 'reg'
		jmp = 0  // any code before this address is conditional
	)
	filterPC := func(pc, jmp int) int {
		if pc < jmp { // is code conditional (inside a jump)?
			return -1 // cannot know who sets that register
		}
		return pc // current position sets that register
	}
	for pc := 0; pc < lastPC; pc++ {
		switch inst := proto.Instrs[pc]; inst.Code() {
			case CALL, TAILCALL:
				if reg >= inst.A() { // affect all registers above base
					set = filterPC(pc, jmp)
				}
			case TFORCALL:
				if reg >= inst.A() + 2 { // affect all registers above its base
					set = filterPC(pc, jmp)
				}
			case LOADNIL:
				if a, b := inst.A(), inst.B(); reg >= a && reg <= a + b {
					// set register from 'a' to 'a+b'
					set = filterPC(pc, jmp)
				}
			case JMP:
				// jump is forward and do not skip 'lastpc'?
				if dst := pc + 1 + inst.SBX(); dst > pc && dst <= lastPC {
					if dst > jmp {
						jmp = dst // update jump target
					}
				}
			default:
				if inst.Code().Mask().SetA() && (reg == inst.A()) {
					// any instruction that set A
					set = filterPC(pc, jmp)
				}
		}
	}
	// 	return set
	return set
}

// ObjectName tries to find a name for the object at register reg.
//
// Returns the name and a string "what" describing context.
func ObjectName(proto *Proto, lastpc, reg int) (name, what string) {
	if name = LocalName(proto, lastpc, reg+1); name != "" {
		return name, "local"
	}
	// try symbolic execution
	if pc := FindSetRegister(proto, lastpc, reg); pc != -1 { // instruction found?
		switch inst := proto.Instrs[pc]; inst.Code() {
			case GETTABUP:
				var (
					t = inst.B() // table index
					k = inst.C() // key index
				)
				if what = "?"; t < len(proto.UpVars) {
					if up := proto.UpVars[t]; up.Name != "" {
						if up.Name == "_ENV" {
							what = "global"
						} else {
							what = "field"
						}
					}
				}
				return FindNameRK(proto, pc, k), what
			case GETTABLE:
				var (
					t = inst.B() // table index
					k = inst.C() // key index
				)
				switch LocalName(proto, pc, t+1) {
					case "_ENV":
						what = "global"
					default:
						what = "field"
				}
				return FindNameRK(proto, pc, k), what
			case GETUPVAL:
				if name = "?"; inst.B() < len(proto.UpVars) {
					if up := proto.UpVars[inst.B()]; up.Name != "" {
						name = up.Name
					}
				}
				return name, "upvalue"
			case LOADKX:
				kst := proto.Consts[proto.Instrs[pc+1].AX()]
				if s, ok := kst.(string); ok {
					return s, "constant"
				}
			case LOADK:
				kst := proto.Consts[inst.BX()]
				if s, ok := kst.(string); ok {
					return s, "constant"
				}
			case SELF:
				return FindNameRK(proto, pc, inst.C()), "method"
			case MOVE:
				if inst.B() < inst.A() { // move from 'b' to 'a'
					return ObjectName(proto, pc, inst.B()) // get name for 'b'
				}
		}
	}
	return "", ""
}

// Find a "name" for the RK value at 'kst'.
func FindNameRK(proto *Proto, pc, rk int) string {
	if IsKst(rk) { // is 'rk' a constant?
		if s, ok := proto.Consts[ToKst(rk)].(string); ok { // literal constant?
			return s // it is its own name
		}
		// else no reasonable name found
		return ""
	}
	// else 'rk' is a register
	name, what := ObjectName(proto, pc, rk)
	if what == "constant" {
		return name
	}
	return "?"
}

// LocalName searches for the n-th local variable in function.
//
// Returns the empty string if not found.
func LocalName(proto *Proto, pc, local int) string {
	for i := 0; i < len(proto.Locals) && proto.Locals[i].Live <= int32(pc); i++ {
		if int32(pc) < proto.Locals[i].Dead { // is variable active?
			if local--; local == 0 {
				return proto.Locals[i].Name
			}
		}
	}
	// not found
	return ""
}