package cryptore

import (
	"github.com/ngchain/go-randomx"
	"sync"
)

type RandxVm struct {
	sync.Mutex
	cache   randomx.Cache
	dataset randomx.Dataset
	vm      randomx.VM
}

func NewRandxVm(key []byte) (ret *RandxVm, err error) {
	cache, err := randomx.AllocCache(randomx.FlagDefault)
	if nil != err {
		return
	}
	randomx.InitCache(cache, key)

	dataset, err := randomx.AllocDataset(randomx.FlagDefault)
	if nil != err {
		return
	}
	randomx.InitDataset(dataset, cache, 0, randomx.DatasetItemCount()) // todo: multi core acceleration

	vm, err := randomx.CreateVM(cache, dataset, randomx.FlagJIT, randomx.FlagHardAES, randomx.FlagFullMEM)
	if nil != err {
		return
	}

	ret = &RandxVm{
		cache:   cache,
		dataset: dataset,
		vm:      vm,
	}

	return
}

func (this *RandxVm) Close() {
	randomx.DestroyVM(this.vm)
	randomx.ReleaseDataset(this.dataset)
	randomx.ReleaseCache(this.cache)
}

func (this *RandxVm) Hash(buf []byte) (ret []byte) {
	return randomx.CalculateHash(this.vm, buf)
}

func randomxhash(vm *RandxVm, buf []byte) (ret []byte) {
	ret = vm.Hash(buf)
	return
}
