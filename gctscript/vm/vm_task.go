package vm

import (
	"time"
)

func (vm *VM) runner() {
	vm.S = make(chan struct{}, 1)
	waitTime := time.NewTicker(vm.T)
	vm.NextRun = time.Now().Add(vm.T)

	go func() {
		for {
			select {
			case <-waitTime.C:
				vm.NextRun = time.Now().Add(vm.T)
				err := vm.RunCtx()
				if err != nil {
					return
				}
			case <-vm.S:
				waitTime.Stop()
				return
			}
		}
	}()
}
