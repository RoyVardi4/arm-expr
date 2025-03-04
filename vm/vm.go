package vm

//go:generate sh -c "go run ./func_types > ./func_types[generated].go"

import (
	"fmt"
	"reflect"
	"time"

	"expr/file"
	"expr/internal/deref"
	"expr/vm/runtime"
)

func Run(program *Program, env any) (any, error) {
	if program == nil {
		return nil, fmt.Errorf("program is nil")
	}

	vm := VM{}
	return vm.Run(program, env)
}

func Debug() *VM {
	vm := &VM{
		debug: true,
		step:  make(chan struct{}),
		curr:  make(chan int),
	}
	return vm
}

type VM struct {
	Stack        []any
	Scopes       []*Scope
	ip           int
	memory       uint
	memoryBudget uint
	debug        bool
	step         chan struct{}
	curr         chan int
}

func (vm *VM) Run(program *Program, env any) (_ any, err error) {
	defer func() {
		if r := recover(); r != nil {
			var location file.Location
			if vm.ip-1 < len(program.locations) {
				location = program.locations[vm.ip-1]
			}
			f := &file.Error{
				Location: location,
				Message:  fmt.Sprintf("%v", r),
			}
			if err, ok := r.(error); ok {
				f.Wrap(err)
			}
			err = f.Bind(program.source)
		}
	}()

	if vm.Stack == nil {
		vm.Stack = make([]any, 0, 2)
	} else {
		vm.Stack = vm.Stack[0:0]
	}
	if vm.Scopes != nil {
		vm.Scopes = vm.Scopes[0:0]
	}

	vm.memoryBudget = MemoryBudget
	vm.memory = 0
	vm.ip = 0

	for vm.ip < len(program.Bytecode) {
		if debug && vm.debug {
			<-vm.step
		}

		op := program.Bytecode[vm.ip]
		arg := program.Arguments[vm.ip]
		vm.ip += 1

		switch op {

		case OpInvalid:
			panic("invalid opcode")

		case OpPush:
			vm.push(program.Constants[arg])

		case OpLoadConst:
			vm.push(runtime.Fetch(env, program.Constants[arg]))

		case OpLoadField:
			vm.push(runtime.FetchField(env, program.Constants[arg].(*runtime.Field)))

		case OpLoadFast:
			vm.push(env.(map[string]any)[program.Constants[arg].(string)])

		case OpLoadMethod:
			vm.push(runtime.FetchMethod(env, program.Constants[arg].(*runtime.Method)))

		case OpLoadFunc:
			vm.push(program.functions[arg])

		case OpFetch:
			b := vm.pop()
			a := vm.pop()
			vm.push(runtime.Fetch(a, b))

		case OpFetchField:
			a := vm.pop()
			vm.push(runtime.FetchField(a, program.Constants[arg].(*runtime.Field)))

		case OpMethod:
			a := vm.pop()
			vm.push(runtime.FetchMethod(a, program.Constants[arg].(*runtime.Method)))

		case OpTrue:
			vm.push(true)

		case OpFalse:
			vm.push(false)

		case OpNil:
			vm.push(nil)

		case OpCall:
			fn := reflect.ValueOf(vm.pop())
			size := arg
			in := make([]reflect.Value, size)
			for i := int(size) - 1; i >= 0; i-- {
				param := vm.pop()
				if param == nil && reflect.TypeOf(param) == nil {
					// In case of nil value and nil type use this hack,
					// otherwise reflect.Call will panic on zero value.
					in[i] = reflect.ValueOf(&param).Elem()
				} else {
					in[i] = reflect.ValueOf(param)
				}
			}
			out := fn.Call(in)
			if len(out) == 2 && out[1].Type() == errorType && !out[1].IsNil() {
				panic(out[1].Interface().(error))
			}
			vm.push(out[0].Interface())

		case OpCall0:
			out, err := program.functions[arg]()
			if err != nil {
				panic(err)
			}
			vm.push(out)

		case OpCall1:
			a := vm.pop()
			out, err := program.functions[arg](a)
			if err != nil {
				panic(err)
			}
			vm.push(out)

		case OpCall2:
			b := vm.pop()
			a := vm.pop()
			out, err := program.functions[arg](a, b)
			if err != nil {
				panic(err)
			}
			vm.push(out)

		case OpCall3:
			c := vm.pop()
			b := vm.pop()
			a := vm.pop()
			out, err := program.functions[arg](a, b, c)
			if err != nil {
				panic(err)
			}
			vm.push(out)

		case OpCallN:
			fn := vm.pop().(Function)
			size := arg
			in := make([]any, size)
			for i := int(size) - 1; i >= 0; i-- {
				in[i] = vm.pop()
			}
			out, err := fn(in...)
			if err != nil {
				panic(err)
			}
			vm.push(out)

		case OpCallFast:
			fn := vm.pop().(func(...any) any)
			size := arg
			in := make([]any, size)
			for i := int(size) - 1; i >= 0; i-- {
				in[i] = vm.pop()
			}
			vm.push(fn(in...))

		case OpCallTyped:
			vm.push(vm.call(vm.pop(), arg))

		case OpCast:
			switch arg {
			case 0:
				vm.push(runtime.ToInt(vm.pop()))
			case 1:
				vm.push(runtime.ToInt64(vm.pop()))
			case 2:
				vm.push(runtime.ToFloat64(vm.pop()))
			}

		case OpDeref:
			a := vm.pop()
			vm.push(deref.Deref(a))

		case OpProfileStart:
			span := program.Constants[arg].(*Span)
			span.start = time.Now()

		case OpProfileEnd:
			span := program.Constants[arg].(*Span)
			span.Duration += time.Since(span.start).Nanoseconds()

		case OpBegin:
			a := vm.pop()
			array := reflect.ValueOf(a)
			vm.Scopes = append(vm.Scopes, &Scope{
				Array: array,
				Len:   array.Len(),
			})

		case OpEnd:
			vm.Scopes = vm.Scopes[:len(vm.Scopes)-1]

		default:
			panic(fmt.Sprintf("unknown bytecode %#x", op))
		}

		if debug && vm.debug {
			vm.curr <- vm.ip
		}
	}

	if debug && vm.debug {
		close(vm.curr)
		close(vm.step)
	}

	if len(vm.Stack) > 0 {
		return vm.pop(), nil
	}

	return nil, nil
}

func (vm *VM) push(value any) {
	vm.Stack = append(vm.Stack, value)
}

func (vm *VM) current() any {
	return vm.Stack[len(vm.Stack)-1]
}

func (vm *VM) pop() any {
	value := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return value
}

func (vm *VM) memGrow(size uint) {
	vm.memory += size
	if vm.memory >= vm.memoryBudget {
		panic("memory budget exceeded")
	}
}

func (vm *VM) scope() *Scope {
	return vm.Scopes[len(vm.Scopes)-1]
}

func (vm *VM) Step() {
	vm.step <- struct{}{}
}

func (vm *VM) Position() chan int {
	return vm.curr
}
